package signals

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/application/signals"
)

type Handler struct {
	Signals      []os.Signal
	Logger       logger.Logger
	RunFunc      signals.SignalHandlerRunFunc
	ShutdownFunc signals.SignalHandlerShutdownFunc
}

// NewSignalHandler
func NewSignalHandler(s []os.Signal, l logger.Logger) Handler {

	return Handler{
		Signals: s,
		Logger:  l,
	}
}

func (h *Handler) SetRunFunc(f signals.SignalHandlerRunFunc) error {

	h.RunFunc = f

	return nil
}

func (h *Handler) SetShutdownFunc(f signals.SignalHandlerShutdownFunc) error {

	h.ShutdownFunc = f

	return nil
}

func (h *Handler) Run() error {

	if len(h.Signals) == 0 {
		h.Logger.Info("No signals set, using defaults of SIGHUP and SIGTERM")
		h.Signals = []os.Signal{syscall.SIGHUP, syscall.SIGTERM}
	}

	wg := sync.WaitGroup{}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel)
	// Reduce the impact of https://github.com/golang/go/issues/37942
	signal.Reset(syscall.SIGURG)
	runComplete := newBoolMutex()
	shutdownComplete := newBoolMutex()

	var runErr error
	shutdownErr := newErrorMutex()

	go func() {
		sigHupOrTerm := <-signalChannel
		h.Logger.WithFields(logger.Fields{
			"signal": sigHupOrTerm.String(),
		}).Debug("signal received")
		//intentionally only increment the group if a signal has been received
		//Otherwise, a race condition exists where the shutdown handler may be
		//terminated prematurely
		wg.Add(1)
		defer wg.Done()

		defer shutdownComplete.setTrue()

		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					h.Logger.With("err", err).Error("shutdown handler panicked")
					shutdownErr.set(err)
				} else {
					h.Logger.With("err", r).Error("shutdown handler panicked")
					shutdownErr.set(fmt.Errorf("shutdown handler panicked: %v", r))
				}
			}
		}()

		if runComplete.read() {
			h.Logger.Debug("run completed, not running shutdown handler")
			return
		}
		shutdownErr.set(h.ShutdownFunc(sigHupOrTerm))
		if shutdownErr.read() != nil {
			h.Logger.Error(fmt.Errorf("run: %w", shutdownErr.read()))
		}
	}()
	runErr = h.RunFunc()
	runComplete.setTrue()
	if runErr != nil {
		h.Logger.Error(fmt.Errorf("run: %w", runErr))
	}

	wg.Wait()

	//only log the run handler failure if shutdown wasn't executed
	//This is because structs like net/http.Server _always_ return errors when
	//methods like ListenAndServe() are called
	if shutdownComplete.read() {
		return shutdownErr.read()
	}

	return runErr
}
