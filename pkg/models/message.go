package models

import "time"

// Conversation - One conversation per coach-client pair.
// Dedicated table enables fast inbox listing without scanning all messages.
type Conversation struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	CoachID  uint `gorm:"index;not null" json:"coach_id"`
	ClientID uint `gorm:"index;not null" json:"client_id"`

	LastMessageAt *time.Time `gorm:"index" json:"last_message_at"` // for sorting inbox by most recent

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Coach    CoachProfile  `gorm:"foreignKey:CoachID" json:"coach,omitempty"`
	Client   ClientProfile `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	Messages []Message     `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

func (Conversation) TableName() string {
	return "conversations"
}

// Message - Individual message within a conversation with read receipts and optional media.
type Message struct {
	ID             uint `gorm:"primaryKey" json:"id"`
	ConversationID uint `gorm:"index;not null" json:"conversation_id"`
	SenderID       uint `gorm:"not null" json:"sender_id"` // UserID of whoever sent it (coach or client)

	Content   *string `gorm:"type:text" json:"content"`
	MediaURL  *string `json:"media_url"`  // S3 link for image/video attachment
	MediaType *string `json:"media_type"` // "image", "video"

	// Read receipt - timestamp when the other party read this message
	ReadAt *time.Time `json:"read_at"`

	CreatedAt time.Time `json:"created_at"`

	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"-"`
}

func (Message) TableName() string {
	return "messages"
}
