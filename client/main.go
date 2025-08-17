package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	pb "github.com/barathsurya2004/expenses/proto"
	"github.com/barathsurya2004/expenses/services/models"
)

const (
	addr      = "localhost:50051"
	imagePath = "reciept.jpg" // Path to the image you want to send
	chunkSize = 64 * 1024     // 64 KB chunk size
	port      = ":8080"       // Port for the HTTP server
)

type server struct {
	pb.UnimplementedExpensesServiceServer
	conn *grpc.ClientConn
	r    *mux.Router
}

func main() {
	server := &server{}
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	server.conn = conn

	r := mux.NewRouter()
	server.r = r

	r.Use(CorsMiddleWare)
	r.HandleFunc("/create-user", server.CreateUser).Methods("POST")
	r.HandleFunc("/get-user", server.GetUser).Methods("GET")
	r.HandleFunc("/create-expense", server.CreateExpense).Methods("POST")

	err = http.ListenAndServe(port, server.r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}

func CorsMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) GetUser(w http.ResponseWriter, r *http.Request) {
	pClient := pb.NewUsersServiceClient(s.conn)
	ctx := r.Context()

	var req models.GetUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	res, err := pClient.GetUser(ctx, &pb.GetUserRequest{
		UserId:   req.UserId,
		Password: req.Password,
	})
	if err != nil {
		log.Printf("Error getting user: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, `{"user_id": "%s", "username": "%s", "email": "%s", "first_name": "%s", "last_name": "%s"}`, res.GetUserId(), res.GetUsername(), res.GetEmail(), res.GetFirstName(), res.GetLastName())
}

func (s *server) CreateUser(w http.ResponseWriter, r *http.Request) {
	pClient := pb.NewUsersServiceClient(s.conn)
	ctx := r.Context()

	user := &models.Users{}

	json.NewDecoder(r.Body).Decode(&user)
	res, err := pClient.CreateUser(ctx, &pb.CreateUserRequest{
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Password:  user.Password,
	})
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "%s", "user_id": "%s"}`, res.GetMessage(), res.GetUserId())
	log.Printf("User created successfully: %s", res.GetMessage())

}

func (s *server) CreateExpense(w http.ResponseWriter, r *http.Request) {
	pClient := pb.NewExpensesServiceClient(s.conn)
	ctx := r.Context()

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	stream, err := pClient.CreateExpense(ctx)
	if err != nil {
		log.Printf("Error creating gRPC stream: %v", err)
		http.Error(w, "Failed to create gRPC stream", http.StatusInternalServerError)
		return
	}

	buffer := make([]byte, chunkSize)

	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading file: %v", err)
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		if err := stream.Send(&pb.CreateExpenseRequest{Chunks: buffer[:n]}); err != nil {
			log.Printf("Error sending chunk: %v", err)
			http.Error(w, "Failed to send file chunk", http.StatusInternalServerError)
			return
		}
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("Error receiving gRPC response: %v", err)
		http.Error(w, "Failed to receive response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, response.GetStatus())
}
