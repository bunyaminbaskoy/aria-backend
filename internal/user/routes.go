package user

import (
	"github.com/gin-gonic/gin"

	"music-curation/internal/middleware"
)

// RegisterRoutes registers all user routes to the given router group.
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	users := router.Group("/users")
	users.Use(middleware.AuthMiddleware())
	{
		users.GET("", handler.GetAll)
		users.GET("/:id", handler.GetByID)
	}
}
