package http

import (
	"context"
	"fmt"
	"homework/internal/usecase"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"
)

type Server struct {
	host   string
	port   uint16
	router *gin.Engine
}

type UseCases struct {
	Event  *usecase.Event
	Sensor *usecase.Sensor
	User   *usecase.User
}

func NewServer(useCases UseCases, options ...func(*Server)) *Server {
	r := gin.Default()
	setupRouter(r, useCases, NewWebSocketHandler(useCases))

	s := &Server{router: r, host: "localhost", port: 8080}
	for _, o := range options {
		o(s)
	}

	return s
}

func WithHost(host string) func(*Server) {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port uint16) func(*Server) {
	return func(s *Server) {
		s.port = port
	}
}

func (s *Server) Run(c context.Context) error {
	eg, appCtx := errgroup.WithContext(c)
	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)
	eg.Go(func() error {
		s := <-sigQuit
		return fmt.Errorf("captured signal: %s", s)
	})
	eg.Go(func() error {
		return s.router.Run(fmt.Sprintf("%s:%d", s.host, s.port))
	})
	eg.Go(func() error {
		<-appCtx.Done()
		return fmt.Errorf("done")
	})
	return eg.Wait()
}
