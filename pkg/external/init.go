package external

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/external/expo"
	"chalk-api/pkg/external/openfoodfacts"
	"chalk-api/pkg/external/revenuecat"
	"log/slog"
)

// Collection contains all external API integrations
type Collection struct {
	OpenFoodFacts openfoodfacts.API
	RevenueCat    revenuecat.API
	Expo          expo.API
}

// Initialize creates all external API integrations
func Initialize(cfg config.Environment) *Collection {
	webhookAuthorization := cfg.RevenueCatWebhookAuthorization
	if webhookAuthorization == "" {
		// Backward compatibility for existing deployments that still use REVENUECAT_WEBHOOK_SECRET.
		webhookAuthorization = cfg.RevenueCatWebhookSecret
	}

	collection := &Collection{
		OpenFoodFacts: openfoodfacts.New(cfg.OpenFoodFactsUserAgent),
		RevenueCat:    revenuecat.New(cfg.RevenueCatAPIKey, webhookAuthorization),
		Expo:          expo.New(cfg.ExpoAccessToken),
	}

	// Log which integrations are configured
	if cfg.RevenueCatAPIKey != "" {
		slog.Info("RevenueCat integration configured")
	} else {
		slog.Warn("RevenueCat API key not set, subscription features disabled")
	}

	if cfg.ExpoAccessToken != "" {
		slog.Info("Expo push notifications configured with auth")
	} else {
		slog.Info("Expo push notifications configured without auth (rate limited)")
	}

	slog.Info("Open Food Facts integration configured", "userAgent", cfg.OpenFoodFactsUserAgent)

	return collection
}
