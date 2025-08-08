package main

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"google.golang.org/genai"

	// Replace with your actual proto package import
	pb "github.com/barathsurya2004/expenses/proto"
)

const (
	port = ":50051"
)

// server implements the gRPC ExpensesServiceServer interface.
type expenseServer struct {
	pb.UnimplementedExpensesServiceServer
}

// CreateExpense is a client-streaming RPC that receives an image and processes it.
func (s *expenseServer) CreateExpense(stream pb.ExpensesService_CreateExpenseServer) error {
	log.Println("Receiving image chunks from client...")

	var imageBytes []byte

	// Read image chunks from the client stream.
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break // End of stream
		}
		if err != nil {
			log.Printf("Error receiving chunk: %v", err)
			return err
		}

		imageBytes = append(imageBytes, chunk.Chunks...)
	}
	log.Println("Image chunks received and saved successfully.")

	prompt := `
	You are a helpful assistant. Extract the following details from the receipt image and return them as a JSON object.
		Do not include any extra text before or after the JSON.

		Fields to extract:
		- "transaction_id": A unique identifier for the receipt.
		- "merchant_details":
		- "name": The name of the store or service.
		- "transaction_details":
		- "date": The transaction date in "YYYY-MM-DD" format.
		- "time": The transaction time in "HH:MM:SS" format.
		- "payment_method": The payment method used (e.g., "Credit Card", "Cash", "Debit Card").
		- "total_amount": The total amount spent, as a float.
		- "currency": The currency of the total amount (e.g., "USD", "EUR").
		- "items": An array of objects, where each object has:
		- "item_name": The name of the product or service.
		- "price": The item's price as a float.
		- "quantity": The number of units purchased.
		- "category": A classification of the item (e.g., "Groceries", "Household", "Dining").
		- "spending_category": A top-level classification for the entire receipt (e.g., "Groceries", "Dining Out", "Utilities").

		if there are some data missing from the receipt, you can leave them empty.
		Return only the JSON object.
	`

	responseText, err := genAI(imageBytes, prompt)
	if err != nil {
		return err
	}

	log.Println("Successfully processed image with GenAI. Sending response to client.")

	// Send the final JSON response back to the client and close the stream.
	return stream.SendAndClose(&pb.CreateExpenseResponse{Status: responseText})
}

func genAI(image []byte, prompt string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API"),
	})
	if err != nil {
		return "", err
	}

	parts := []*genai.Part{
		genai.NewPartFromText(strings.TrimSpace(prompt)),
		genai.NewPartFromBytes(image, "image/jpeg"),
	}
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		contents,
		nil,
	)
	if err != nil {
		return "", err
	}

	return result.Text(), nil
}
