package services

import (
	"chalk-api/pkg/external/revenuecat"
	"chalk-api/pkg/models"
	"chalk-api/pkg/repositories"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidSubscriptionWebhookAuth = errors.New("invalid subscription webhook authorization")
	ErrSubscriptionWebhookPayload     = errors.New("invalid subscription webhook payload")
	ErrFeatureNameRequired            = errors.New("feature name is required")
)

type SubscriptionService struct {
	repos                 *repositories.RepositoriesCollection
	subscriptionRepo      *repositories.SubscriptionRepository
	revenueCat            revenuecat.API
	supportedWebhookTypes map[string]struct{}
}

type FeatureAccessResult struct {
	Feature            string `json:"feature"`
	Allowed            bool   `json:"allowed"`
	Reason             string `json:"reason"`
	SubscriptionStatus string `json:"subscription_status"`
}

func NewSubscriptionService(
	repos *repositories.RepositoriesCollection,
	revenueCatAPI revenuecat.API,
) *SubscriptionService {
	return &SubscriptionService{
		repos:            repos,
		subscriptionRepo: repos.Subscription,
		revenueCat:       revenueCatAPI,
		supportedWebhookTypes: map[string]struct{}{
			revenuecat.EventTypeTest:                 {},
			revenuecat.EventTypeInitialPurchase:      {},
			revenuecat.EventTypeRenewal:              {},
			revenuecat.EventTypeCancellation:         {},
			revenuecat.EventTypeBillingIssue:         {},
			revenuecat.EventTypeExpiration:           {},
			revenuecat.EventTypeUncancellation:       {},
			revenuecat.EventTypeProductChange:        {},
			revenuecat.EventTypeSubscriptionExtended: {},
			revenuecat.EventTypeTransfer:             {},
		},
	}
}

