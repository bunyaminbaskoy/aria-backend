package auth

// SignupRequest — Kayıt isteği.
type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest — Giriş isteği.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse — Giriş yanıtı.
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
}

// RefreshRequest — Token yenileme isteği.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest — Çıkış isteği.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
