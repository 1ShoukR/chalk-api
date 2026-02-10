package services

import (
	"chalk-api/pkg/events"
	"chalk-api/pkg/models"
	"chalk-api/pkg/repositories"
	"context"
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrConversationNotFound   = errors.New("conversation not found")
	ErrConversationForbidden  = errors.New("conversation does not belong to this user")
	ErrMessageContentRequired = errors.New("message content or media is required")
	ErrClientProfileRequired  = errors.New("client profile id is required")
	ErrClientProfileInvalid   = errors.New("client profile does not belong to this user")
)

type CreateConversationInput struct {
	ClientProfileID uint `json:"client_profile_id" binding:"required"`
}

type SendMessageInput struct {
	Content   *string `json:"content"`
	MediaURL  *string `json:"media_url"`
	MediaType *string `json:"media_type"`
}

type MessageService struct {
	repos       *repositories.RepositoriesCollection
	messageRepo *repositories.MessageRepository
	clientRepo  *repositories.ClientRepository
	coachRepo   *repositories.CoachRepository
	events      *events.Publisher
}

func NewMessageService(
	repos *repositories.RepositoriesCollection,
	eventsPublisher *events.Publisher,
) *MessageService {
	return &MessageService{
		repos:       repos,
		messageRepo: repos.Message,
		clientRepo:  repos.Client,
		coachRepo:   repos.Coach,
		events:      eventsPublisher,
	}
}

func (s *MessageService) ListConversations(ctx context.Context, userID uint) ([]models.Conversation, error) {
	return s.messageRepo.ListConversations(ctx, userID)
}

func (s *MessageService) GetConversation(ctx context.Context, userID, conversationID uint) (*models.Conversation, error) {
	conversation, err := s.messageRepo.GetConversation(ctx, conversationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConversationNotFound
		}
		return nil, err
	}

	if !isConversationParticipant(userID, conversation) {
		return nil, ErrConversationForbidden
	}

	return conversation, nil
}

// GetOrCreateConversationByClientProfile resolves coach/client relationship authorization
// and then gets/creates a single conversation for that pair.
func (s *MessageService) GetOrCreateConversationByClientProfile(ctx context.Context, userID uint, input CreateConversationInput) (*models.Conversation, error) {
	if input.ClientProfileID == 0 {
		return nil, ErrClientProfileRequired
	}

	clientProfile, err := s.clientRepo.GetByID(ctx, input.ClientProfileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClientProfileInvalid
		}
		return nil, err
	}

	coachProfile, err := s.coachRepo.GetByID(ctx, clientProfile.CoachID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClientProfileInvalid
		}
		return nil, err
	}

	// Allow either side to initiate: coach user or client user in this relationship.
	if userID != clientProfile.UserID && userID != coachProfile.UserID {
		return nil, ErrClientProfileInvalid
	}

	conversation, err := s.messageRepo.GetOrCreateConversation(ctx, coachProfile.ID, clientProfile.ID)
	if err != nil {
		return nil, err
	}

	return s.messageRepo.GetConversation(ctx, conversation.ID)
}

func (s *MessageService) ListMessages(ctx context.Context, userID, conversationID uint, limit, offset int) ([]models.Message, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	if _, err := s.GetConversation(ctx, userID, conversationID); err != nil {
		return nil, 0, err
	}

	return s.messageRepo.ListMessages(ctx, conversationID, limit, offset)
}

func (s *MessageService) SendMessage(ctx context.Context, userID, conversationID uint, input SendMessageInput) (*models.Message, error) {
	content := trimPtr(input.Content)
	mediaURL := trimPtr(input.MediaURL)

	if content == nil && mediaURL == nil {
		return nil, ErrMessageContentRequired
	}

	conversation, err := s.GetConversation(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	recipientID := resolveRecipientUserID(userID, conversation)
	if recipientID == 0 {
		return nil, ErrConversationForbidden
	}

	message := &models.Message{
		ConversationID: conversationID,
		SenderID:       userID,
		Content:        content,
		MediaURL:       mediaURL,
		MediaType:      trimPtr(input.MediaType),
	}

	if err := s.repos.WithTransaction(ctx, func(tx *gorm.DB, txRepos *repositories.RepositoriesCollection) error {
		if err := txRepos.Message.CreateMessageTx(ctx, tx, message); err != nil {
			return err
		}

		if s.events != nil {
			payload := events.MessageSentPayload{
				MessageID:      message.ID,
				ConversationID: message.ConversationID,
				SenderID:       message.SenderID,
				RecipientID:    recipientID,
				ContentPreview: buildMessagePreview(content),
			}
			idempotencyKey := events.BuildIdempotencyKey(
				events.EventTypeMessageSent,
				strconv.FormatUint(uint64(message.ID), 10),
			)
			if err := s.events.PublishInTx(
				ctx,
				tx,
				events.EventTypeMessageSent,
				"message",
				strconv.FormatUint(uint64(message.ID), 10),
				idempotencyKey,
				payload,
			); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *MessageService) MarkAsRead(ctx context.Context, userID, conversationID uint) error {
	if _, err := s.GetConversation(ctx, userID, conversationID); err != nil {
		return err
	}
	return s.messageRepo.MarkAsRead(ctx, conversationID, userID)
}

func (s *MessageService) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.messageRepo.GetUnreadCount(ctx, userID)
}

func isConversationParticipant(userID uint, conversation *models.Conversation) bool {
	return conversation.Coach.UserID == userID || conversation.Client.UserID == userID
}

func resolveRecipientUserID(senderID uint, conversation *models.Conversation) uint {
	if conversation.Coach.UserID == senderID {
		return conversation.Client.UserID
	}
	if conversation.Client.UserID == senderID {
		return conversation.Coach.UserID
	}
	return 0
}

func trimPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func buildMessagePreview(content *string) *string {
	cleaned := trimPtr(content)
	if cleaned == nil {
		return nil
	}

	const maxLen = 120
	runes := []rune(*cleaned)
	if len(runes) <= maxLen {
		return cleaned
	}

	preview := string(runes[:maxLen]) + "..."
	return &preview
}
