package main

import (
	"time"
)

func main() {
	nodeA := NewNode("A", "localhost:5051")
	go nodeA.StartServer()
	time.Sleep(time.Second)

	nodeB := NewNode("B", "localhost:5052")
	go nodeB.StartServer()
	time.Sleep(time.Second)

	nodeC := NewNode("C", "localhost:5053")
	go nodeC.StartServer()
	time.Sleep(time.Second)

	nodeA.ConnectToNode("B", "localhost:5052")
	nodeB.ConnectToNode("C", "localhost:5053")
	nodeC.ConnectToNode("A", "localhost:5051")

	go nodeA.loop(true)
	go nodeB.loop(false)
	go nodeC.loop(false)

	select {}
}
