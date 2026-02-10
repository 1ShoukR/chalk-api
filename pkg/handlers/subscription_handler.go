package handlers

import (
	"chalk-api/pkg/services"
	"chalk-api/pkg/utils"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscriptionService *services.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionService: subscriptionService}
}

func (h *SubscriptionHandler) RevenueCatWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
		return
	}

	if err := h.subscriptionService.HandleRevenueCatWebhook(
		c.Request.Context(),
		body,
		c.GetHeader("Authorization"),
	); err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidSubscriptionWebhookAuth):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook authorization"})
		case errors.Is(err, services.ErrSubscriptionWebhookPayload):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process subscription webhook"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *SubscriptionHandler) GetMySubscription(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subscription, err := h.subscriptionService.GetMySubscription(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch subscription"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

func (h *SubscriptionHandler) CheckFeatureAccess(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	feature := c.Param("feature")
	result, err := h.subscriptionService.CheckFeatureAccess(c.Request.Context(), userID, feature)
	if err != nil {
		if errors.Is(err, services.ErrFeatureNameRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "feature is required"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check feature access"})
		return
	}

	c.JSON(http.StatusOK, result)
}
