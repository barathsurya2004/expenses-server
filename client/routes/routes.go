package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"github.com/barathsurya2004/expenses/client/middleware"
	pb "github.com/barathsurya2004/expenses/proto"
	"github.com/barathsurya2004/expenses/services/models"
)

type Server struct {
	Conn *grpc.ClientConn
}

func RegisterRoutes(r *mux.Router, conn *grpc.ClientConn) {
	server := &Server{Conn: conn}
	r.HandleFunc("/create-user", server.CreateUser).Methods("POST")
	r.HandleFunc("/get-user", server.GetUser).Methods("GET")
	r.Handle("/create-expense", middleware.AuthorizationMiddleware(http.HandlerFunc(server.CreateExpense))).Methods("POST")
}

func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	pClient := pb.NewUsersServiceClient(s.Conn)
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

	fmt.Fprintf(w, `{"user_id": "%s", "username": "%s", "email": "%s", "first_name": "%s", "last_name": "%s"}`,
		res.GetUserId(), res.GetUsername(), res.GetEmail(), res.GetFirstName(), res.GetLastName())
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	pClient := pb.NewUsersServiceClient(s.Conn)
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
	fmt.Fprintf(w, `{"message": "%s", "user_id": "%s"}`,
		res.GetMessage(), res.GetUserId())
	log.Printf("User created successfully: %s", res.GetMessage())
}

func (s *Server) CreateExpense(w http.ResponseWriter, r *http.Request) {
	// limit total upload size to 50 MB (adjust as needed)
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}

	// retrieve file under the key "file"
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("File received: filename=%q, header=%+v", handler.Filename, handler.Header)

	pClient := pb.NewExpensesServiceClient(s.Conn)
	ctx := r.Context()
	stream, err := pClient.CreateExpense(ctx)
	if err != nil {
		log.Printf("Error creating gRPC stream: %v", err)
		http.Error(w, "Failed to create gRPC stream", http.StatusInternalServerError)
		return
	}

	buffer := make([]byte, 64*1024)
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
