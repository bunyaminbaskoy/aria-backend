package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError — Özel hata tipi.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error — error arayüzü.
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError — Yeni hata oluşturur.
func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Hazır hata tanımları.
var (
	ErrUnauthorized    = NewAppError(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden       = NewAppError(http.StatusForbidden, "Forbidden")
	ErrNotFound        = NewAppError(http.StatusNotFound, "Resource not found")
	ErrBadRequest      = NewAppError(http.StatusBadRequest, "Bad request")
	ErrInternalServer  = NewAppError(http.StatusInternalServerError, "Internal server error")
	ErrConflict        = NewAppError(http.StatusConflict, "Resource already exists")
	ErrTooManyRequests = NewAppError(http.StatusTooManyRequests, "Too many requests")
)

// ErrorHandler — Hataları yakalar, JSON olarak döndürür.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// İsteği işle ve hata kontrolü yap
		c.Next()

		// Hata var mı?
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Bizim hata tipimiz mi?
			if appErr, ok := err.(*AppError); ok {
				c.JSON(appErr.Code, gin.H{
					"error": appErr.Message,
				})
				return
			}

			// Bilinmeyen hata — 500 döndür
			log.Printf("Unhandled error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
	}
}

// RecoveryHandler — Panic olursa sunucu çökmesin.
func RecoveryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}
