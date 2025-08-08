package main

import (
	"database/sql"
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
	dbConn, err := OpenConnection(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()
	s := grpc.NewServer()
	pb.RegisterExpensesServiceServer(s, &expenseServer{})
	pb.RegisterUsersServiceServer(s, &usersServer{
		db: dbConn,
	})

	log.Println("Server is running on port ", port)
	if err := s.Serve(conn); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}

func OpenConnection(conn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
