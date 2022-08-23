// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"log"

	testdata "github.com/tegarajipangestu/rk-grpc/v2/example/middleware/proto/testdata"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var logger, _ = zap.NewDevelopment()

// In this example, we will create a simple gRpc client and enable RK style logging interceptor.
func main() {
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	// 1: Create grpc client
	conn, client := createClient(opts...)
	defer conn.Close()

	// 2: Add header
	ctx := metadata.NewOutgoingContext(context.TODO(), metadata.Pairs("domain", "test"))

	// 2: Call server
	if resp, err := client.SayHello(ctx, &testdata.HelloRequest{}); err != nil {
		logger.Fatal("Failed to send request to server.", zap.Error(err))
	} else {
		logger.Info(fmt.Sprintf("[Message]: %s", resp.String()))
	}
}

func createClient(opts ...grpc.DialOption) (*grpc.ClientConn, testdata.GreeterClient) {
	// 1: Set up a connection to the server.
	conn, err := grpc.DialContext(context.Background(), "localhost:8080", opts...)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// 2: Create grpc client
	client := testdata.NewGreeterClient(conn)

	return conn, client
}
