// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgrpcjwt

import (
	"context"
	rkmid "github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/jwt"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
	"testing"
)

func TestUnaryServerInterceptor(t *testing.T) {
	beforeCtx := rkmidjwt.NewBeforeCtx()
	mock := rkmidjwt.NewOptionSetMock(beforeCtx)
	inter := UnaryServerInterceptor(rkmidjwt.WithMockOptionSet(mock))

	// case 1: with error response
	beforeCtx.Output.ErrResp = rkmid.GetErrorBuilder().New(http.StatusInternalServerError, "")
	_, err := inter(NewUnaryServerInput())
	assert.NotNil(t, err)

	// case 2: happy case
	beforeCtx.Output.ErrResp = nil
	_, err = inter(NewUnaryServerInput())
	assert.Nil(t, err)
}

func TestStreamServerInterceptor(t *testing.T) {
	beforeCtx := rkmidjwt.NewBeforeCtx()
	mock := rkmidjwt.NewOptionSetMock(beforeCtx)
	inter := StreamServerInterceptor(rkmidjwt.WithMockOptionSet(mock))

	// case 1: with error response
	beforeCtx.Output.ErrResp = rkmid.GetErrorBuilder().New(http.StatusInternalServerError, "")
	err := inter(NewStreamServerInput())
	assert.NotNil(t, err)

	// case 2: happy case
	beforeCtx.Output.ErrResp = nil
	err = inter(NewStreamServerInput())
	assert.Nil(t, err)
}

// ************ Test utility ************

type ServerStreamMock struct {
	ctx context.Context
}

func (f ServerStreamMock) SetHeader(md metadata.MD) error {
	return nil
}

func (f ServerStreamMock) SendHeader(md metadata.MD) error {
	return nil
}

func (f ServerStreamMock) SetTrailer(md metadata.MD) {
	return
}

func (f ServerStreamMock) Context() context.Context {
	return f.ctx
}

func (f ServerStreamMock) SendMsg(m interface{}) error {
	return nil
}

func (f ServerStreamMock) RecvMsg(m interface{}) error {
	return nil
}

func NewUnaryServerInput() (context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) {
	ctx := context.TODO()
	info := &grpc.UnaryServerInfo{
		FullMethod: "ut-method",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}

	return ctx, nil, info, handler
}

func NewStreamServerInput() (interface{}, grpc.ServerStream, *grpc.StreamServerInfo, grpc.StreamHandler) {
	serverStream := &ServerStreamMock{ctx: context.TODO()}
	info := &grpc.StreamServerInfo{
		FullMethod: "ut-method",
	}
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	return nil, serverStream, info, handler
}
