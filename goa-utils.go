package goautils

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/codeclysm/ctxlog/v2"
	httpmdlwr "goa.design/goa/v3/http/middleware"
	goamdlwr "goa.design/goa/v3/middleware"
)

type logger interface {
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Error(msg string, fields ...map[string]interface{})
}

// ErrorHandler returns a function that writes and logs the given error.
func ErrorHandler(log logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		log.Error(err.Error())
	}
}

// RequestID is a wrapper around the goa middleware with the same name,
// except it also augments the ctxlog with the request id
func RequestID() func(http.Handler) http.Handler {
	goaReqID := httpmdlwr.RequestID(
		httpmdlwr.UseXRequestIDHeaderOption(true),
		httpmdlwr.XRequestHeaderLimitOption(128),
	)

	return func(h http.Handler) http.Handler {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			reqID := ctx.Value(goamdlwr.RequestIDKey)

			ctx = ctxlog.WithFields(ctx, map[string]interface{}{"reqID": reqID})

			h.ServeHTTP(w, r.WithContext(ctx))
		})

		return goaReqID(handler)
	}
}

// ListenGracefully will start a listener on a specified address and blocks
// until it receives an interrupt or a sigterm, at which point it will start
// a graceful shutdown and finally return.
// You MUST provide a logger, it's the only way to see if there are errors
func ListenGracefully(addr string, handler http.Handler, log logger) {
	srv := &http.Server{Addr: addr, Handler: handler}

	idleConnsClosed := make(chan struct{})

	log.Info("Listen and gracefully shutdown", map[string]interface{}{
		"addr": addr,
	})

	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)
		// sigterm signal sent from kubernetes
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		log.Info("Start graceful shutdown. Waiting for current request to finish")
		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Error("HTTP server Shutdown", map[string]interface{}{
				"err": err,
			})
		}

		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Error("HTTP server ListenAndServe", map[string]interface{}{
			"err": err,
		})
	}

	<-idleConnsClosed

	log.Info("Shutdown")
}
