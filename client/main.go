package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"github.com/barathsurya2004/expenses/client/middleware"
	"github.com/barathsurya2004/expenses/client/routes"
)

const (
	addr = "localhost:50051"
	port = ":8080"
)

func main() {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	r := mux.NewRouter()
	routes.RegisterRoutes(r, conn)

	err = http.ListenAndServe(port, middleware.CorsMiddleWare(r))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
