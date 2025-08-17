package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"google.golang.org/genai"

	// Replace with your actual proto package import
	pb "github.com/barathsurya2004/expenses/proto"
	"github.com/barathsurya2004/expenses/services/models"
)

const (
	port = ":50051"
)

// server implements the gRPC ExpensesServiceServer interface.
type expenseServer struct {
	pb.UnimplementedExpensesServiceServer
	db *sql.DB
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
		- "date_and_time" : The date and time of the transaction in ISO 8601 format (e.g., "2023-10-01T12:00:00Z").
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

	jsonParsed := strings.Replace(responseText, "```json\n", "", -1)
	jsonParsed = strings.Replace(jsonParsed, "```", "", -1)
	responseText = strings.TrimSpace(jsonParsed)

	// Parse the JSON response into the Expenses model.
	var expense models.Transaction
	err = json.Unmarshal([]byte(responseText), &expense)
	if err != nil {
		log.Printf("Error parsing JSON response: %v", err)
	}

	// Write the expense data to the database.
	err = s.WriteExpenseToDB(expense)
	if err != nil {
		log.Printf("Error writing expense to database: %v", err)
	}

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

func (s *expenseServer) WriteExpenseToDB(expense models.Transaction) error {

	query := `INSERT INTO expense_data (uuid,date_and_time, place, mode_of_payment, amount, currency, category) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	uuid := "01f07b63-fa0f-6ed9-b910-00155d4c459b"

	_, err := s.db.ExecContext(context.Background(), query,
		uuid,
		expense.TransactionDetails.DateTime,
		expense.MerchantDetails.Name,
		expense.TransactionDetails.PaymentMethod,
		expense.TransactionDetails.TotalAmount,
		expense.TransactionDetails.Currency,
		expense.SpendingCategory,
	)

	return err

}
