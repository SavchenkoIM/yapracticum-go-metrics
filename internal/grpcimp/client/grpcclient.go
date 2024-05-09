package client

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"yaprakticum-go-track2/internal/grpcimp"
	"yaprakticum-go-track2/internal/storage/storagecommons"
)

type MetricsGRPCClient struct {
	conn   *grpc.ClientConn
	addr   string
	logger *zap.Logger
}

func NewMetricsGRPCClient(addr string, logger *zap.Logger) *MetricsGRPCClient {
	return &MetricsGRPCClient{addr: addr, logger: logger}
}

func (c *MetricsGRPCClient) SendMetricsData(ctx context.Context, data storagecommons.MetricsDB) error {

	// Connection
	var err error
	c.conn, err = grpc.DialContext(ctx, c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		c.logger.Error("Failed to connect to server", zap.Error(err))
		return err
	}
	defer c.conn.Close()
	client := grpcimp.NewMetricsClient(c.conn)

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

	_, err = client.UpdateMetrics(context.Background(), &grpcimp.UpdateMetricsRequest{Data: mdata})
	if err != nil {
		return err
	}

	return nil
}
