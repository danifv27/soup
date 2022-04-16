package rest

import (
	"context"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/infrastructure/executor"
	"github.com/danifv27/soup/internal/infrastructure/rest/probes"
	"github.com/danifv27/soup/internal/infrastructure/signals"
	"github.com/gorilla/mux"
)

//Server
type Server struct {
	apps       application.Applications
	httpServer *http.Server
	router     *mux.Router
	root       string
}

//NewServer
func NewServer(apps application.Applications) *Server {

	httpServer := &Server{
		apps:   apps,
		router: mux.NewRouter(),
	}

	return httpServer
}

func (s *Server) SetActuatorRoot(root string) {

	s.root = root
}

func (s *Server) addProbeRoutes() {

	s.router.HandleFunc(s.root+"/liveness", probes.NewHandler(s.apps).GetLiveness).Methods("GET")
	s.router.HandleFunc(s.root+"/readiness", probes.NewHandler(s.apps).GetReadiness).Methods("GET")
}

func (s *Server) Start(address string, wg *sync.WaitGroup) {

	s.addProbeRoutes()
	s.httpServer = &http.Server{
		Addr:    address,
		Handler: s.router,
	}

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, s.apps.LoggerService)
	h.SetRunFunc(func() error {
		s.apps.LoggerService.Infof("starting actuators on %s", address)
		err := s.httpServer.ListenAndServe() // Blocks!
		if err != http.ErrServerClosed {
			s.apps.LoggerService.With("err", err).Error("http server stopped unexpected")
			s.Shutdown()
		} else {
			s.apps.LoggerService.With("err", err).Info("http server stopped")
		}

		return nil
	})
	h.SetShutdownFunc(func(sig os.Signal) error {
		s.Shutdown()

		return nil
	})

	loop := executor.NewLoop(s.apps, &h)
	loop.Exec(wg)
}

func (s *Server) Shutdown() {
	if s.httpServer != nil {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			s.apps.LoggerService.With("err", err).Error("failed to shutdown http server gracefully")
		} else {
			s.httpServer = nil
		}
	}
}
