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

	runComplete := newBoolMutex()
	shutdownComplete := newBoolMutex()

	var runErr error
	shutdownErr := newErrorMutex()

	go func() {
		sigHupOrTerm := <-signalChannel

		//intentionally only increment the group if a signal has been received
		//Otherwise, a race condition exists where the shutdown handler may be
		//terminated prematurely
		wg.Add(1)
		defer wg.Done()

		defer shutdownComplete.setTrue()

		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					h.Logger.Error("shutdown handler panicked with error:", err)
					shutdownErr.set(err)
				} else {
					h.Logger.Error("shutdown handler panicked:", r)
					shutdownErr.set(fmt.Errorf("shutdown handler panicked: %v", r))
				}
			}
		}()

		if runComplete.read() {
			h.Logger.Debug("run completed, not running shutdown handler")
			return
		}

		h.Logger.Debug(fmt.Sprintf("Received shutdown signal: %s", sigHupOrTerm))
		// h.Logger.Debug("Calling shutdown handler")
		shutdownErr.set(h.ShutdownFunc(sigHupOrTerm))
		// h.Logger.Debug("Shutdown handler complete")

		if shutdownErr.read() != nil {
			h.Logger.Error(fmt.Errorf("run: %w", shutdownErr.read()))
		}
	}()

	// h.Logger.Debug("Calling run handler")
	runErr = h.RunFunc()
	runComplete.setTrue()
	// h.Logger.Debug("Run handler complete")

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
