package server

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"http/v1/dev/internal/server/web"
	"log/slog"
	"net"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("template has not been found")

type Option func(s *Server)

type Server struct {
	db        *sql.DB
	logger    *slog.Logger
	addr      string
	dbPath    string
	templates map[string]*template.Template
}

func NewServer(opts ...Option) (*Server, error) {
	srv := &Server{
		addr:   ":8080",
		logger: slog.Default(),
		dbPath: "dev/internal/server/db.json",
		db:     nil,
	}

	for _, opt := range opts {
		opt(srv)
	}

	tmps, err := parseTemplates(web.TemplatesFS)
	if err != nil {
		return nil, err
	}

	srv.templates = tmps

	return srv, nil
}

func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

func WithDatabase(db *sql.DB) Option {
	return func(s *Server) {
		s.db = db
	}
}

func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

func WithDBPath(path string) Option {
	return func(s *Server) {
		s.dbPath = path
	}
}

func (s *Server) Run(ctx context.Context) error {
	httpSrv := &http.Server{
		Addr:              s.addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,

		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("Http server starting", "addr", s.addr)

		err := httpSrv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}

		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info("Shutdown signal received, draining connections")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return httpSrv.Shutdown(shutdownCtx)
	}
}

func (s *Server) Handler() http.Handler {
	return s.routes()
}
