package main

import (
	"context"
	"flag"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	howdah_server "howdah/internal/pkg/howdah-server"
	"howdah/pb"
	"net"
	"net/http"
	"strconv"
	"github.com/sirupsen/logrus"

)

func main() {
	logger := logrus.New()

	var (
		grpcPort *int
		gwPort *int
		listener net.Listener
	)

	gwPort = flag.Int("port", 9090, "")
	grpcPort = flag.Int("grpc_port", 9091, "")
	flag.Parse()

	listener, err := net.Listen("tcp", ":" + strconv.Itoa(*grpcPort))
	defer listener.Close()
	if err != nil {
		logger.Fatalf("Failed to listen on port %d.\n%s",
			*grpcPort)
	}
	adminStore := howdah_server.NewMockAdminStore()
	authService, err := howdah_server.NewAuthService(adminStore)
	if err != nil {
		logger.Fatalf("Failed on building AuthService.")
	}

	server := grpc.NewServer()
	defer server.Stop()
	pb.RegisterAuthServiceServer(server, authService)
	go func() {
		logger.Fatalln(server.Serve(listener))
	}()

	logger.Infof("Serving gRPC Server(AuthService) on 0.0.0.0: %d",
		*grpcPort)
	conn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:" + strconv.Itoa(*grpcPort),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		logger.Fatalln("Failed to dial server:", err)
	}

	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard, &runtime.JSONPb{
				EmitDefaults: true,
			}),
	)

	err = pb.RegisterAuthServiceHandler(
		context.Background(),
		gwMux, conn,
	)

	if err != nil {
		logger.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:              ":" + strconv.Itoa(*gwPort),
		Handler:           gwMux,
	}

	logger.Infof("Serving gRPC-Gateway on 0.0.0.0:%s\n",
		gwServer.Addr)
	logger.Fatalln(gwServer.ListenAndServe())

}
