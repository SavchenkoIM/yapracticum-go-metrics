package middlware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/grpcimp"
	"yaprakticum-go-track2/internal/grpcimp/grpccommon"
)

type GRPCClientMiddleware struct {
	Cfg config.ClientConfig
}

func (gmw GRPCClientMiddleware) AddHMAC256(ctx context.Context, method string, req interface{},
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	if gmw.Cfg.Key != "" {
		reqp := req.(*grpcimp.UpdateMetricsRequest)
		b := grpccommon.MetricDataToByteSlice(reqp.Data)

		hmc := hmac.New(sha256.New, []byte(gmw.Cfg.Key))
		hmc.Write(b)
		ctx = metadata.AppendToOutgoingContext(ctx, "HashSHA256", hex.EncodeToString(hmc.Sum(nil)))
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}

func (gmw GRPCClientMiddleware) AddAgentIP(ctx context.Context, method string, req interface{},
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	if gmw.Cfg.RealIP != nil {
		ctx = metadata.AppendToOutgoingContext(ctx, "X-Real-IP", gmw.Cfg.RealIP.String())
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
