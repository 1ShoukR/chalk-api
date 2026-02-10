package repositories

import (
	"chalk-api/pkg/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// GetOrCreateConversation finds an existing conversation or creates one (idempotent)
func (r *MessageRepository) GetOrCreateConversation(ctx context.Context, coachID, clientID uint) (*models.Conversation, error) {
	var convo models.Conversation
	err := r.db.WithContext(ctx).
		Where("coach_id = ? AND client_id = ?", coachID, clientID).
		First(&convo).Error

	if err == gorm.ErrRecordNotFound {
		convo = models.Conversation{
			CoachID:  coachID,
			ClientID: clientID,
		}
		if err := r.db.WithContext(ctx).Create(&convo).Error; err != nil {
			return nil, err
		}
		return &convo, nil
	}
	if err != nil {
		return nil, err
	}
	return &convo, nil
}

// ListConversations returns all conversations for a user (as coach or client) sorted by most recent message
func (r *MessageRepository) ListConversations(ctx context.Context, userID uint) ([]models.Conversation, error) {
	var convos []models.Conversation

	// Find conversations where user is either the coach or the client
	err := r.db.WithContext(ctx).
		Preload("Coach.User.Profile").
		Preload("Client.User.Profile").
		Joins("LEFT JOIN coach_profiles ON coach_profiles.id = conversations.coach_id").
		Joins("LEFT JOIN client_profiles ON client_profiles.id = conversations.client_id").
		Where("coach_profiles.user_id = ? OR client_profiles.user_id = ?", userID, userID).
		Order("last_message_at DESC NULLS LAST").
		Find(&convos).Error

	return convos, err
}

func (r *MessageRepository) GetConversation(ctx context.Context, id uint) (*models.Conversation, error) {
	var convo models.Conversation
	err := r.db.WithContext(ctx).
		Preload("Coach.User.Profile").
		Preload("Client.User.Profile").
		First(&convo, id).Error
	if err != nil {
		return nil, err
	}
	return &convo, nil
}

// CreateMessage creates a message and updates the conversation's last_message_at in one transaction
func (r *MessageRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.CreateMessageTx(ctx, tx, message)
	})
}

// CreateMessageTx creates a message and updates conversation last_message_at within an existing transaction.
func (r *MessageRepository) CreateMessageTx(ctx context.Context, tx *gorm.DB, message *models.Message) error {
	if err := tx.WithContext(ctx).Create(message).Error; err != nil {
		return err
	}
	return tx.WithContext(ctx).
		Model(&models.Conversation{}).
		Where("id = ?", message.ConversationID).
		Update("last_message_at", message.CreatedAt).Error
}

func (r *MessageRepository) ListMessages(ctx context.Context, conversationID uint, limit, offset int) ([]models.Message, int64, error) {
	var messages []models.Message
	var total int64

	query := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID)

	if err := query.Model(&models.Message{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&messages).Error

	return messages, total, err
}

// MarkAsRead marks all unread messages in a conversation as read for the given user
func (r *MessageRepository) MarkAsRead(ctx context.Context, conversationID, senderID uint) error {
	now := time.Now()
	// Mark messages as read where the sender is NOT the current user (you read their messages)
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND read_at IS NULL", conversationID, senderID).
		Update("read_at", now).Error
}

// GetUnreadCount returns the number of unread messages across all conversations for a user
func (r *MessageRepository) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Joins("JOIN conversations ON conversations.id = messages.conversation_id").
		Joins("LEFT JOIN coach_profiles ON coach_profiles.id = conversations.coach_id").
		Joins("LEFT JOIN client_profiles ON client_profiles.id = conversations.client_id").
		Where("(coach_profiles.user_id = ? OR client_profiles.user_id = ?) AND messages.sender_id != ? AND messages.read_at IS NULL",
			userID, userID, userID).
		Count(&count).Error

	return count, err
}
