package middlware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"time"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/grpcimp"
	"yaprakticum-go-track2/internal/grpcimp/grpccommon"
	"yaprakticum-go-track2/internal/shared"
)

type GRPCServerMiddleware struct {
	Cfg config.ServerConfig
}

func (gmw GRPCServerMiddleware) WithHMAC256Check(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var token string
	var values []string
	var ok bool
	var md metadata.MD
	if md, ok = metadata.FromIncomingContext(ctx); ok {
		values = md.Get("HashSHA256")
		if len(values) > 0 {
			token = values[0]
		} else {
			if gmw.Cfg.Key != "" {
				shared.Logger.Info("No Hash info in request found")
			}
		}
	}

	if ok && len(values) > 0 {

		hmacSha256, err := hex.DecodeString(token)
		if err != nil {
			shared.Logger.Info("Incorrect Header HashSHA256")
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		reqp := req.(*grpcimp.UpdateMetricsRequest)
		b := grpccommon.MetricDataToByteSlice(reqp.Data)

		hmc := hmac.New(sha256.New, []byte(gmw.Cfg.Key))
		hmc.Write(b)

		if !hmac.Equal(hmc.Sum(nil), hmacSha256) {
			shared.Logger.Info("Incorrect HMAC SHA256")
			return nil, status.Error(codes.Unauthenticated, "Incorrect HMAC passed")
		}

	}

	return handler(ctx, req)
}

func (gmw GRPCServerMiddleware) WithLogging(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// Calls the handler
	h, err := handler(ctx, req)

	// Logging with grpclog (grpclog.LoggerV2)
	shared.Logger.Sugar().Infof("gRPC Method: %+v, Runtime: %d msec, Error:%v",
		info.FullMethod, time.Since(start).Milliseconds(), err)

	return h, err
}

func (gmw GRPCServerMiddleware) WithTrustedNetworkCheck(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	// Also it is possible to use ":address" automatic metadata in the way:
	/*
		addr := "localhost:3200"
		addr = strings.Split(addr, ":")[0]

		var addrs []string
		var err error
		ip := net.ParseIP(addr)

		if ip == nil {
			addrs, err = net.LookupHost(addr)
		}

		println(ip, addrs)
	*/

	if gmw.Cfg.TrustedSubnet == nil {
		return handler(ctx, req)
	}

	var token string
	var values []string
	var ok bool
	var md metadata.MD
	if md, ok = metadata.FromIncomingContext(ctx); ok {
		values = md.Get("X-Real-IP")
		if len(values) > 0 {
			token = values[0]
		} else {
			shared.Logger.Info("No agent IP info in request found")
			return nil, status.Error(codes.PermissionDenied, "No agent IP info in request found")
		}
	} else {
		shared.Logger.Info("Failed to parse request metadata")
		return nil, status.Error(codes.PermissionDenied, "Failed to parse request metadata")
	}

	ip := net.ParseIP(token)
	if !gmw.Cfg.TrustedSubnet.Contains(ip) {
		shared.Logger.Info("Agent IP is not in trusted subnet")
		return nil, status.Error(codes.PermissionDenied, "Agent IP is not in trusted subnet")
	}

	return handler(ctx, req)
}
