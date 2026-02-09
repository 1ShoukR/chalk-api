package models

import "time"

// NutritionTarget - Macro/calorie goals for a client, set by either coach or client.
// Effective date allows scheduling future target changes (e.g., cut → bulk transition).
type NutritionTarget struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	ClientID uint `gorm:"index;not null" json:"client_id"`

	Calories     *int `json:"calories"`
	ProteinGrams *int `json:"protein_grams"`
	CarbsGrams   *int `json:"carbs_grams"`
	FatGrams     *int `json:"fat_grams"`
	FiberGrams   *int `json:"fiber_grams"`

	EffectiveDate string `gorm:"type:date;not null" json:"effective_date"` // when these targets take effect
	CreatedBy     uint   `gorm:"not null" json:"created_by"`              // UserID of who set this (coach or client)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Client ClientProfile `gorm:"foreignKey:ClientID" json:"-"`
}

func (NutritionTarget) TableName() string {
	return "nutrition_targets"
}

// FoodItem - Local cache of food data from Open Food Facts or custom entries.
// Caching third-party data locally avoids repeated API calls (Open Food Facts: 100 req/min limit).
type FoodItem struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Name             string   `gorm:"not null;index" json:"name"`
	Brand            *string  `json:"brand"`
	ServingSize      *string  `json:"serving_size"`       // "1 cup", "100g"
	ServingSizeGrams *float64 `json:"serving_size_grams"` // normalized to grams for math

	// Nutritional values per serving
	Calories     *int     `json:"calories"`
	ProteinGrams *float64 `json:"protein_grams"`
	CarbsGrams   *float64 `json:"carbs_grams"`
	FatGrams     *float64 `json:"fat_grams"`
	FiberGrams   *float64 `json:"fiber_grams"`
	SugarGrams   *float64 `json:"sugar_grams"`
	SodiumMg     *float64 `json:"sodium_mg"`

	// Barcode for Open Food Facts lookups and scanning
	Barcode  *string `gorm:"index" json:"barcode"`
	ImageURL *string `json:"image_url"`

	// Content source - same pattern as Exercise model
	// "openfoodfacts", "chalk", "coach_custom", "client_custom"
	Source     string  `gorm:"not null;default:'chalk';index" json:"source"`
	ExternalID *string `gorm:"index" json:"external_id"` // Open Food Facts product ID for cache invalidation

	// System items are shared, custom items belong to a user
	IsSystem  bool  `gorm:"default:false;index" json:"is_system"`
	CreatedBy *uint `gorm:"index" json:"created_by"` // null for system/API items, set for user-created

	IsActive bool `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (FoodItem) TableName() string {
	return "food_items"
}

// FoodLogEntry - A logged food item with serving info for a specific date and meal.
// Stores computed macros at log time so edits to the FoodItem don't retroactively change history.
type FoodLogEntry struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	ClientID uint `gorm:"index;not null" json:"client_id"`

	FoodItemID uint    `gorm:"not null" json:"food_item_id"`
	LoggedDate string  `gorm:"type:date;not null;index" json:"logged_date"` // "2026-02-15"
	MealType   string  `gorm:"not null" json:"meal_type"`                   // "breakfast", "lunch", "dinner", "snack"
	Servings   float64 `gorm:"default:1" json:"servings"`

	// Snapshot of computed values at log time (servings * per-serving values)
	Calories     *int     `json:"calories"`
	ProteinGrams *float64 `json:"protein_grams"`
	CarbsGrams   *float64 `json:"carbs_grams"`
	FatGrams     *float64 `json:"fat_grams"`

	Notes *string `gorm:"type:text" json:"notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Client   ClientProfile `gorm:"foreignKey:ClientID" json:"-"`
	FoodItem FoodItem      `gorm:"foreignKey:FoodItemID" json:"food_item,omitempty"`
}

func (FoodLogEntry) TableName() string {
	return "food_log_entries"
}

// QuickMacroEntry - Manual macro/calorie entry without linking to a specific food item.
// Reduces friction for clients who just want to log "500 cal lunch" quickly.
type QuickMacroEntry struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	ClientID uint `gorm:"index;not null" json:"client_id"`

	LoggedDate  string  `gorm:"type:date;not null;index" json:"logged_date"`
	MealType    string  `gorm:"not null" json:"meal_type"` // "breakfast", "lunch", "dinner", "snack"
	Description *string `json:"description"`               // "Chipotle bowl", "Protein shake"

	Calories     *int     `json:"calories"`
	ProteinGrams *float64 `json:"protein_grams"`
	CarbsGrams   *float64 `json:"carbs_grams"`
	FatGrams     *float64 `json:"fat_grams"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Client ClientProfile `gorm:"foreignKey:ClientID" json:"-"`
}

func (QuickMacroEntry) TableName() string {
	return "quick_macro_entries"
}
