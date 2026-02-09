package openfoodfacts

// SearchResponse is the API response from Open Food Facts search endpoint
type SearchResponse struct {
	Count    int       `json:"count"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Products []Product `json:"products"`
}

// Product represents a food product from Open Food Facts
type Product struct {
	Code        string `json:"code"` // Barcode
	ProductName string `json:"product_name"`
	Brands      string `json:"brands"`
	ImageURL    string `json:"image_url"`

	// Serving info
	ServingSize     string  `json:"serving_size"`
	ServingQuantity float64 `json:"serving_quantity"` // in grams

	// Nutriments per 100g
	Nutriments Nutriments `json:"nutriments"`

	// Quality indicators
	NutriscoreGrade string `json:"nutriscore_grade"` // a, b, c, d, e
	NovaGroup       int    `json:"nova_group"`       // 1-4 (processing level)
}

// Nutriments contains nutritional values per 100g
type Nutriments struct {
	// Per 100g values
	EnergyKcal100g    float64 `json:"energy-kcal_100g"`
	Proteins100g      float64 `json:"proteins_100g"`
	Carbohydrates100g float64 `json:"carbohydrates_100g"`
	Fat100g           float64 `json:"fat_100g"`
	Fiber100g         float64 `json:"fiber_100g"`
	Sugars100g        float64 `json:"sugars_100g"`
	Salt100g          float64 `json:"salt_100g"`
	Sodium100g        float64 `json:"sodium_100g"`

	// Per serving values (if available)
	EnergyKcalServing    float64 `json:"energy-kcal_serving"`
	ProteinsServing      float64 `json:"proteins_serving"`
	CarbohydratesServing float64 `json:"carbohydrates_serving"`
	FatServing           float64 `json:"fat_serving"`
	FiberServing         float64 `json:"fiber_serving"`
	SugarsServing        float64 `json:"sugars_serving"`
}

// ProductResponse is the API response when fetching a single product
type ProductResponse struct {
	Status        int     `json:"status"` // 1 = found, 0 = not found
	StatusVerbose string  `json:"status_verbose"`
	Product       Product `json:"product"`
}
