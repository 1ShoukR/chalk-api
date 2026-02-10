package handlers

import (
	"chalk-api/pkg/services"
	"chalk-api/pkg/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CoachHandler struct {
	coachService *services.CoachService
}

func NewCoachHandler(coachService *services.CoachService) *CoachHandler {
	return &CoachHandler{coachService: coachService}
}

func (h *CoachHandler) GetMyProfile(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	profile, err := h.coachService.GetMyProfile(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch coach profile"})
		}
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *CoachHandler) UpsertMyProfile(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.UpsertCoachProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	profile, err := h.coachService.UpsertMyProfile(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save coach profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *CoachHandler) CreateInviteCode(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.CreateInviteCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		// Allow empty body and defaults.
		input = services.CreateInviteCodeInput{}
	}

	invite, err := h.coachService.CreateInviteCode(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invite code"})
		}
		return
	}

	c.JSON(http.StatusCreated, invite)
}

func (h *CoachHandler) ListInviteCodes(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	invites, err := h.coachService.ListInviteCodes(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list invite codes"})
		}
		return
	}

	c.JSON(http.StatusOK, invites)
}

func (h *CoachHandler) DeactivateInviteCode(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	inviteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || inviteID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite id"})
		return
	}

	if err := h.coachService.DeactivateInviteCode(c.Request.Context(), userID, uint(inviteID)); err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		case errors.Is(err, services.ErrInviteCodeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "invite code not found"})
		case errors.Is(err, services.ErrInviteForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "invite code does not belong to this coach"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate invite code"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invite code deactivated"})
}
