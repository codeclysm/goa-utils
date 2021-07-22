goa utils
=========

This is just a collection of something I find useful while writing applications using https://goa.design/

# Logging
I try not to use external dependencies when writing libraries, so for logging I stand by the approach of https://github.com/logur/logur.

You can just pass a logger that implements this simple interface (which is compatible with the logur one):

```
type logger interface {
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Error(msg string, fields ...map[string]interface{})
}
```

For example if you want to use logrus:

```
import (
	"github.com/sirupsen/logrus"
	logrusadapter "logur.dev/adapter/logrus"
)

	logrusLog := logrus.New()
	log := logrusadapter.New(logrusLog)
```