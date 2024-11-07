package main

import (
	"time"
)

func main() {
	nodeA := NewNode("A", "localhost:50051")
	go nodeA.StartServer()

	time.Sleep(time.Second)

	nodeB := NewNode("B", "localhost:50052")
	go nodeB.StartServer()
	time.Sleep(time.Second)

	nodeC := NewNode("C", "localhost:50053")
	go nodeC.StartServer()
	time.Sleep(time.Second)

	nodeA.ConnectToNode("B", "localhost:50052")
	nodeA.ConnectToNode("C", "localhost:50053")
	nodeB.ConnectToNode("A", "localhost:50051")
	nodeB.ConnectToNode("C", "localhost:50053")
	nodeC.ConnectToNode("A", "localhost:50051")
	nodeC.ConnectToNode("B", "localhost:50052")

	go nodeA.RequestResource()

	time.Sleep(time.Second)
	go nodeB.RequestResource()

	time.Sleep(2 * time.Second)
	go nodeC.RequestResource()

	select {}
}
