package main

import (
	"google.golang.org/grpc"
	"hand-in-4/proto"
	"log"
	"net"
)

func createAndStartServer() {
	server := &Server{
		serverLamport: 0,
	}
	server.start()
}

type Server struct {
	proto.UnimplementedServiceServer
	serverLamport int32
}

func (cs *Server) start() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterServiceServer(grpcServer, cs)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}
}
