// Package logger defines the interfaces used for logging.
package logger

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// Logger is the interface for loggers used in the soup components.
type Logger interface {
	Debug(...interface{})
	Debugln(...interface{})
	Debugf(string, ...interface{})

	Info(...interface{})
	Infoln(...interface{})
	Infof(string, ...interface{})

	Warn(...interface{})
	Warnln(...interface{})
	Warnf(string, ...interface{})

	Error(...interface{})
	Errorln(...interface{})
	Errorf(string, ...interface{})

	Fatal(...interface{})
	Fatalln(...interface{})
	Fatalf(string, ...interface{})

	WithFields(fields Fields) Logger
	With(key string, value interface{}) Logger

	SetFormat(string) error
	SetLevel(string) error
}
