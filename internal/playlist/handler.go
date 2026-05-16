package playlist

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var ErrNotOwner = errors.New("playlist sahibi değilsiniz")

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func userID(c *gin.Context) (uint, bool) {
	raw, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	id, ok := raw.(uint)
	return id, ok
}

type CreateRequest struct {
	Name   string       `json:"name" binding:"required,max=255"`
	Tracks []TrackInput `json:"tracks" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme gerekli"})
		return
	}
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.service.Create(uid, req.Name, req.Tracks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Playlist oluşturulamadı"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": p})
}

func (h *Handler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme gerekli"})
		return
	}
	playlists, err := h.service.GetUserPlaylists(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Playlistler yüklenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": playlists})
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz ID"})
		return
	}
	p, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist bulunamadı"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

type RenameRequest struct {
	Name string `json:"name" binding:"required,max=255"`
}

func (h *Handler) Rename(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme gerekli"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz ID"})
		return
	}
	var req RenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.service.Rename(uint(id), uid, req.Name)
	if err != nil {
		if errors.Is(err, ErrNotOwner) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Playlist güncellenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h *Handler) Delete(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme gerekli"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz ID"})
		return
	}
	if err := h.service.Delete(uint(id), uid); err != nil {
		if errors.Is(err, ErrNotOwner) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Playlist silinemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Playlist silindi"})
}

func (h *Handler) RemoveTrack(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme gerekli"})
		return
	}
	playlistID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz playlist ID"})
		return
	}
	trackID, err := strconv.ParseUint(c.Param("trackId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz track ID"})
		return
	}
	if err := h.service.RemoveTrack(uint(playlistID), uint(trackID), uid); err != nil {
		if errors.Is(err, ErrNotOwner) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Şarkı kaldırılamadı"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Şarkı kaldırıldı"})
}
