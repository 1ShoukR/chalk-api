package revenuecat

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	baseURL        = "https://api.revenuecat.com/v1"
	defaultTimeout = 10 * time.Second
)

// API defines the interface for RevenueCat operations
type API interface {
	// GetSubscriber fetches subscriber info by app user ID
	GetSubscriber(appUserID string) (*Subscriber, error)
	// ValidateWebhook validates the webhook signature and parses the event
	ValidateWebhook(body []byte, signature string) (*WebhookEvent, error)
}

// RevenueCat implements the API interface
type RevenueCat struct {
	httpClient    *http.Client
	apiKey        string
	webhookSecret string
}

// New creates a new RevenueCat API instance
func New(apiKey, webhookSecret string) *RevenueCat {
	return &RevenueCat{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
	}
}

// IsConfigured returns true if the API key is set
func (r *RevenueCat) IsConfigured() bool {
	return r.apiKey != ""
}

// GetSubscriber fetches subscriber info from RevenueCat
func (r *RevenueCat) GetSubscriber(appUserID string) (*Subscriber, error) {
	if !r.IsConfigured() {
		return nil, fmt.Errorf("RevenueCat API key not configured")
	}

	if appUserID == "" {
		return nil, fmt.Errorf("app user ID is required")
	}

	endpoint := fmt.Sprintf("%s/subscribers/%s", baseURL, appUserID)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Accept", "application/json")

	slog.Debug("RevenueCat request", "endpoint", "GetSubscriber", "userID", appUserID)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Subscriber not found
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result SubscriberResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result.Subscriber, nil
}

// ValidateWebhook validates the webhook signature and parses the event
// RevenueCat uses HMAC-SHA256 for webhook signatures
func (r *RevenueCat) ValidateWebhook(body []byte, signature string) (*WebhookEvent, error) {
	if r.webhookSecret == "" {
		// If no secret configured, skip validation (development mode)
		slog.Warn("RevenueCat webhook secret not configured, skipping signature validation")
	} else {
		// Validate HMAC signature
		mac := hmac.New(sha256.New, []byte(r.webhookSecret))
		mac.Write(body)
		expectedSig := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
			return nil, fmt.Errorf("invalid webhook signature")
		}
	}

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}

	slog.Debug("RevenueCat webhook received",
		"type", event.Event.Type,
		"userID", event.Event.AppUserID,
		"productID", event.Event.ProductID,
	)

	return &event, nil
}

// IsActiveSubscription checks if any subscription is currently active
func (s *Subscriber) IsActiveSubscription() bool {
	now := time.Now()
	for _, sub := range s.Subscriptions {
		if sub.ExpiresDate.After(now) {
			return true
		}
		// Check grace period
		if sub.GracePeriodExpiresDate != nil && sub.GracePeriodExpiresDate.After(now) {
			return true
		}
	}
	return false
}

// HasEntitlement checks if the subscriber has a specific entitlement
func (s *Subscriber) HasEntitlement(entitlementID string) bool {
	ent, ok := s.Entitlements[entitlementID]
	if !ok {
		return false
	}
	// Check if entitlement is active
	if ent.ExpiresDate == nil {
		return true // Lifetime entitlement
	}
	return ent.ExpiresDate.After(time.Now())
}

// GetActiveProductID returns the product ID of the active subscription (if any)
func (s *Subscriber) GetActiveProductID() string {
	now := time.Now()
	for productID, sub := range s.Subscriptions {
		if sub.ExpiresDate.After(now) {
			return productID
		}
	}
	return ""
}
