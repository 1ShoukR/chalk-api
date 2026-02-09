package expo

// PushMessage represents a push notification to send
type PushMessage struct {
	To       []string          `json:"to"`                 // Expo push tokens
	Title    string            `json:"title,omitempty"`
	Body     string            `json:"body"`
	Data     map[string]any    `json:"data,omitempty"`     // Custom data payload
	Sound    string            `json:"sound,omitempty"`    // "default" or custom sound
	Badge    *int              `json:"badge,omitempty"`    // iOS badge count
	TTL      int               `json:"ttl,omitempty"`      // Time to live in seconds
	Priority string            `json:"priority,omitempty"` // "default", "normal", "high"
	Subtitle string            `json:"subtitle,omitempty"` // iOS subtitle
	ChannelID string           `json:"channelId,omitempty"` // Android notification channel
	CategoryID string          `json:"categoryId,omitempty"` // Notification category for actions
}

// PushTicket is the response from sending a push notification
type PushTicket struct {
	ID      string       `json:"id,omitempty"`     // Ticket ID for checking receipt
	Status  string       `json:"status"`           // "ok" or "error"
	Message string       `json:"message,omitempty"` // Error message if status is "error"
	Details *TicketError `json:"details,omitempty"`
}

// TicketError contains error details
type TicketError struct {
	Error string `json:"error"` // Error code like "DeviceNotRegistered"
}

// PushReceipt is the delivery receipt for a sent notification
type PushReceipt struct {
	Status  string        `json:"status"`           // "ok" or "error"
	Message string        `json:"message,omitempty"`
	Details *ReceiptError `json:"details,omitempty"`
}

// ReceiptError contains error details for a receipt
type ReceiptError struct {
	Error string `json:"error"`
}

// SendResponse is the API response when sending notifications
type SendResponse struct {
	Data []PushTicket `json:"data"`
}

// ReceiptsResponse is the API response when fetching receipts
type ReceiptsResponse struct {
	Data map[string]PushReceipt `json:"data"`
}

// Error codes from Expo
const (
	ErrorDeviceNotRegistered = "DeviceNotRegistered"
	ErrorMessageTooBig       = "MessageTooBig"
	ErrorMessageRateExceeded = "MessageRateExceeded"
	ErrorInvalidCredentials  = "InvalidCredentials"
)

// Notification types for our app (used in Data payload)
const (
	NotificationTypeWorkoutAssigned   = "workout_assigned"
	NotificationTypeWorkoutCompleted  = "workout_completed"
	NotificationTypeSessionBooked     = "session_booked"
	NotificationTypeSessionReminder   = "session_reminder"
	NotificationTypeSessionCancelled  = "session_cancelled"
	NotificationTypeNewMessage        = "new_message"
	NotificationTypeInviteAccepted    = "invite_accepted"
	NotificationTypeProgressUpdate    = "progress_update"
)
