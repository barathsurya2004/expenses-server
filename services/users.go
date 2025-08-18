package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver

	pb "github.com/barathsurya2004/expenses/proto"
)

type usersServer struct {
	pb.UnimplementedUsersServiceServer
	db *sql.DB
}

type User struct {
	ID        string
	Username  string
	Email     string
	FirstName string
	LastName  string
	Password  string
}

func (s *usersServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := User{
		Username:  req.GetUsername(),
		Email:     req.GetEmail(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		Password:  req.GetPassword(),
	}
	uuid, err := uuid.NewV6()
	if err != nil {
		log.Fatalf("Failed to generate UUID: %v", err)
		return nil, err
	}
	user.ID = uuid.String()

	hashedPassword, err := passwordHash(user.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return nil, err
	}

	query := `INSERT INTO user_data (uuid, username, email, first_name, last_name, password_hash) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = s.db.ExecContext(ctx, query, user.ID, user.Username, user.Email, user.FirstName, user.LastName, hashedPassword)
	if err != nil {
		log.Fatalf("Failed to insert user: %v", err)
		return nil, err
	}
	//get the user ID from the database
	var userId string
	err = s.db.QueryRowContext(ctx, "SELECT uuid FROM user_data WHERE username = $1", user.Username).Scan(&userId)
	if err != nil {
		log.Fatalf("Failed to get user ID: %v", err)
		return nil, err
	}

	return &pb.CreateUserResponse{
		Message: "User created successfully",
		UserId:  userId,
	}, nil

}

func (s *usersServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {

	var user User

	query := `SELECT uuid,password_hash FROM user_data WHERE username = $1`
	if req.GetUsername() == "" {
		log.Printf("Username is required")
		return nil, fmt.Errorf("username is required")
	}
	err := s.db.QueryRowContext(ctx, query, req.GetUsername()).Scan(&user.ID, &user.Password)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		return nil, err
	}

	if req.GetPassword() != "" {
		if !passwordCheck(user.Password, req.GetPassword()) {
			log.Printf("Password check failed for user %s", user.Username)
			return nil, fmt.Errorf("invalid password for user %s", user.Username)
		}
		log.Printf("Password check successful for user %s", user.Username)
	}

	queryForCheckingAuthToken := `SELECT token, expires_at FROM token_data WHERE uuid = $1`
	var authToken string
	var expiresAt sql.NullTime
	err = s.db.QueryRowContext(ctx, queryForCheckingAuthToken, user.ID).Scan(&authToken, &expiresAt)
	if err == sql.ErrNoRows {
		log.Printf("No auth token found for user %s", user.Username)
		// Generate a new auth token if it doesn't exist
		authToken = uuid.New().String()           // Replace with actual token generation logic
		expires := time.Now().Add(24 * time.Hour) // Example: token valid for 24 hours
		_, err = s.db.ExecContext(ctx, "INSERT INTO token_data (uuid, token, expires_at,context) VALUES ($1, $2, $3,$4)", user.ID, authToken, expires, "user_auth")
		if err != nil {
			log.Printf("Failed to insert auth token: %v", err)
			return nil, err
		}
		log.Printf("Generated new auth token for user %s", user.Username)
	} else if err != nil {
		log.Printf("Failed to get auth token: %v", err)
		return nil, err
	} else if expiresAt.Valid && time.Now().Before(expiresAt.Time) {
		// Token is still valid, delete and create a new one
		_, err = s.db.ExecContext(ctx, "DELETE FROM token_data WHERE uuid = $1", user.ID)
		if err != nil {
			log.Printf("Failed to delete expired auth token: %v", err)
			return nil, err
		}
		authToken = uuid.New().String() // Replace with actual token generation logic
		expires := time.Now().Add(24 * time.Hour)
		_, err = s.db.ExecContext(ctx, "INSERT INTO token_data (uuid, token, expires_at,context) VALUES ($1, $2, $3,$4)", user.ID, authToken, expires, "user_auth")
		if err != nil {
			log.Printf("Failed to insert new auth token: %v", err)
			return nil, err
		}
		log.Printf("Deleted old token and generated new auth token for user %s", user.Username)
	}

	return &pb.GetUserResponse{
		UserId:    user.ID,
		AuthToken: authToken, // Replace with actual token generation logic
	}, nil

}

func passwordHash(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	return hash, nil

}

func passwordCheck(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Printf("Password check failed: %v", err)
		return false
	}
	return true
}
