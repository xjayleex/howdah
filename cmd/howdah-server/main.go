package main

import (
	"context"
	"flag"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"howdah/internal/pkg/common/utils"
	howdah_server "howdah/internal/pkg/howdah-server"

	// "howdah/internal/pkg/howdah-server"
	howdah_grpc "howdah/internal/pkg/howdah-server/grpc"
	"howdah/pb"
	"net"
	"net/http"
	"strconv"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	var (
		grpcPort *int
		gwPort *int
		listener net.Listener
	)

	gwPort = flag.Int("port", 9999, "")
	grpcPort = flag.Int("grpc_port", 9091, "")
	flag.Parse()

	listener, err := net.Listen("tcp", ":" + strconv.Itoa(*grpcPort))
	defer listener.Close()
	if err != nil {
		logger.Fatalf("Failed to listen on port %d.\n%s",
			*grpcPort)
	}

	server := grpc.NewServer()
	defer server.Stop()
	// Auth service.
	adminStore := howdah_grpc.NewMockAdminStore()
	authService, err := howdah_grpc.NewAuthService(adminStore)
	if err != nil {
		logger.Fatalf("Failed on building AuthService.")
	}
	pb.RegisterAuthenticationServer(server, authService)
	// Reception Service
	howdah_server.NewConcurrentQueue(
		goconcurrentqueue.NewFIFO())

	agentInfoStore := howdah_grpc.NewMockAgentStore()
	receptionist := howdah_grpc.NewMockReceptionist(agentInfoStore)
	heartbeatQueue := howdah_server.NewConcurrentQueue(
		goconcurrentqueue.NewFIFO())
	heartbeatProcessor := howdah_server.NewHeartbeatProcessor(heartbeatQueue)
	heartbeatHandler := howdah_server.NewHeartbeatHandler(heartbeatProcessor,
														utils.NewTimestamper())
	receptionService := howdah_grpc.NewHeartbeatReceptionServer(logger, &receptionist, heartbeatHandler)
	pb.RegisterHeartbeatReceptionServer(server, receptionService)
	// HowdahEvent Service
	eventHandler := howdah_grpc.NewHowdahEventHandler(goconcurrentqueue.NewFIFO())
	eventService := howdah_grpc.NewHowdahEventServer(eventHandler)
	pb.RegisterHowdahEventServer(server, eventService)



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

	err = pb.RegisterAuthenticationHandler(
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