func (s *SubscriptionService) HandleRevenueCatWebhook(
	ctx context.Context,
	rawBody []byte,
	authorizationHeader string,
) error {
	if s.revenueCat == nil {
		return fmt.Errorf("revenuecat integration is not configured")
	}

	webhookEvent, err := s.revenueCat.ValidateWebhook(rawBody, authorizationHeader)
	if err != nil {
		errLower := strings.ToLower(err.Error())
		if strings.Contains(errLower, "authorization") || strings.Contains(errLower, "missing webhook") {
			return ErrInvalidSubscriptionWebhookAuth
		}
		return ErrSubscriptionWebhookPayload
	}

	if _, supported := s.supportedWebhookTypes[webhookEvent.Event.Type]; !supported {
		// Ignore events we intentionally do not process yet.
		return nil
	}

	if webhookEvent.Event.Type == revenuecat.EventTypeTest {
		return nil
	}

	lookupAppUserID := deriveLookupAppUserID(&webhookEvent.Event)
	userID := deriveLocalUserID(&webhookEvent.Event)
	eventID := strings.TrimSpace(webhookEvent.Event.ID)

	// RevenueCat recommends syncing status via GET /subscribers after receiving webhooks.
	// If this call fails, we still proceed with the webhook payload to favor availability.
	subscriber, subscriberErr := s.fetchSubscriber(ctx, lookupAppUserID)
	if subscriberErr != nil {
		slog.Warn("RevenueCat subscriber sync failed, falling back to webhook payload",
			"event_id", eventID,
			"lookup_app_user_id", lookupAppUserID,
			"error", subscriberErr,
		)
	}

	if userID == 0 && subscriber != nil {
		userID = parseUintUserID(subscriber.OriginalAppUserID)
	}
	if userID == 0 {
		// We intentionally ack unknown users to avoid perpetual webhook retries.
		slog.Warn("Skipping webhook: could not map RevenueCat app_user_id to local user",
			"event_id", eventID,
			"app_user_id", webhookEvent.Event.AppUserID,
			"original_app_user_id", webhookEvent.Event.OriginalAppUserID,
		)
		return nil
	}

	return s.repos.WithTransaction(ctx, func(tx *gorm.DB, txRepos *repositories.RepositoriesCollection) error {
		if eventID != "" {
			if _, err := txRepos.Subscription.GetEventByRevenueCatID(ctx, eventID); err == nil {
				return nil
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		if _, err := txRepos.User.GetByID(ctx, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Warn("Skipping webhook: local user not found", "event_id", eventID, "user_id", userID)
				return nil
			}
			return err
		}

		subscription, err := txRepos.Subscription.GetByUserID(ctx, userID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			subscription = &models.Subscription{
				UserID: userID,
				Status: "inactive",
			}
			if err := txRepos.Subscription.Create(ctx, subscription); err != nil {
				return err
			}
		}

		applyWebhookToSubscription(subscription, webhookEvent, subscriber, lookupAppUserID)

		if err := txRepos.Subscription.Update(ctx, subscription); err != nil {
			return err
		}

		eventRecord := buildSubscriptionEventRecord(subscription.ID, webhookEvent, rawBody)
		if err := txRepos.Subscription.CreateEvent(ctx, eventRecord); err != nil {
			if isDuplicateConstraintError(err) {
				return nil
			}
			return err
		}

		return nil
	})
}

func (s *SubscriptionService) GetMySubscription(ctx context.Context, userID uint) (*models.Subscription, error) {
	sub, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.Subscription{
				UserID:    userID,
				Status:    "inactive",
				WillRenew: false,
			}, nil
		}
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) CheckFeatureAccess(ctx context.Context, userID uint, feature string) (*FeatureAccessResult, error) {
	normalizedFeature := strings.TrimSpace(strings.ToLower(feature))
	if normalizedFeature == "" {
		return nil, ErrFeatureNameRequired
	}

	sub, err := s.GetMySubscription(ctx, userID)
	if err != nil {
		return nil, err
	}

	if isFeatureFree(normalizedFeature) {
		return &FeatureAccessResult{
			Feature:            normalizedFeature,
			Allowed:            true,
			Reason:             "free_feature",
			SubscriptionStatus: sub.Status,
		}, nil
	}

	hasPaidAccess := hasPaidSubscriptionAccess(sub.Status)
	reason := "subscription_required"
	if hasPaidAccess {
		reason = "subscription_active"
	}

	return &FeatureAccessResult{
		Feature:            normalizedFeature,
		Allowed:            hasPaidAccess,
		Reason:             reason,
		SubscriptionStatus: sub.Status,
	}, nil
}

func (s *SubscriptionService) fetchSubscriber(ctx context.Context, appUserID string) (*revenuecat.Subscriber, error) {
	appUserID = strings.TrimSpace(appUserID)
	if appUserID == "" {
		return nil, nil
	}
	if s.revenueCat == nil {
		return nil, nil
	}
	return s.revenueCat.GetSubscriber(appUserID)
}

