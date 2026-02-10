package handlers

import (
	"chalk-api/pkg/services"
	"chalk-api/pkg/utils"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input services.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.authService.Register(c.Request.Context(), input, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), input, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		case errors.Is(err, services.ErrUserDisabled):
			c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var input services.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), input, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidRefresh):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		case errors.Is(err, services.ErrUserDisabled):
			c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh token"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.LogoutInput
	// Allow empty JSON body for default logout behavior.
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), userID, input); err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidRefresh):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
