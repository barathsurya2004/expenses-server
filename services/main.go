package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/barathsurya2004/expenses/proto"
)

func main() {
	conn, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	s := grpc.NewServer()
	pb.RegisterExpensesServiceServer(s, &server{})

	log.Println("Server is running on port ", port)
	if err := s.Serve(conn); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
