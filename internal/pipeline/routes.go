package pipeline

import (
	"github.com/gin-gonic/gin"

	"music-curation/internal/middleware"
)

// RegisterRoutes, orchestrator endpoint'lerini verilen router grubuna
// kaydeder. AuthMiddleware her zaman bağlanır; ek middleware'ler
// (örn. rate limiter) `extra` ile dışarıdan enjekte edilir. Bu sayede
// pipeline modülü Redis client'ı gibi altyapı bileşenlerine doğrudan
// bağımlı olmaz — modüler monolith disiplinine uygun.
//
// Kayıtlanan route'lar:
//
//	POST /api/v1/recommendations/generate
//	     → AuthMiddleware + extra... + handler.Generate
//
// Not: /recommendations grubu altında yer alıyor olması frontend için
// tutarlılık sağlar; "/generate" static segment'i, recommendation
// modülündeki "/recommendations/:id" param route'undan Gin'in radix
// tree önceliği nedeniyle daha öncelikli eşleşir — çakışma yoktur.
func RegisterRoutes(router *gin.RouterGroup, handler *Handler, extra ...gin.HandlerFunc) {
	g := router.Group("/recommendations")
	g.Use(middleware.AuthMiddleware())
	for _, mw := range extra {
		g.Use(mw)
	}
	g.POST("/generate", handler.Generate)
}
