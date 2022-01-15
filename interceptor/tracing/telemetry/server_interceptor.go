// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgrpctrace

import (
	"context"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidtrace "github.com/rookie-ninja/rk-entry/middleware/tracing"
	"github.com/rookie-ninja/rk-grpc/interceptor"
	"github.com/rookie-ninja/rk-grpc/interceptor/context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor Create new unary server interceptor.
func UnaryServerInterceptor(opts ...rkmidtrace.Option) grpc.UnaryServerInterceptor {
	set := rkmidtrace.NewOptionSet(opts...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = rkgrpcinter.WrapContextForServer(ctx)
		rkgrpcinter.AddToServerContextPayload(ctx, rkmid.EntryNameKey, set.GetEntryName())
		rkgrpcinter.AddToServerContextPayload(ctx, rkmid.TracerKey, set.GetTracer())
		rkgrpcinter.AddToServerContextPayload(ctx, rkmid.TracerProviderKey, set.GetProvider())
		rkgrpcinter.AddToServerContextPayload(ctx, rkmid.PropagatorKey, set.GetPropagator())

		beforeCtx := set.BeforeCtx(nil, false)
		beforeCtx.Input.UrlPath = info.FullMethod
		beforeCtx.Input.RequestCtx = ctx
		beforeCtx.Input.SpanName = info.FullMethod

		// metadata carrier
		incomingMD, _ := metadata.FromIncomingContext(ctx)
		beforeCtx.Input.Carrier = &rkgrpcctx.GrpcMetadataCarrier{Md: &incomingMD}
		// grpc related meta
		beforeCtx.Input.Attributes = append(beforeCtx.Input.Attributes, grpcInfoToAttributes(
			ctx, info.FullMethod, "UnaryServer")...)

		set.Before(beforeCtx)

		// new context and span
		ctx = beforeCtx.Output.NewCtx
		if beforeCtx.Output.Span != nil {
			rkgrpcinter.AddToServerContextPayload(ctx, rkmid.SpanKey, beforeCtx.Output.Span)
			rkgrpcctx.AddHeaderToClient(ctx, rkmid.HeaderTraceId, beforeCtx.Output.Span.SpanContext().TraceID().String())
		}

		// call handler
		resp, err := handler(ctx, req)

		var afterCtx *rkmidtrace.AfterCtx
		if err != nil {
			s, _ := status.FromError(err)
			afterCtx = set.AfterCtx(int(codes.Error), s.Message())
			afterCtx.Input.Attributes = append(afterCtx.Input.Attributes,
				attribute.Int("grpc.code", int(s.Code())),
				attribute.String("grpc.status", s.Code().String()))
		} else {
			afterCtx = set.AfterCtx(200, "")
			afterCtx.Input.Attributes = append(afterCtx.Input.Attributes,
				attribute.Int("grpc.code", int(codes.Ok)),
				attribute.String("grpc.status", codes.Ok.String()))
		}

		set.After(beforeCtx, afterCtx)

		return resp, err
	}
}

// StreamServerInterceptor Create new stream server interceptor.
func StreamServerInterceptor(opts ...rkmidtrace.Option) grpc.StreamServerInterceptor {
	set := rkmidtrace.NewOptionSet(opts...)

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Before invoking
		wrappedStream := rkgrpcctx.WrapServerStream(stream)
		wrappedStream.WrappedContext = rkgrpcinter.WrapContextForServer(wrappedStream.WrappedContext)

		rkgrpcinter.AddToServerContextPayload(wrappedStream.WrappedContext, rkmid.EntryNameKey, set.GetEntryName())
		rkgrpcinter.AddToServerContextPayload(wrappedStream.WrappedContext, rkmid.TracerKey, set.GetTracer())
		rkgrpcinter.AddToServerContextPayload(wrappedStream.WrappedContext, rkmid.TracerProviderKey, set.GetProvider())
		rkgrpcinter.AddToServerContextPayload(wrappedStream.WrappedContext, rkmid.PropagatorKey, set.GetPropagator())

		beforeCtx := set.BeforeCtx(nil, false)
		beforeCtx.Input.UrlPath = info.FullMethod
		beforeCtx.Input.RequestCtx = wrappedStream.WrappedContext
		beforeCtx.Input.SpanName = info.FullMethod

		// metadata carrier
		incomingMD, _ := metadata.FromIncomingContext(wrappedStream.WrappedContext)
		beforeCtx.Input.Carrier = &rkgrpcctx.GrpcMetadataCarrier{Md: &incomingMD}

		// grpc related meta
		beforeCtx.Input.Attributes = append(beforeCtx.Input.Attributes, grpcInfoToAttributes(
			wrappedStream.WrappedContext, info.FullMethod, "UnaryServer")...)

		set.Before(beforeCtx)

		// new context and span
		wrappedStream.WrappedContext = beforeCtx.Output.NewCtx
		rkgrpcinter.AddToServerContextPayload(wrappedStream.WrappedContext, rkmid.SpanKey, beforeCtx.Output.Span)

		// call handler
		err := handler(srv, wrappedStream)
		//rkgrpcinter.AddToServerContextPayload(wrappedStream.WrappedContext, rkgrpcinter.GrpcErrorKey, err)

		var afterCtx *rkmidtrace.AfterCtx
		attrs := make([]attribute.KeyValue, 0)
		if err != nil {
			s, _ := status.FromError(err)
			afterCtx = set.AfterCtx(int(codes.Error), s.Message())
			attrs = append(attrs,
				attribute.Int("grpc.code", int(s.Code())),
				attribute.String("grpc.status", s.Code().String()))
		} else {
			afterCtx = set.AfterCtx(200, "")
			attrs = append(attrs,
				attribute.Int("grpc.code", int(codes.Ok)),
				attribute.String("grpc.status", codes.Ok.String()))
		}

		set.After(beforeCtx, afterCtx)

		return err
	}
}

// Convert grpc information into attributes.
func grpcInfoToAttributes(ctx context.Context, method, rpcType string) []attribute.KeyValue {
	remoteIp, remotePort, _ := rkgrpcinter.GetRemoteAddressSet(ctx)
	grpcService, grpcMethod := rkgrpcinter.GetGrpcInfo(method)
	gwMethod, gwPath, gwScheme, gwUserAgent := rkgrpcinter.GetGwInfo(rkgrpcctx.GetIncomingHeaders(ctx))

	return []attribute.KeyValue{
		attribute.String("local.IP", rkgrpcinter.LocalIp.String),
		attribute.String("local.hostname", rkgrpcinter.LocalHostname.String),
		attribute.String("remote.IP", remoteIp),
		attribute.String("remote.port", remotePort),
		attribute.String("grpc.service", grpcService),
		attribute.String("grpc.method", grpcMethod),
		attribute.String("gw.method", gwMethod),
		attribute.String("gw.path", gwPath),
		attribute.String("gw.scheme", gwScheme),
		attribute.String("gw.userAgent", gwUserAgent),
		attribute.String("server.type", rpcType),
	}
}