func applyWebhookToSubscription(
	sub *models.Subscription,
	webhookEvent *revenuecat.WebhookEvent,
	subscriber *revenuecat.Subscriber,
	lookupAppUserID string,
) {
	event := webhookEvent.Event
	now := time.Now()

	if lookupAppUserID != "" {
		sub.RevenueCatCustomerID = strPtr(lookupAppUserID)
	}

	sub.StoreEnvironment = normalizeEnvironmentPtr(event.Environment)
	sub.CancellationReason = nil

	if event.TransactionID != "" {
		sub.LatestTransactionID = strPtr(event.TransactionID)
	}
	if event.OriginalTransactionID != "" {
		sub.OriginalTransactionID = strPtr(event.OriginalTransactionID)
	}

	if event.PurchasedAtMs > 0 {
		purchasedAt := time.UnixMilli(event.PurchasedAtMs)
		if sub.FirstPurchasedAt == nil || purchasedAt.Before(*sub.FirstPurchasedAt) {
			sub.FirstPurchasedAt = &purchasedAt
		}
		sub.LastRenewalAt = &purchasedAt
	}

	if event.ExpirationAtMs != nil && *event.ExpirationAtMs > 0 {
		exp := time.UnixMilli(*event.ExpirationAtMs)
		sub.CurrentPeriodEnd = &exp
		sub.ExpiresAt = &exp
	}

	productID := strings.TrimSpace(event.ProductID)
	if productID != "" {
		sub.ProductID = &productID
	}
	sub.Platform = normalizePlatformPtr(event.Store)

	latestProductID, latestSubscription := pickLatestSubscription(subscriber)
	if latestSubscription != nil {
		if latestProductID != "" {
			sub.ProductID = &latestProductID
		}
		sub.Platform = normalizePlatformPtr(latestSubscription.Store)

		purchaseDate := latestSubscription.PurchaseDate
		sub.CurrentPeriodStart = &purchaseDate
		sub.LastRenewalAt = &purchaseDate

		expirationDate := latestSubscription.ExpiresDate
		sub.CurrentPeriodEnd = &expirationDate
		sub.ExpiresAt = &expirationDate

		if latestSubscription.PeriodType == "trial" || strings.EqualFold(event.PeriodType, "TRIAL") {
			sub.TrialStart = &purchaseDate
			sub.TrialEnd = &expirationDate
		}

		sub.UnsubscribeDetectedAt = latestSubscription.UnsubscribeDetectedAt
		sub.BillingIssueDetectedAt = latestSubscription.BillingIssuesDetectedAt
		sub.WillRenew = latestSubscription.UnsubscribeDetectedAt == nil
	}

	switch event.Type {
	case revenuecat.EventTypeInitialPurchase, revenuecat.EventTypeRenewal, revenuecat.EventTypeUncancellation:
		sub.CancelledAt = nil
	case revenuecat.EventTypeCancellation:
		cancelledAt := unixMilliOrNow(event.EventTimestampMs, now)
		sub.CancelledAt = &cancelledAt
		sub.WillRenew = false
	case revenuecat.EventTypeExpiration:
		cancelledAt := unixMilliOrNow(event.EventTimestampMs, now)
		sub.CancelledAt = &cancelledAt
		sub.WillRenew = false
	case revenuecat.EventTypeBillingIssue:
		if sub.BillingIssueDetectedAt == nil {
			detectedAt := unixMilliOrNow(event.EventTimestampMs, now)
			sub.BillingIssueDetectedAt = &detectedAt
		}
	}

	if event.CancelReason != nil {
		sub.CancellationReason = event.CancelReason
	}
	if event.ExpirationReason != nil {
		sub.CancellationReason = event.ExpirationReason
	}

	sub.Status = deriveSubscriptionStatus(subscriber, &event)
	if sub.Status == "expired" || sub.Status == "inactive" {
		sub.WillRenew = false
	}
}

func buildSubscriptionEventRecord(
	subscriptionID uint,
	webhookEvent *revenuecat.WebhookEvent,
	rawBody []byte,
) *models.SubscriptionEvent {
	event := webhookEvent.Event
	eventID := strings.TrimSpace(event.ID)
	eventType := strings.TrimSpace(event.Type)
	productID := trimToPtr(event.ProductID)
	currency := trimToPtr(event.Currency)
	platform := normalizePlatformPtr(event.Store)
	raw := string(rawBody)
	processedAt := unixMilliOrNow(event.EventTimestampMs, time.Now())

	return &models.SubscriptionEvent{
		SubscriptionID:    subscriptionID,
		EventType:         eventType,
		RevenueCatEventID: trimToPtr(eventID),
		RawPayload:        &raw,
		ProductID:         productID,
		PriceInCents:      priceToCents(event.PriceInPurchasedCurrency),
		Currency:          currency,
		Platform:          platform,
		ProcessedAt:       processedAt,
	}
}

