package revenuecat

import "time"

// SubscriberResponse is the API response when fetching subscriber info
type SubscriberResponse struct {
	RequestDate int64      `json:"request_date_ms"`
	Subscriber  Subscriber `json:"subscriber"`
}

// Subscriber represents a RevenueCat subscriber
type Subscriber struct {
	FirstSeen           time.Time              `json:"first_seen"`
	LastSeen            time.Time              `json:"last_seen"`
	ManagementURL       string                 `json:"management_url"`
	NonSubscriptions    map[string][]Purchase  `json:"non_subscriptions"`
	OriginalAppUserID   string                 `json:"original_app_user_id"`
	OriginalPurchaseDate *time.Time            `json:"original_purchase_date"`
	Subscriptions       map[string]Subscription `json:"subscriptions"`
	Entitlements        map[string]Entitlement  `json:"entitlements"`
}

// Subscription represents a subscription in RevenueCat
type Subscription struct {
	BillingIssuesDetectedAt *time.Time `json:"billing_issues_detected_at"`
	ExpiresDate             time.Time  `json:"expires_date"`
	GracePeriodExpiresDate  *time.Time `json:"grace_period_expires_date"`
	IsSandbox               bool       `json:"is_sandbox"`
	OriginalPurchaseDate    time.Time  `json:"original_purchase_date"`
	PeriodType              string     `json:"period_type"` // "normal", "trial", "intro"
	PurchaseDate            time.Time  `json:"purchase_date"`
	Store                   string     `json:"store"` // "app_store", "play_store", "stripe"
	UnsubscribeDetectedAt   *time.Time `json:"unsubscribe_detected_at"`
}

// Entitlement represents an entitlement (access level) in RevenueCat
type Entitlement struct {
	ExpiresDate          *time.Time `json:"expires_date"`
	GracePeriodExpiresDate *time.Time `json:"grace_period_expires_date"`
	ProductIdentifier    string     `json:"product_identifier"`
	PurchaseDate         time.Time  `json:"purchase_date"`
}

// Purchase represents a non-subscription purchase
type Purchase struct {
	ID           string    `json:"id"`
	PurchaseDate time.Time `json:"purchase_date"`
	Store        string    `json:"store"`
}

// WebhookEvent represents an incoming webhook from RevenueCat
type WebhookEvent struct {
	APIVersion string       `json:"api_version"`
	Event      EventPayload `json:"event"`
}

// EventPayload contains the webhook event details
type EventPayload struct {
	// Event identification
	ID   string `json:"id"`
	Type string `json:"type"` // See EventType constants

	// User identification
	AppUserID         string `json:"app_user_id"`
	OriginalAppUserID string `json:"original_app_user_id"`

	// Product info
	ProductID          string  `json:"product_id"`
	EntitlementIDs     []string `json:"entitlement_ids"`
	PeriodType         string  `json:"period_type"` // "TRIAL", "INTRO", "NORMAL"
	PresentedOfferingID string `json:"presented_offering_id"`

	// Pricing
	Price         float64 `json:"price"`
	Currency      string  `json:"currency"`
	PriceInPurchasedCurrency float64 `json:"price_in_purchased_currency"`

	// Store info
	Store       string `json:"store"` // "APP_STORE", "PLAY_STORE", "STRIPE"
	Environment string `json:"environment"` // "SANDBOX", "PRODUCTION"

	// Timestamps
	EventTimestampMs      int64      `json:"event_timestamp_ms"`
	PurchasedAtMs         int64      `json:"purchased_at_ms"`
	ExpirationAtMs        *int64     `json:"expiration_at_ms"`
	OriginalPurchaseDateMs int64     `json:"original_purchase_date_ms"`

	// Cancellation
	CancelReason          *string `json:"cancel_reason"`
	ExpirationReason      *string `json:"expiration_reason"`

	// Transaction IDs
	TransactionID         string  `json:"transaction_id"`
	OriginalTransactionID string  `json:"original_transaction_id"`

	// Subscriber attributes
	SubscriberAttributes map[string]AttributeValue `json:"subscriber_attributes"`
}

// AttributeValue represents a subscriber attribute
type AttributeValue struct {
	Value     string `json:"value"`
	UpdatedAt int64  `json:"updated_at_ms"`
}

// Event types from RevenueCat webhooks
const (
	EventTypeInitialPurchase       = "INITIAL_PURCHASE"
	EventTypeRenewal               = "RENEWAL"
	EventTypeCancellation          = "CANCELLATION"
	EventTypeUncancellation        = "UNCANCELLATION"
	EventTypeNonRenewingPurchase   = "NON_RENEWING_PURCHASE"
	EventTypeSubscriptionPaused    = "SUBSCRIPTION_PAUSED"
	EventTypeBillingIssue          = "BILLING_ISSUE"
	EventTypeProductChange         = "PRODUCT_CHANGE"
	EventTypeExpiration            = "EXPIRATION"
	EventTypeSubscriptionExtended  = "SUBSCRIPTION_EXTENDED"
	EventTypeTransfer              = "TRANSFER"
)
