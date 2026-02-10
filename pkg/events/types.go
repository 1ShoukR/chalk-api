package events

import (
	"fmt"
	"strings"
	"time"
)

type EventType string

const (
	EventTypeMessageSent         EventType = "message.sent"
	EventTypeWorkoutAssigned     EventType = "workout.assigned"
	EventTypeWorkoutCompleted    EventType = "workout.completed"
	EventTypeSessionBooked       EventType = "session.booked"
	EventTypeInviteAccepted      EventType = "invite.accepted"
	EventTypeSubscriptionChanged EventType = "subscription.changed"
	EventTypeNotificationPush    EventType = "notification.push"
)

type MessageSentPayload struct {
	MessageID      uint `json:"message_id"`
	ConversationID uint `json:"conversation_id"`
	SenderID       uint `json:"sender_id"`
	RecipientID    uint `json:"recipient_id"`
}

type WorkoutAssignedPayload struct {
	WorkoutID      uint   `json:"workout_id"`
	CoachID        uint   `json:"coach_id"`
	ClientID       uint   `json:"client_id"`
	ScheduledDate  string `json:"scheduled_date"`
	WorkoutName    string `json:"workout_name"`
	AssignedByUser uint   `json:"assigned_by_user"`
}

type WorkoutCompletedPayload struct {
	WorkoutID   uint      `json:"workout_id"`
	CoachID     uint      `json:"coach_id"`
	ClientID    uint      `json:"client_id"`
	CompletedAt time.Time `json:"completed_at"`
}

type SessionBookedPayload struct {
	SessionID   uint      `json:"session_id"`
	CoachID     uint      `json:"coach_id"`
	ClientID    uint      `json:"client_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
	BookedBy    string    `json:"booked_by"` // "coach" or "client"
}

type InviteAcceptedPayload struct {
	InviteCodeID    uint   `json:"invite_code_id"`
	CoachID         uint   `json:"coach_id"`
	ClientUserID    uint   `json:"client_user_id"`
	ClientProfileID uint   `json:"client_profile_id"`
	Code            string `json:"code"`
}

type SubscriptionChangedPayload struct {
	SubscriptionID    uint    `json:"subscription_id"`
	UserID            uint    `json:"user_id"`
	PreviousStatus    string  `json:"previous_status"`
	CurrentStatus     string  `json:"current_status"`
	ProductID         *string `json:"product_id,omitempty"`
	RevenueCatEventID *string `json:"revenuecat_event_id,omitempty"`
}

// PushNotificationPayload is used by notification.push events.
// Domain events can fan out into this event type for delivery.
type PushNotificationPayload struct {
	Tokens []string       `json:"tokens"`
	Title  string         `json:"title"`
	Body   string         `json:"body"`
	Data   map[string]any `json:"data,omitempty"`
}

func BuildIdempotencyKey(eventType EventType, parts ...string) string {
	base := string(eventType)
	if len(parts) == 0 {
		return base
	}
	return fmt.Sprintf("%s:%s", base, strings.Join(parts, ":"))
}