func deriveSubscriptionStatus(subscriber *revenuecat.Subscriber, event *revenuecat.EventPayload) string {
	switch event.Type {
	case revenuecat.EventTypeExpiration:
		return "expired"
	case revenuecat.EventTypeBillingIssue:
		return "grace_period"
	}

	if strings.EqualFold(event.PeriodType, "TRIAL") {
		return "in_trial"
	}
	if subscriber != nil && subscriber.IsActiveSubscription() {
		return "active"
	}
	if event.Type == revenuecat.EventTypeCancellation {
		return "canceled"
	}
	return "inactive"
}

func pickLatestSubscription(subscriber *revenuecat.Subscriber) (string, *revenuecat.Subscription) {
	if subscriber == nil || len(subscriber.Subscriptions) == 0 {
		return "", nil
	}

	now := time.Now()
	var pickedKey string
	var picked *revenuecat.Subscription

	for productID, sub := range subscriber.Subscriptions {
		current := sub
		if picked == nil {
			picked = &current
			pickedKey = productID
			continue
		}

		// Prefer active subscriptions first, then newest expiration.
		pickedActive := picked.ExpiresDate.After(now)
		currentActive := current.ExpiresDate.After(now)
		if currentActive && !pickedActive {
			picked = &current
			pickedKey = productID
			continue
		}
		if currentActive == pickedActive && current.ExpiresDate.After(picked.ExpiresDate) {
			picked = &current
			pickedKey = productID
		}
	}

	return pickedKey, picked
}

func deriveLookupAppUserID(event *revenuecat.EventPayload) string {
	candidates := append([]string{}, event.AppUserID, event.OriginalAppUserID)
	candidates = append(candidates, event.Aliases...)
	candidates = append(candidates, event.TransferredTo...)
	candidates = append(candidates, event.TransferredFrom...)

	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

func deriveLocalUserID(event *revenuecat.EventPayload) uint {
	candidates := append([]string{}, event.AppUserID, event.OriginalAppUserID)
	candidates = append(candidates, event.Aliases...)
	candidates = append(candidates, event.TransferredTo...)
	candidates = append(candidates, event.TransferredFrom...)

	for _, candidate := range candidates {
		if parsed := parseUintUserID(candidate); parsed > 0 {
			return parsed
		}
	}
	return 0
}

func parseUintUserID(raw string) uint {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed == 0 {
		return 0
	}
	return uint(parsed)
}

func hasPaidSubscriptionAccess(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "in_trial", "grace_period":
		return true
	default:
		return false
	}
}

func isFeatureFree(feature string) bool {
	switch feature {
	case "health_check", "public_profile":
		return true
	default:
		return false
	}
}

func normalizePlatformPtr(store string) *string {
	store = strings.ToUpper(strings.TrimSpace(store))
	if store == "" {
		return nil
	}

	var platform string
	switch store {
	case "APP_STORE", "MAC_APP_STORE":
		platform = "ios"
	case "PLAY_STORE":
		platform = "android"
	case "STRIPE", "RC_BILLING":
		platform = "web"
	case "ROKU":
		platform = "roku"
	default:
		platform = strings.ToLower(store)
	}
	return &platform
}

func normalizeEnvironmentPtr(environment string) *string {
	environment = strings.ToUpper(strings.TrimSpace(environment))
	if environment == "" {
		return nil
	}
	normalized := strings.ToLower(environment)
	return &normalized
}

func priceToCents(price float64) *int {
	cents := int(math.Round(price * 100))
	return &cents
}

func trimToPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func strPtr(value string) *string {
	return &value
}

func unixMilliOrNow(ms int64, fallback time.Time) time.Time {
	if ms <= 0 {
		return fallback
	}
	return time.UnixMilli(ms)
}

func isDuplicateConstraintError(err error) bool {
	if err == nil {
		return false
	}
	normalized := strings.ToLower(err.Error())
	return strings.Contains(normalized, "duplicate key") ||
		strings.Contains(normalized, "unique constraint") ||
		strings.Contains(normalized, "unique violation")
}
