package goautils

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
