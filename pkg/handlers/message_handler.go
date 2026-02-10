package handlers

import (
	"chalk-api/pkg/services"
	"chalk-api/pkg/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageService *services.MessageService
}

func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) ListConversations(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversations, err := h.messageService.ListConversations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": conversations})
}

func (h *MessageHandler) GetOrCreateConversation(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.CreateConversationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	conversation, err := h.messageService.GetOrCreateConversationByClientProfile(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrClientProfileRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "client_profile_id is required"})
		case errors.Is(err, services.ErrClientProfileInvalid):
			c.JSON(http.StatusForbidden, gin.H{"error": "client profile does not belong to this user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get or create conversation"})
		}
		return
	}

	c.JSON(http.StatusOK, conversation)
}

func (h *MessageHandler) GetConversation(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	conversation, err := h.messageService.GetConversation(c.Request.Context(), userID, conversationID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrConversationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		case errors.Is(err, services.ErrConversationForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "conversation does not belong to this user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch conversation"})
		}
		return
	}

	c.JSON(http.StatusOK, conversation)
}

func (h *MessageHandler) ListMessages(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	limit := parseQueryInt(c.DefaultQuery("limit", "20"), 20)
	offset := parseQueryInt(c.DefaultQuery("offset", "0"), 0)

	messages, total, err := h.messageService.ListMessages(c.Request.Context(), userID, conversationID, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrConversationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		case errors.Is(err, services.ErrConversationForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "conversation does not belong to this user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list messages"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   messages,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	var input services.SendMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	message, err := h.messageService.SendMessage(c.Request.Context(), userID, conversationID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrConversationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		case errors.Is(err, services.ErrConversationForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "conversation does not belong to this user"})
		case errors.Is(err, services.ErrMessageContentRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "content or media_url is required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
		}
		return
	}

	c.JSON(http.StatusCreated, message)
}

func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	if err := h.messageService.MarkAsRead(c.Request.Context(), userID, conversationID); err != nil {
		switch {
		case errors.Is(err, services.ErrConversationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		case errors.Is(err, services.ErrConversationForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "conversation does not belong to this user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark conversation as read"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation marked as read"})
}

func (h *MessageHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := h.messageService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}
