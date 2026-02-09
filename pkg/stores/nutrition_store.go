package stores

import (
	"chalk-api/pkg/models"
	"time"
)

// NutritionStore handles food item caching
// Aggressive caching for Open Food Facts API to minimize API calls
type NutritionStore struct {
	redis *RedisClient
}

const (
	// Open Food Facts data is stable, cache aggressively
	FoodItemTTL     = 7 * 24 * time.Hour // 7 days
	FoodSearchTTL   = 24 * time.Hour     // 1 day for search results
)

// NewNutritionStore creates a new nutrition store
func NewNutritionStore(redis *RedisClient) *NutritionStore {
	return &NutritionStore{redis: redis}
}

// CachedFoodItem represents a cached food item from Open Food Facts or custom
// Mirrors model pointer types to avoid unnecessary conversions
type CachedFoodItem struct {
	ID               uint     `json:"id"`
	Name             string   `json:"name"`
	Brand            *string  `json:"brand,omitempty"`
	Barcode          *string  `json:"barcode,omitempty"`
	ServingSize      *string  `json:"serving_size,omitempty"`
	ServingSizeGrams *float64 `json:"serving_size_grams,omitempty"`
	Calories         *int     `json:"calories,omitempty"`
	ProteinGrams     *float64 `json:"protein_grams,omitempty"`
	CarbsGrams       *float64 `json:"carbs_grams,omitempty"`
	FatGrams         *float64 `json:"fat_grams,omitempty"`
	FiberGrams       *float64 `json:"fiber_grams,omitempty"`
	SugarGrams       *float64 `json:"sugar_grams,omitempty"`
	SodiumMg         *float64 `json:"sodium_mg,omitempty"`
	Source           string   `json:"source"`
	ExternalID       *string  `json:"external_id,omitempty"`
	IsSystem         bool     `json:"is_system"`
}

// ToCachedFoodItem converts a models.FoodItem to cached version
func ToCachedFoodItem(f *models.FoodItem) *CachedFoodItem {
	if f == nil {
		return nil
	}
	return &CachedFoodItem{
		ID:               f.ID,
		Name:             f.Name,
		Brand:            f.Brand,
		Barcode:          f.Barcode,
		ServingSize:      f.ServingSize,
		ServingSizeGrams: f.ServingSizeGrams,
		Calories:         f.Calories,
		ProteinGrams:     f.ProteinGrams,
		CarbsGrams:       f.CarbsGrams,
		FatGrams:         f.FatGrams,
		FiberGrams:       f.FiberGrams,
		SugarGrams:       f.SugarGrams,
		SodiumMg:         f.SodiumMg,
		Source:           f.Source,
		ExternalID:       f.ExternalID,
		IsSystem:         f.IsSystem,
	}
}

// GetByBarcode retrieves a cached food item by barcode
func (s *NutritionStore) GetByBarcode(barcode string) (*CachedFoodItem, bool) {
	if !s.redis.IsAvailable() || barcode == "" {
		return nil, false
	}

	var food CachedFoodItem
	if s.redis.GetJSON(KeyFoodByBarcode(barcode), &food) {
		return &food, true
	}
	return nil, false
}

// SetByBarcode caches a food item by barcode
func (s *NutritionStore) SetByBarcode(food *models.FoodItem) {
	if !s.redis.IsAvailable() || food == nil || food.Barcode == nil || *food.Barcode == "" {
		return
	}

	cached := ToCachedFoodItem(food)
	s.redis.SetJSON(KeyFoodByBarcode(*food.Barcode), cached, FoodItemTTL)
}

// GetByExternalID retrieves a cached food item by source and external ID
func (s *NutritionStore) GetByExternalID(source, externalID string) (*CachedFoodItem, bool) {
	if !s.redis.IsAvailable() || source == "" || externalID == "" {
		return nil, false
	}

	var food CachedFoodItem
	if s.redis.GetJSON(KeyFoodByExternalID(source, externalID), &food) {
		return &food, true
	}
	return nil, false
}

// SetByExternalID caches a food item by source and external ID
func (s *NutritionStore) SetByExternalID(food *models.FoodItem) {
	if !s.redis.IsAvailable() || food == nil || food.Source == "" || food.ExternalID == nil || *food.ExternalID == "" {
		return
	}

	cached := ToCachedFoodItem(food)
	s.redis.SetJSON(KeyFoodByExternalID(food.Source, *food.ExternalID), cached, FoodItemTTL)
}

// Set caches a food item by both barcode and external ID if available
func (s *NutritionStore) Set(food *models.FoodItem) {
	if !s.redis.IsAvailable() || food == nil {
		return
	}

	s.SetByBarcode(food)
	s.SetByExternalID(food)
}

// GetSearchResults retrieves cached search results
func (s *NutritionStore) GetSearchResults(query string, page int) ([]CachedFoodItem, bool) {
	if !s.redis.IsAvailable() || query == "" {
		return nil, false
	}

	var foods []CachedFoodItem
	if s.redis.GetJSON(KeyFoodSearch(query, page), &foods) {
		return foods, true
	}
	return nil, false
}

// SetSearchResults caches search results
func (s *NutritionStore) SetSearchResults(query string, page int, foods []models.FoodItem) {
	if !s.redis.IsAvailable() || query == "" {
		return
	}

	cached := make([]CachedFoodItem, len(foods))
	for i := range foods {
		cached[i] = *ToCachedFoodItem(&foods[i])
	}

	s.redis.SetJSON(KeyFoodSearch(query, page), cached, FoodSearchTTL)
}

// InvalidateByBarcode removes a food item cache by barcode
func (s *NutritionStore) InvalidateByBarcode(barcode string) {
	if s.redis.IsAvailable() && barcode != "" {
		s.redis.Delete(KeyFoodByBarcode(barcode))
	}
}

// InvalidateSearchResults clears all search result caches
// Call this when new foods are added that might affect search results
func (s *NutritionStore) InvalidateSearchResults() {
	if s.redis.IsAvailable() {
		s.redis.DeletePattern("food:search:*")
	}
}
