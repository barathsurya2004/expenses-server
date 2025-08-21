package models

import "time"

type Users struct {
	UUID      string `json:"uuid"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

type GetUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// The top-level struct to hold the entire JSON object
type Transaction struct {
	UUID               string            `json:"uuid"`
	TransactionID      string            `json:"transaction_id"`
	MerchantDetails    Merchant          `json:"merchant_details"`
	TransactionDetails TransactionDetail `json:"transaction_details"`
	Items              []Item            `json:"items"`
	SpendingCategory   string            `json:"spending_category"`
}

// A nested struct to handle the "merchant_details" object
type Merchant struct {
	Name string `json:"name"`
}

// A nested struct to handle the "transaction_details" object
// Note: It's a good practice to use time.Time for timestamps to leverage Go's
// built-in time handling and to be compatible with database drivers.
type TransactionDetail struct {
	DateTime      time.Time `json:"date_and_time"`
	PaymentMethod string    `json:"payment_method"`
	TotalAmount   float64   `json:"total_amount"`
	Currency      string    `json:"currency"`
}

// A nested struct to handle each object within the "items" array
type Item struct {
	ItemName string  `json:"item_name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Category string  `json:"category"`
}

type HeatMapDay struct {
	Day      string `json:"day"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}
