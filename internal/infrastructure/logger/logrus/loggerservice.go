package logrus

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/danifv27/soup/internal/application/logger"
	"github.com/sirupsen/logrus"
)

// LoggerService provides a logrus implementation of the Service
type LoggerService struct {
	entry *logrus.Entry
}

func (l LoggerService) SetLevel(level string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("SetLevel: %w", err)
	}

	l.entry.Logger.Level = lvl

	return nil
}

func (l LoggerService) SetFormat(format string) error {
	u, err := url.Parse(format)
	if err != nil {
		return fmt.Errorf("SetFormat: %w", err)
	}
	if u.Scheme != "logger" {
		return fmt.Errorf("invalid scheme %s", u.Scheme)
	}
	jsonq := u.Query().Get("json")
	if jsonq == "true" {
		l.entry.Logger.Formatter = &logrus.JSONFormatter{}
	}

	switch u.Opaque {
	// case "syslog":
	// 	if setSyslogFormatter == nil {
	// 		return fmt.Errorf("system does not support syslog")
	// 	}
	// 	appname := u.Query().Get("appname")
	// 	facility := u.Query().Get("local")
	// 	return setSyslogFormatter(l, appname, facility)
	// case "eventlog":
	// 	if setEventlogFormatter == nil {
	// 		return fmt.Errorf("system does not support eventlog")
	// 	}
	// 	name := u.Query().Get("name")
	// 	debugAsInfo := false
	// 	debugAsInfoRaw := u.Query().Get("debugAsInfo")
	// 	if parsedDebugAsInfo, err := strconv.ParseBool(debugAsInfoRaw); err == nil {
	// 		debugAsInfo = parsedDebugAsInfo
	// 	}
	// 	return setEventlogFormatter(l, name, debugAsInfo)
	case "stdout":
		l.entry.Logger.Out = os.Stdout
	case "stderr":
		l.entry.Logger.Out = os.Stderr
	default:
		return fmt.Errorf("unsupported logger %q", u.Opaque)
	}
	return nil
}

// NewLoggerService creates a new `log.Logger` from the provided entry
func NewLoggerService() logger.Logger {

	l := logrus.New()

	out := LoggerService{
		entry: logrus.NewEntry(l),
	}

	return &out
}

// Debug logs a message at level Debug on the standard logger.
func (l LoggerService) Debug(args ...interface{}) {

	l.sourced().Debug(args...)
}

// Debug logs a message at level Debug on the standard logger.
func (l LoggerService) Debugln(args ...interface{}) {
	l.sourced().Debugln(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func (l LoggerService) Debugf(format string, args ...interface{}) {
	l.sourced().Debugf(format, args...)
}

// Info logs a message at level Info on the standard logger.
func (l LoggerService) Info(args ...interface{}) {
	l.sourced().Info(args...)
}

// Info logs a message at level Info on the standard logger.
func (l LoggerService) Infoln(args ...interface{}) {
	l.sourced().Infoln(args...)
}

// Infof logs a message at level Info on the standard logger.
func (l LoggerService) Infof(format string, args ...interface{}) {
	l.sourced().Infof(format, args...)
}

// Warn logs a message at level Warn on the standard logger.
func (l LoggerService) Warn(args ...interface{}) {
	l.sourced().Warn(args...)
}

// Warn logs a message at level Warn on the standard logger.
func (l LoggerService) Warnln(args ...interface{}) {
	l.sourced().Warnln(args...)
}

// Warnf logs a message at level Warn on the standard logger.
func (l LoggerService) Warnf(format string, args ...interface{}) {
	l.sourced().Warnf(format, args...)
}

// Error logs a message at level Error on the standard logger.
func (l LoggerService) Error(args ...interface{}) {
	l.sourced().Error(args...)
}

// Error logs a message at level Error on the standard logger.
func (l LoggerService) Errorln(args ...interface{}) {
	l.sourced().Errorln(args...)
}

// Errorf logs a message at level Error on the standard logger.
func (l LoggerService) Errorf(format string, args ...interface{}) {
	l.sourced().Errorf(format, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (l LoggerService) Fatal(args ...interface{}) {
	l.sourced().Fatal(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (l LoggerService) Fatalln(args ...interface{}) {
	l.sourced().Fatalln(args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func (l LoggerService) Fatalf(format string, args ...interface{}) {
	l.sourced().Fatalf(format, args...)
}

func (l LoggerService) With(key string, value interface{}) logger.Logger {

	out := LoggerService{
		entry: l.entry.WithField(key, value),
	}

	return &out
}

func (l LoggerService) WithFields(fields logger.Fields) logger.Logger {

	out := LoggerService{
		entry: l.entry.WithFields(logrus.Fields(fields)),
	}

	return &out
}

// sourced adds a source field to the logger that contains
// the file name and line where the logging happened.
func (l LoggerService) sourced() *logrus.Entry {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		file = file[slash+1:]
	}

	return l.entry.WithField("source", fmt.Sprintf("%s:%d", file, line))
}
