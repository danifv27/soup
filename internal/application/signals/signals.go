package signals

import "os"

type SignalHandlerRunFunc func() error
type SignalHandlerShutdownFunc func(os.Signal) error

type SignalHandler interface {
	Run() error
	SetRunFunc(SignalHandlerRunFunc) error
	SetShutdownFunc(SignalHandlerShutdownFunc) error
}
