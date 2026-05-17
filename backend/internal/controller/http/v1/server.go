package v1

import (
	"context"
	"github.com/mosgor/Evently/backend/config"
	"net/http"
	"time"
)

type HttpServer struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

func NewServer(cfg config.Server, handler http.Handler) *HttpServer {
	server := &http.Server{
		Addr:         "0.0.0.0:" + cfg.Port,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.Timeout * 2,
		Handler:      handler,
	}

	h := &HttpServer{
		server:          server,
		notify:          make(chan error, 1),
		shutdownTimeout: cfg.Timeout / 2,
	}
	h.start()

	return h
}

func (h *HttpServer) start() {
	go func() {
		h.notify <- h.server.ListenAndServe()
		close(h.notify)
	}()
}

func (h *HttpServer) Notify() <-chan error {
	return h.notify
}

func (h *HttpServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), h.shutdownTimeout)
	defer cancel()

	return h.server.Shutdown(ctx)
}
