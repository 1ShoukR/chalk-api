package events

import (
	"chalk-api/pkg/external"
	"chalk-api/pkg/external/expo"
	"chalk-api/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

func RegisterDefaultHandlers(dispatcher *Dispatcher, integrations *external.Collection) error {
	if integrations != nil && integrations.Expo != nil {
		if err := dispatcher.Register(EventTypeNotificationPush, NewPushNotificationHandler(integrations.Expo)); err != nil {
			return err
		}
	}

	// Domain event handlers are logging placeholders for now.
	// These are ready to be upgraded into real side-effect handlers as services are implemented.
	if err := dispatcher.Register(EventTypeMessageSent, NewLoggingHandler("message.sent")); err != nil {
		return err
	}
	if err := dispatcher.Register(EventTypeWorkoutAssigned, NewLoggingHandler("workout.assigned")); err != nil {
		return err
	}
	if err := dispatcher.Register(EventTypeWorkoutCompleted, NewLoggingHandler("workout.completed")); err != nil {
		return err
	}
	if err := dispatcher.Register(EventTypeSessionBooked, NewLoggingHandler("session.booked")); err != nil {
		return err
	}
	if err := dispatcher.Register(EventTypeInviteAccepted, NewLoggingHandler("invite.accepted")); err != nil {
		return err
	}
	if err := dispatcher.Register(EventTypeSubscriptionChanged, NewLoggingHandler("subscription.changed")); err != nil {
		return err
	}

	return nil
}

func NewLoggingHandler(eventName string) Handler {
	return HandlerFunc(func(ctx context.Context, event models.OutboxEvent) error {
		slog.Info("Processed domain event", "event_name", eventName, "event_id", event.ID, "aggregate_id", event.AggregateID)
		return nil
	})
}

type PushNotificationHandler struct {
	expoAPI expo.API
}

func NewPushNotificationHandler(expoAPI expo.API) *PushNotificationHandler {
	return &PushNotificationHandler{expoAPI: expoAPI}
}

func (h *PushNotificationHandler) Handle(ctx context.Context, event models.OutboxEvent) error {
	var payload PushNotificationPayload
	if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
		return Permanent(fmt.Errorf("decode notification payload: %w", err))
	}

	if len(payload.Tokens) == 0 {
		return Permanent(fmt.Errorf("notification payload missing tokens"))
	}
	if payload.Body == "" {
		return Permanent(fmt.Errorf("notification payload missing body"))
	}

	message := expo.PushMessage{
		To:    payload.Tokens,
		Title: payload.Title,
		Body:  payload.Body,
		Data:  payload.Data,
		Sound: "default",
	}

	tickets, err := h.expoAPI.SendPush([]expo.PushMessage{message})
	if err != nil {
		return fmt.Errorf("send expo push: %w", err)
	}

	var transientFailures []string
	for _, ticket := range tickets {
		if ticket.Status != "error" {
			continue
		}

		errorCode := ""
		if ticket.Details != nil {
			errorCode = ticket.Details.Error
		}

		switch errorCode {
		case expo.ErrorMessageRateExceeded:
			transientFailures = append(transientFailures, fmt.Sprintf("%s: %s", errorCode, ticket.Message))
		case expo.ErrorDeviceNotRegistered, expo.ErrorMessageTooBig, expo.ErrorInvalidCredentials:
			slog.Warn("Non-retryable Expo ticket error",
				"event_id", event.ID,
				"error_code", errorCode,
				"message", ticket.Message,
			)
		default:
			// Unknown error: assume transient to avoid dropping a possibly recoverable delivery.
			transientFailures = append(transientFailures, fmt.Sprintf("%s: %s", errorCode, ticket.Message))
		}
	}

	if len(transientFailures) > 0 {
		return fmt.Errorf("expo transient ticket errors: %s", strings.Join(transientFailures, "; "))
	}

	return nil
}
