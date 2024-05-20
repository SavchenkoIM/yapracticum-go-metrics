package server

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/grpcimp"
	"yaprakticum-go-track2/internal/grpcimp/server/middlware"
	"yaprakticum-go-track2/internal/storage/storagecommons"
)

type MetricsGRPCServer struct {
	grpcimp.UnimplementedMetricsServer
	dataStorage storagecommons.Storager
	gsrv        *grpc.Server
	srv         net.Listener
	cfg         config.ServerConfig
	logger      *zap.Logger
}

func NewGRPCMetricsServer(dataStorage storagecommons.Storager, cfg config.ServerConfig, logger *zap.Logger) *MetricsGRPCServer {
	return &MetricsGRPCServer{dataStorage: dataStorage, cfg: cfg, logger: logger}
}

func (s *MetricsGRPCServer) ListenAndServeAsync() {
	var err error
	s.srv, err = net.Listen("tcp", s.cfg.EndpGRPC)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	mw := middlware.GRPCServerMiddleware{Cfg: s.cfg}
	s.gsrv = grpc.NewServer(grpc.ChainUnaryInterceptor(mw.WithLogging, mw.WithHMAC256Check, mw.WithTrustedNetworkCheck))
	grpcimp.RegisterMetricsServer(s.gsrv, s)

	go func() {
		s.logger.Info("gRPC server running at " + s.cfg.EndpGRPC)

		if err := s.gsrv.Serve(s.srv); err != nil {
			s.logger.Error(err.Error())
			return
		}
	}()
}

func (s *MetricsGRPCServer) Shutdown(context.Context) error {
	s.gsrv.Stop()
	return nil
}

func (s *MetricsGRPCServer) UpdateMetrics(ctx context.Context, r *grpcimp.UpdateMetricsRequest) (*grpcimp.UpdateMetricsResponse, error) {
	res := grpcimp.UpdateMetricsResponse{}
	var dta storagecommons.MetricsDB

	for _, v := range r.Data {
		v := v
		m := storagecommons.Metrics{}
		m.ID = v.Name

		switch v.Type {
		case grpcimp.MetricData_COUNTER:
			m = storagecommons.Metrics{
				Delta: &v.Delta,
				Value: nil,
				ID:    v.Name,
				MType: "counter",
			}
		case grpcimp.MetricData_GAUGE:
			m = storagecommons.Metrics{
				Delta: nil,
				Value: &v.Value,
				ID:    v.Name,
				MType: "gauge",
			}
		}

		dta.MetricsDB = append(dta.MetricsDB, m)
	}

	err := s.dataStorage.WriteDataMulti(ctx, dta)
	if err != nil {
		res.Error = err.Error()
	}

	return &res, err
}
