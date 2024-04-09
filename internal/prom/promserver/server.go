package promserver

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

type PromServer struct {
	http.Server
	logger *zap.Logger
}

func NewServer(addr string, logger *zap.Logger) *PromServer {
	if addr == "" {
		return nil
	}
	return &PromServer{Server: http.Server{Addr: addr, Handler: promhttp.Handler()}, logger: logger}
}

func (ps *PromServer) ListenAndServeAsync() {
	go func() {
		ps.logger.Info("Prom server running at " + ps.Server.Addr)
		if err := ps.ListenAndServe(); err != nil {
			ps.logger.Error(err.Error())
			return
		}
	}()
}
