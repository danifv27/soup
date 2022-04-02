// Package testing defines the interfaces used for logging.
package testing

import (
	"testing"

	"github.com/danifv27/soup/internal/application/logger"
)

type TestingLogger struct {
	t *testing.T
}

// NewLoggerService creates a new `log.Logger` from the provided entry
func NewTestingLogger(t *testing.T) logger.Logger {

	out := TestingLogger{
		t: t,
	}

	return &out
}

func (TestingLogger) Debug(...interface{})          						{}
func (TestingLogger) Debugln(...interface{})        						{}
func (TestingLogger) Debugf(string, ...interface{}) 						{}

func (TestingLogger) Info(...interface{})           						{}
func (TestingLogger) Infof(string, ...interface{})  						{}
func (TestingLogger) Infoln(...interface{})        							{}
func (TestingLogger) Warn(...interface{})           						{}
func (TestingLogger) Warnf(string, ...interface{})  						{}
func (TestingLogger) Warnln(...interface{})        							{}
func (TestingLogger) Error(...interface{})          						{}
func (TestingLogger) Errorf(string, ...interface{}) 						{}
func (TestingLogger) Errorln(...interface{})        						{}
func (TestingLogger) Fatal(...interface{})          						{}
func (TestingLogger) Fatalf(string, ...interface{}) 						{}
func (TestingLogger) Fatalln(...interface{})        						{}

func (l TestingLogger) WithField(string, interface{}) logger.Logger 		{ return l }
func (l TestingLogger) WithFields(logger.Fields) logger.Logger				{ return l }
func (l TestingLogger) WithError(error) logger.Logger               		{ return l }
func (l TestingLogger) With(key string, value interface{}) logger.Logger	{ return l }

func (l TestingLogger) SetFormat(string) error 								{ return nil }
func (l TestingLogger) SetLevel(string) error 								{ return nil }
