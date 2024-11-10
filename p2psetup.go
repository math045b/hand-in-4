package main

import (
	"context"
	"fmt"
	pb "hand-in-4/proto"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/proto.proto
type Node struct {
	pb.UnimplementedServiceServer
	ID           string
	address      string
	server       *grpc.Server
	client       pb.ServiceClient
	mutex        sync.Mutex
	timestamp    int64
	hasToken     bool
	requestQueue []pb.ServiceClient
}

func NewNode(id, address string) *Node {
	return &Node{
		ID:           id,
		address:      address,
		server:       grpc.NewServer(),
		requestQueue: make([]pb.ServiceClient, 0),
	}
}

func (n *Node) loop(hasToken bool) {
	n.hasToken = hasToken
	for {
		time.Sleep(1 * time.Second)
		if n.hasToken {
			n.UseResource()
			n.hasToken = false

			n.timestamp++
			req := &pb.GrantTokenRequest{
				SenderId:  n.ID,
				Timestamp: n.timestamp,
			}
			response, err := n.client.GrantToken(context.Background(), req)
			if err != nil || response.Success == false {
				n.hasToken = true
			}
			n.timestamp = max(response.Timestamp, n.timestamp)
		}
	}
}

func (n *Node) GrantToken(ctx context.Context, req *pb.GrantTokenRequest) (*pb.GrantTokenResponse, error) {
	n.timestamp = max(req.Timestamp, n.timestamp)
	n.timestamp++
	log.Printf("(%d) Node %s received token from %s\n", n.timestamp, n.ID, req.SenderId)
	n.hasToken = true
	return &pb.GrantTokenResponse{Success: true, Timestamp: n.timestamp}, nil
}

func (n *Node) UseResource() {
	log.Printf("(%d) Node %s is using the resource \n", n.timestamp, n.ID)
	time.Sleep(2 * time.Second)
	n.timestamp++
	log.Printf("(%d) Node %s is finished with the resource", n.timestamp, n.ID)
}

func (n *Node) ConnectToNode(nodeID, address string) error {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %v", nodeID, err)
	}

	client := pb.NewServiceClient(conn)
	n.client = client
	log.Printf("Node %s connected to node %s at %s\n", n.ID, nodeID, address)
	return nil
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
	n.timestamp = 0
}
