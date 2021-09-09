// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgrpcmeta

import (
	"context"
	"github.com/rookie-ninja/rk-common/common"
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-grpc/interceptor"
	"github.com/rookie-ninja/rk-grpc/interceptor/context"
	"google.golang.org/grpc"
	"time"
)

// UnaryServerInterceptor Add common headers as extension style in http response.
// The key is defined as bellow:
// 1: X-Request-Id: Request id generated by interceptor.
// 2: X-<Prefix-App: Application name.
// 3: X-<Prefix>-App-Version: Version of application.
// 4: X-<Prefix>-App-Unix-Time: Unix time of current application.
// 5: X-<Prefix>-Request-Received-Time: Time of current request received by application.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	set := newOptionSet(rkgrpcinter.RpcTypeUnaryServer, opts...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = rkgrpcinter.WrapContextForServer(ctx)
		rkgrpcinter.AddToServerContextPayload(ctx, rkgrpcinter.RpcEntryNameKey, set.EntryName)

		requestId := rkcommon.GenerateRequestId()
		rkgrpcctx.AddHeaderToClient(ctx, rkgrpcctx.RequestIdKey, requestId)

		event := rkgrpcctx.GetEvent(ctx)
		event.SetRequestId(requestId)
		event.SetEventId(requestId)

		rkgrpcctx.AddHeaderToClient(ctx, set.AppNameKey, rkentry.GlobalAppCtx.GetAppInfoEntry().AppName)
		rkgrpcctx.AddHeaderToClient(ctx, set.AppVersionKey, rkentry.GlobalAppCtx.GetAppInfoEntry().Version)

		now := time.Now()
		rkgrpcctx.AddHeaderToClient(ctx, set.AppUnixTimeKey, now.Format(time.RFC3339Nano))
		rkgrpcctx.AddHeaderToClient(ctx, set.ReceivedTimeKey, now.Format(time.RFC3339Nano))

		resp, err := handler(ctx, req)
		rkgrpcinter.AddToClientContextPayload(ctx, rkgrpcinter.RpcErrorKey, err)

		return resp, err

	}
}

// StreamServerInterceptor Add common headers as extension style in http response.
// The key is defined as bellow:
// 1: X-Request-Id: Request id generated by interceptor.
// 2: X-<Prefix-App: Application name.
// 3: X-<Prefix>-App-Version: Version of application.
// 4: X-<Prefix>-App-Unix-Time: Unix time of current application.
// 5: X-<Prefix>-Request-Received-Time: Time of current request received by application.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	set := newOptionSet(rkgrpcinter.RpcTypeStreamServer, opts...)

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Before invoking
		wrappedStream := rkgrpcctx.WrapServerStream(stream)
		wrappedStream.WrappedContext = rkgrpcinter.WrapContextForServer(wrappedStream.WrappedContext)

		rkgrpcinter.AddToServerContextPayload(wrappedStream.WrappedContext, rkgrpcinter.RpcEntryNameKey, set.EntryName)

		requestId := rkcommon.GenerateRequestId()
		rkgrpcctx.AddHeaderToClient(wrappedStream.WrappedContext, rkgrpcctx.RequestIdKey, requestId)

		event := rkgrpcctx.GetEvent(wrappedStream.WrappedContext)
		event.SetRequestId(requestId)
		event.SetEventId(requestId)

		rkgrpcctx.AddHeaderToClient(wrappedStream.WrappedContext, set.AppNameKey, rkentry.GlobalAppCtx.GetAppInfoEntry().AppName)
		rkgrpcctx.AddHeaderToClient(wrappedStream.WrappedContext, set.AppVersionKey, rkentry.GlobalAppCtx.GetAppInfoEntry().Version)

		now := time.Now()
		rkgrpcctx.AddHeaderToClient(wrappedStream.WrappedContext, set.AppUnixTimeKey, now.Format(time.RFC3339Nano))
		rkgrpcctx.AddHeaderToClient(wrappedStream.WrappedContext, set.ReceivedTimeKey, now.Format(time.RFC3339Nano))

		return handler(srv, wrappedStream)
	}
}
