package models

import "time"

// Subscription - Tracks user subscription status (updated via RevenueCat webhooks)
type Subscription struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"uniqueIndex;not null" json:"user_id"` // One subscription per user

	// RevenueCat Integration
	RevenueCatCustomerID *string `gorm:"index" json:"revenuecat_customer_id"` // Customer ID in RevenueCat
	ProductID            *string `json:"product_id"`                          // Which subscription tier (e.g., "pro_monthly", "enterprise_yearly")
	Platform             *string `json:"platform"`                            // "ios", "android", "stripe" (web)

	// Subscription Status
	Status string `gorm:"default:'inactive';index" json:"status"` // "active", "inactive", "canceled", "expired", "in_trial", "grace_period"

	// Billing Period
	CurrentPeriodStart *time.Time `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end"`

	// Trial
	TrialStart *time.Time `json:"trial_start"`
	TrialEnd   *time.Time `json:"trial_end"`

	// Cancellation
	CancelledAt               *time.Time `json:"cancelled_at"`
	CancellationReason        *string    `json:"cancellation_reason"`
	WillRenew                 bool       `gorm:"default:true" json:"will_renew"` // False if user cancelled but still in billing period
	UnsubscribeDetectedAt     *time.Time `json:"unsubscribe_detected_at"`
	BillingIssueDetectedAt    *time.Time `json:"billing_issue_detected_at"`

	// Store-specific data
	OriginalTransactionID     *string `json:"original_transaction_id"`      // Apple/Google original purchase ID
	LatestTransactionID       *string `json:"latest_transaction_id"`        // Most recent renewal
	StoreEnvironment          *string `json:"store_environment"`            // "production", "sandbox"

	// Timestamps
	FirstPurchasedAt          *time.Time `json:"first_purchased_at"` // When they first became a paying customer
	LastRenewalAt             *time.Time `json:"last_renewal_at"`
	ExpiresAt                 *time.Time `gorm:"index" json:"expires_at"` // When current access expires

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

// SubscriptionEvent - Log of all subscription events from RevenueCat (for debugging/audit)
type SubscriptionEvent struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	SubscriptionID uint   `gorm:"index;not null" json:"subscription_id"`
	EventType      string `gorm:"not null;index" json:"event_type"` // "initial_purchase", "renewal", "cancellation", "billing_issue", "expiration", "reactivation"

	// Event data (store raw webhook payload)
	RevenueCatEventID   *string `gorm:"uniqueIndex" json:"revenuecat_event_id"` // Prevent duplicate processing
	RawPayload          *string `gorm:"type:jsonb" json:"-"`                    // Full webhook JSON for debugging
	
	// Key fields from event
	ProductID           *string `json:"product_id"`
	PriceInCents        *int    `json:"price_in_cents"`
	Currency            *string `json:"currency"`
	Platform            *string `json:"platform"`

	ProcessedAt         time.Time `json:"processed_at"`

	// Relations
	Subscription Subscription `gorm:"foreignKey:SubscriptionID" json:"-"`
}

func (SubscriptionEvent) TableName() string {
	return "subscription_events"
}
