package signals

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/danifv27/soup/internal/application/signals"
	tl "github.com/danifv27/soup/internal/infrastructure/logger/testing"
)

type termSignal struct {
	signal syscall.Signal
	delay  time.Duration
}

func (s *termSignal) run(t *testing.T) {
	t.Log("sleeping for", s.delay)
	time.Sleep(s.delay)

	intSignal := int(s.signal)
	cmdArgs := []string{"kill", fmt.Sprintf("-%s", strconv.Itoa(intSignal)), strconv.Itoa(syscall.Getpid())}

	t.Log("creating command:", cmdArgs)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	t.Log("sending SIGALRM")
	err := cmd.Run()
	t.Log("sent SIGALRM")
	assert.Nil(t, err)
}

func run(t *testing.T, run signals.SignalHandlerRunFunc, shutdown signals.SignalHandlerShutdownFunc, termSig termSignal, recognizedSignals ...os.Signal) error {

	tl := tl.NewTestingLogger(t)
	handler := &Handler{
		Signals:      recognizedSignals,
		Logger:       tl,
		RunFunc:      run,
		ShutdownFunc: shutdown,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	var runErr error

	go func() {
		runErr = handler.Run()

		t.Log("completing wait group")
		wg.Done()
	}()

	termSig.run(t)

	wg.Wait()

	return runErr
}

func Test_ShutdownSuccessful(t *testing.T) {
	shouldRun := true
	shutdownComplete := false

	err := run(
		t,
		func() error {
			for shouldRun {
				t.Log("run func running")
				time.Sleep(time.Millisecond * 10)
			}

			return nil
		},
		func(s os.Signal) error {
			assert.Equal(t, syscall.SIGALRM, s)
			shouldRun = false
			shutdownComplete = true
			return nil
		},
		termSignal{signal: syscall.SIGALRM},
		syscall.SIGALRM,
	)

	assert.True(t, shutdownComplete)
	assert.Nil(t, err)
}

func Test_ShutdownShutdownError(t *testing.T) {
	expectedErr := errors.New("shutdown failed")
	shouldRun := true
	shutdownComplete := false

	err := run(
		t,
		func() error {
			for shouldRun {
				t.Log("run func running")
				time.Sleep(time.Millisecond * 10)
			}

			return nil
		},
		func(s os.Signal) error {
			assert.Equal(t, syscall.SIGALRM, s)
			shouldRun = false
			shutdownComplete = true
			return expectedErr
		},
		termSignal{signal: syscall.SIGALRM},
		syscall.SIGALRM,
	)

	assert.True(t, shutdownComplete)
	assert.Equal(t, expectedErr, err)
}

func Test_ShutdownShutdownPanic(t *testing.T) {
	expectedErr := errors.New("shutdown failed")

	err := run(
		t,
		func() error {
			time.Sleep(time.Millisecond * 20)
			return nil
		},
		func(s os.Signal) error {
			assert.Equal(t, syscall.SIGALRM, s)
			panic(expectedErr)
			return nil
		},
		termSignal{signal: syscall.SIGALRM},
		syscall.SIGALRM,
	)

	assert.Equal(t, expectedErr, err)
}

func Test_ShutdownRunFailure(t *testing.T) {
	expectedErr := errors.New("run failed")
	shouldRun := true
	shutdownComplete := false

	err := run(
		t,
		func() error {
			for shouldRun {
				t.Log("run func running")
				time.Sleep(time.Millisecond * 10)
			}

			return expectedErr
		},
		func(s os.Signal) error {
			assert.Equal(t, syscall.SIGALRM, s)
			shouldRun = false
			shutdownComplete = true
			return nil
		},
		termSignal{signal: syscall.SIGALRM},
		syscall.SIGALRM,
	)

	assert.True(t, shutdownComplete)
	assert.Nil(t, err)
}

func Test_RunFailure(t *testing.T) {
	expectedErr := errors.New("run failed")

	err := run(
		t,
		func() error {
			time.Sleep(time.Millisecond * 10)
			return expectedErr
		},
		func(s os.Signal) error {
			assert.Fail(t, "shutdown handler should not be called")
			return nil
		},
		termSignal{
			signal: syscall.SIGALRM,
			delay:  time.Millisecond * 20,
		},
		syscall.SIGALRM,
	)

	assert.Equal(t, expectedErr, err)
}