package auth

// SignupRequest represents the request body for user registration.
type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the authentication response with token and user info.
type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}
