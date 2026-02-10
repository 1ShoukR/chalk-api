package handlers

import (
	"chalk-api/pkg/services"
	"chalk-api/pkg/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InviteHandler struct {
	coachService *services.CoachService
}

func NewInviteHandler(coachService *services.CoachService) *InviteHandler {
	return &InviteHandler{coachService: coachService}
}

// GetPreview returns public invite metadata for deep-link flows before login/signup.
func (h *InviteHandler) GetPreview(c *gin.Context) {
	code := c.Param("code")
	preview, err := h.coachService.GetInvitePreview(c.Request.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInviteCodeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "invite code not found or expired"})
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch invite preview"})
		}
		return
	}

	c.JSON(http.StatusOK, preview)
}

func (h *InviteHandler) Accept(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.AcceptInviteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.coachService.AcceptInvite(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInviteCodeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "invite code not found or expired"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to accept invite"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}
