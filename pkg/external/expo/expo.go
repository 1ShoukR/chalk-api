package expo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	pushURL        = "https://exp.host/--/api/v2/push/send"
	receiptsURL    = "https://exp.host/--/api/v2/push/getReceipts"
	defaultTimeout = 10 * time.Second
	maxBatchSize   = 100 // Expo's limit per request
)

// API defines the interface for Expo Push operations
type API interface {
	// SendPush sends push notifications to one or more devices
	SendPush(messages []PushMessage) ([]PushTicket, error)
	// GetReceipts fetches delivery receipts for sent notifications
	GetReceipts(ticketIDs []string) (map[string]PushReceipt, error)
}

// Expo implements the API interface
type Expo struct {
	httpClient  *http.Client
	accessToken string
}

// New creates a new Expo Push API instance
func New(accessToken string) *Expo {
	return &Expo{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		accessToken: accessToken,
	}
}

// IsConfigured returns true if access token is set
// Note: Expo push works without auth, but rate limits are higher with token
func (e *Expo) IsConfigured() bool {
	return e.accessToken != ""
}

// SendPush sends push notifications
// Automatically batches large requests to respect Expo's limits
func (e *Expo) SendPush(messages []PushMessage) ([]PushTicket, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	// Flatten messages with multiple recipients into individual messages
	var expandedMessages []PushMessage
	for _, msg := range messages {
		if len(msg.To) <= 1 {
			expandedMessages = append(expandedMessages, msg)
		} else {
			// Split into individual messages for each recipient
			for _, token := range msg.To {
				copy := msg
				copy.To = []string{token}
				expandedMessages = append(expandedMessages, copy)
			}
		}
	}

	var allTickets []PushTicket

	// Process in batches
	for i := 0; i < len(expandedMessages); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(expandedMessages) {
			end = len(expandedMessages)
		}
		batch := expandedMessages[i:end]

		tickets, err := e.sendBatch(batch)
		if err != nil {
			return allTickets, fmt.Errorf("batch %d failed: %w", i/maxBatchSize, err)
		}
		allTickets = append(allTickets, tickets...)
	}

	return allTickets, nil
}

// sendBatch sends a single batch of messages
func (e *Expo) sendBatch(messages []PushMessage) ([]PushTicket, error) {
	body, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal messages: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, pushURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if e.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+e.accessToken)
	}

	slog.Debug("Expo push request", "count", len(messages))

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result SendResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Data, nil
}

// GetReceipts fetches delivery receipts for the given ticket IDs
func (e *Expo) GetReceipts(ticketIDs []string) (map[string]PushReceipt, error) {
	if len(ticketIDs) == 0 {
		return nil, nil
	}

	payload := map[string][]string{"ids": ticketIDs}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, receiptsURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if e.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+e.accessToken)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result ReceiptsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Data, nil
}

// Helper functions for building common notifications

// NewWorkoutAssignedNotification creates a notification for when a workout is assigned
func NewWorkoutAssignedNotification(token string, coachName, workoutName string) PushMessage {
	return PushMessage{
		To:    []string{token},
		Title: "New Workout Assigned",
		Body:  fmt.Sprintf("%s assigned you a workout: %s", coachName, workoutName),
		Sound: "default",
		Data: map[string]any{
			"type": NotificationTypeWorkoutAssigned,
		},
	}
}

// NewMessageNotification creates a notification for a new message
func NewMessageNotification(token string, senderName, preview string) PushMessage {
	return PushMessage{
		To:    []string{token},
		Title: senderName,
		Body:  preview,
		Sound: "default",
		Data: map[string]any{
			"type": NotificationTypeNewMessage,
		},
	}
}

// NewSessionReminderNotification creates a reminder notification for an upcoming session
func NewSessionReminderNotification(token string, sessionTime time.Time, otherPartyName string) PushMessage {
	return PushMessage{
		To:    []string{token},
		Title: "Session Reminder",
		Body:  fmt.Sprintf("Your session with %s starts in 1 hour", otherPartyName),
		Sound: "default",
		Data: map[string]any{
			"type":        NotificationTypeSessionReminder,
			"sessionTime": sessionTime.Unix(),
		},
	}
}
