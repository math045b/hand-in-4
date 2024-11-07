package main

import (
	"context"
	"fmt"
	pb "hand-in-4/proto"
	"log"
	"net"
	"os"
	"strconv"
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
	clients      map[string]pb.ServiceClient
	mutex        sync.Mutex
	timestamp    int64
	requestQueue map[string]int64
}

var usingResource bool

func NewNode(id, address string) *Node {
	return &Node{
		ID:           id,
		address:      address,
		server:       grpc.NewServer(),
		clients:      make(map[string]pb.ServiceClient),
		requestQueue: make(map[string]int64),
	}
}

func main() {
	myId := os.Args[2]
	myPort := os.Args[3]
	totalClients, _ := strconv.Atoi(os.Args[1])

	node := NewNode(myId, myPort)
	usingResource = false
	go node.StartServer()

	if len(os.Args) > 4 {
		joinPort := os.Args[4]
		node.ConnectToNode(os.Args[2], joinPort)
	}
}

func (n *Node) RequestAccess(ctx context.Context, req *pb.AccessRequest) (*pb.AccessResponse, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.timestamp = max(n.timestamp, req.Timestamp) + 1

	if !usingResource && (len(n.requestQueue) == 0 || req.Timestamp < n.timestamp) {
		return &pb.AccessResponse{Granted: true}, nil
	}

	n.requestQueue[req.SenderId] = req.Timestamp
	return &pb.AccessResponse{Granted: false}, nil
}

func (n *Node) GrantAccess(ctx context.Context, req *pb.GrantRequest) (*pb.GrantResponse, error) {
	log.Printf("Node %s received grant from %s\n", n.ID, req.SenderId)
	return &pb.GrantResponse{Success: true}, nil
}

func (n *Node) RequestResource() {
	n.mutex.Lock()
	n.timestamp++
	timestamp := n.timestamp
	n.mutex.Unlock()

	for _, client := range n.clients {
		req := &pb.AccessRequest{
			SenderId:  n.ID,
			Timestamp: timestamp,
		}
		response, err := client.RequestAccess(context.Background(), req)
		log.Printf("Node %s recieved answer: %t\n", n.ID, response.Granted)
		if err != nil || !response.Granted {
			log.Printf("Node %s did not receive access from one or more nodes\n", n.ID)
			return
		}
	}

	n.UseResource()
}

func (n *Node) UseResource() {
	log.Printf("Node %s is using the resource \n", n.ID)
	time.Sleep(2 * time.Second)

	n.mutex.Lock()
	//n.usingResource = false
	n.mutex.Unlock()
	n.ProcessQueue()
	log.Printf("Node %s is finished with the resource", n.ID)
}

func (n *Node) ProcessQueue() {
	for requester, _ := range n.requestQueue {
		client, exists := n.clients[requester]
		if exists {
			req := &pb.GrantRequest{SenderId: n.ID}
			client.GrantAccess(context.Background(), req)
			delete(n.requestQueue, requester)
		}
	}
}

func (n *Node) ConnectToNode(nodeID, address string) error {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %v", nodeID, err)
	}

	client := pb.NewServiceClient(conn)
	n.clients[nodeID] = client
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
}

func max(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}