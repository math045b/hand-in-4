package main

import (
	"context"
	pb "hand-in-4/proto"
	"log"
	"net"

	"google.golang.org/grpc"
)

// protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/proto.proto
type Node struct {
	pb.UnimplementedServiceServer
	ID      string
	address string
	server  *grpc.Server
	clients map[string]pb.ServiceClient
}

func NewNode(id, address string) *Node {
	return &Node{
		ID:      id,
		address: address,
		server:  grpc.NewServer(),
		clients: make(map[string]pb.ServiceClient),
	}
}

func (n *Node) StartServer() {
	lis, err := net.Listen("tcp", n.address)
	if err != nil {
		log.Fatalf("Failed to listen on %v: %v", n.address, err)
	}

	pb.RegisterServiceServer(n.server, n)
	log.Printf("Node %s listening on %s\n", n.ID, n.address)
	if err := n.server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (n *Node) SendMessage(ctx context.Context, req *pb.Message) *pb.MessageResponse
