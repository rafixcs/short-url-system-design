package http

type LoginRequest struct {
	User     string `json:"user"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
