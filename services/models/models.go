package models

type Users struct {
	UUID      string `json:"uuid"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

type GetUserRequest struct {
	UserId   string `json:"user_id"`
	Password string `json:"password"`
}
