package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	pb "github.com/barathsurya2004/expenses/proto"
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

type User struct {
	UUID      string `json:"uuid"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

func (s *server) GetUser(w http.ResponseWriter, r *http.Request) {
	pClient := pb.NewUsersServiceClient(s.conn)
	ctx := r.Context()

	type request struct {
		UserId   string `json:"user_id"`
		Password string `json:"password"`
	}

	var req request
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

	// Here you would extract user details from the request body.
	// For simplicity, we are using hardcoded values.
	user := &User{}

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
	pClient := pb.NewUsersServiceClient(s.conn)
	ctx := r.Context()

	_, err := pClient.GetUser(ctx, &pb.GetUserRequest{
		UserId: "some-uuid",
	})
	if err != nil {
		log.Printf("Error getting user: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Note: The above is
	// Here you would implement the logic to handle the expense creation.
	// For example, you could read the image file and send it in chunks.
	// This is just a placeholder implementation.
}

// func main() {
// 	// Connect to the gRPC server.
// 	conn, err = grpc.Dial(addr, grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("Did not connect: %v", err)
// 	}
// 	defer conn.Close()

// 	c := pb.NewExpensesServiceClient(conn)

// 	// Open the image file.
// 	//print the current working directory
// 	file, err := os.Open(imagePath)
// 	if err != nil {
// 		log.Fatalf("Failed to open image file: %v", err)
// 	}
// 	defer file.Close()

// 	// Create a new context with a timeout for the RPC.
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	// res, err := c1.CreateUser(ctx, &pb.CreateUserRequest{
// 	// 	Username:  "johndoe",
// 	// 	Email:     "johndoe@email.com",
// 	// 	FirstName: "John",
// 	// 	LastName:  "Doe",
// 	// 	Password:  "password123",
// 	// })

// 	// if err != nil {
// 	// 	log.Fatalf("Error creating user: %v", err)
// 	// }
// 	// fmt.Println("User created successfully:", res.GetMessage(), "User ID:", res.GetUserId())

// 	// Call the client-streaming RPC.
// 	stream, err := c.CreateExpense(ctx)

// 	if err != nil {
// 		log.Fatalf("Error creating stream: %v", err)
// 	}

// 	reader := bufio.NewReader(file)
// 	buffer := make([]byte, chunkSize)

// 	// Read the file in chunks and stream them to the server.
// 	for {
// 		n, err := reader.Read(buffer)
// 		if err == io.EOF {
// 			break // End of file
// 		}
// 		if err != nil {
// 			log.Fatalf("Error reading file chunk: %v", err)
// 		}

// 		// Send the chunk to the server.
// 		if err := stream.Send(&pb.CreateExpenseRequest{Chunks: buffer[:n]}); err != nil {
// 			log.Fatalf("Error sending chunk: %v", err)
// 		}
// 	}

// 	// Close the stream and wait for the final response from the server.
// 	response, err := stream.CloseAndRecv()
// 	if err != nil {
// 		log.Fatalf("Error receiving response from server: %v", err)
// 	}

// 	// Print the JSON response.
// 	fmt.Println("Server Response:")
// 	fmt.Println(response.GetStatus())
// }
