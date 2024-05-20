package client

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/grpcimp"
	"yaprakticum-go-track2/internal/grpcimp/client/middlware"
	"yaprakticum-go-track2/internal/storage/storagecommons"
)

type MetricsGRPCClient struct {
	conn      *grpc.ClientConn
	cfg       config.ClientConfig
	logger    *zap.Logger
	client    grpcimp.MetricsClient
	sendError chan error
}

func NewMetricsGRPCClient(cfg config.ClientConfig, logger *zap.Logger) *MetricsGRPCClient {
	return &MetricsGRPCClient{cfg: cfg, logger: logger, sendError: make(chan error, 1)}
}

func (c *MetricsGRPCClient) reconnectWorker(ctx context.Context) {
	for ctx.Err() == nil {
		select {
		case <-c.sendError:
			var err error
			c.conn, err = grpc.DialContext(ctx, c.cfg.Endp, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				c.logger.Error("Failed to connect to server", zap.Error(err))
			}
		default:
			time.Sleep(time.Second)
		}
	}
}

func (c *MetricsGRPCClient) Start(ctx context.Context) {
	// Connection
	var err error
	mw := middlware.GRPCClientMiddleware{Cfg: c.cfg}
	c.conn, err = grpc.DialContext(ctx, c.cfg.Endp, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(mw.AddAgentIP, mw.AddHMAC256))
	if err != nil {
		c.logger.Error("Failed to connect to server", zap.Error(err))
	}
	c.client = grpcimp.NewMetricsClient(c.conn)

	go c.reconnectWorker(ctx)
}

func (c *MetricsGRPCClient) Stop(ctx context.Context) error {
	return c.conn.Close()
}

func (c *MetricsGRPCClient) SendMetricsData(ctx context.Context, data storagecommons.MetricsDB) error {
	// Data send
	mdata := make([]*grpcimp.MetricData, 0)
	for _, v := range data.MetricsDB {
		v := v
		typ := grpcimp.MetricData_UNSPECIFIED
		switch v.MType {
		case "gauge":
			typ = grpcimp.MetricData_GAUGE
		case "counter":
			typ = grpcimp.MetricData_COUNTER
		}
		var (
			val   float64
			delta int64
		)
		if v.Delta != nil {
			delta = *v.Delta
		}
		if v.Value != nil {
			val = *v.Value
		}

		mdata = append(mdata, &grpcimp.MetricData{
			Type:  typ,
			Name:  v.ID,
			Value: val,
			Delta: delta,
		})
	}

	_, err := c.client.UpdateMetrics(ctx, &grpcimp.UpdateMetricsRequest{Data: mdata})
	if err != nil {
		select {
		case c.sendError <- err:
		default:
		}
		return err
	}

	return nil
}
