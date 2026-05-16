package stats

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"music-curation/internal/middleware"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	s := router.Group("/stats")
	s.Use(middleware.AuthMiddleware())
	{
		s.GET("", handler.GetStats)
	}
}

type MoodCount struct {
	DominantEmotion string `json:"dominant_emotion"`
	Count           int    `json:"count"`
}

type WeekDay struct {
	Day             string  `json:"day"`
	Count           int     `json:"count"`
	DominantEmotion string  `json:"dominant_emotion"`
	AvgEnergy       float64 `json:"avg_energy"`
}

func (h *Handler) GetStats(c *gin.Context) {
	raw, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme gerekli"})
		return
	}
	userID := raw.(uint)

	// Total analysis count
	var analysisCount int64
	h.db.Table("moods").Where("user_id = ? AND status = 'analyzed'", userID).Count(&analysisCount)

	// Playlist count
	var playlistCount int64
	h.db.Table("playlists").Where("user_id = ?", userID).Count(&playlistCount)

	// Saved track count
	var savedCount int64
	h.db.Table("saved_tracks").Where("user_id = ?", userID).Count(&savedCount)

	// Unique moods discovered
	var uniqueMoods int64
	h.db.Table("moods").Where("user_id = ? AND status = 'analyzed' AND dominant_emotion != ''", userID).
		Distinct("dominant_emotion").Count(&uniqueMoods)

	// Top moods distribution
	var moodCounts []MoodCount
	h.db.Table("moods").
		Select("dominant_emotion, COUNT(*) as count").
		Where("user_id = ? AND status = 'analyzed' AND dominant_emotion != ''", userID).
		Group("dominant_emotion").
		Order("count DESC").
		Limit(6).
		Scan(&moodCounts)

	// Weekly mood chart (last 7 days)
	weekAgo := time.Now().AddDate(0, 0, -7)
	var weeklyRaw []struct {
		DayNum          int     `json:"day_num"`
		Count           int     `json:"count"`
		DominantEmotion string  `json:"dominant_emotion"`
		AvgEnergy       float64 `json:"avg_energy"`
	}
	h.db.Table("moods").
		Select("EXTRACT(DOW FROM created_at)::int as day_num, COUNT(*) as count, MODE() WITHIN GROUP (ORDER BY dominant_emotion) as dominant_emotion, AVG(energy) as avg_energy").
		Where("user_id = ? AND status = 'analyzed' AND created_at >= ?", userID, weekAgo).
		Group("day_num").
		Order("day_num").
		Scan(&weeklyRaw)

	dayNames := []string{"Paz", "Pzt", "Sal", "Çar", "Per", "Cum", "Cmt"}
	weekly := make([]WeekDay, 7)
	for i := 0; i < 7; i++ {
		weekly[i] = WeekDay{Day: dayNames[i]}
	}
	for _, r := range weeklyRaw {
		if r.DayNum >= 0 && r.DayNum < 7 {
			weekly[r.DayNum] = WeekDay{
				Day:             dayNames[r.DayNum],
				Count:           r.Count,
				DominantEmotion: r.DominantEmotion,
				AvgEnergy:       r.AvgEnergy,
			}
		}
	}

	// Recent mood history (last 10)
	type MoodHistory struct {
		RawText         string    `json:"raw_text"`
		DominantEmotion string    `json:"dominant_emotion"`
		SentimentLabel  string    `json:"sentiment_label"`
		Energy          float64   `json:"energy"`
		Valence         float64   `json:"valence"`
		CreatedAt       time.Time `json:"created_at"`
	}
	var history []MoodHistory
	h.db.Table("moods").
		Select("raw_text, dominant_emotion, sentiment_label, energy, valence, created_at").
		Where("user_id = ? AND status = 'analyzed'", userID).
		Order("created_at DESC").
		Limit(10).
		Scan(&history)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"analysis_count": analysisCount,
			"playlist_count": playlistCount,
			"saved_count":    savedCount,
			"unique_moods":   uniqueMoods,
			"top_moods":      moodCounts,
			"weekly":         weekly,
			"history":        history,
		},
	})
}
