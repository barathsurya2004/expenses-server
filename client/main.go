package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	// Replace with your actual proto package import
	pb "github.com/barathsurya2004/expenses/proto"
)

const (
	addr      = "localhost:50051"
	imagePath = "reciept.jpg" // Path to the image you want to send
	chunkSize = 64 * 1024     // 64 KB chunk size
)

func main() {
	// Connect to the gRPC server.
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	// Create a new gRPC client.
	c := pb.NewExpensesServiceClient(conn)

	// Open the image file.
	//print the current working directory
	fmt.Println("Current working directory:", os.Getenv("PWD"))
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("Failed to open image file: %v", err)
	}
	defer file.Close()

	// Create a new context with a timeout for the RPC.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call the client-streaming RPC.
	stream, err := c.CreateExpense(ctx)
	if err != nil {
		log.Fatalf("Error creating stream: %v", err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, chunkSize)

	// Read the file in chunks and stream them to the server.
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break // End of file
		}
		if err != nil {
			log.Fatalf("Error reading file chunk: %v", err)
		}

		// Send the chunk to the server.
		if err := stream.Send(&pb.CreateExpenseRequest{Chunks: buffer[:n]}); err != nil {
			log.Fatalf("Error sending chunk: %v", err)
		}
	}

	// Close the stream and wait for the final response from the server.
	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error receiving response from server: %v", err)
	}

	// Print the JSON response.
	fmt.Println("Server Response:")
	fmt.Println(response.GetStatus())
}
