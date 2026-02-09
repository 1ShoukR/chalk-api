package repositories

import (
	"chalk-api/pkg/models"
	"context"

	"gorm.io/gorm"
)

type NutritionRepository struct {
	db *gorm.DB
}

func NewNutritionRepository(db *gorm.DB) *NutritionRepository {
	return &NutritionRepository{db: db}
}

// --- Targets ---

func (r *NutritionRepository) CreateTarget(ctx context.Context, target *models.NutritionTarget) error {
	return r.db.WithContext(ctx).Create(target).Error
}

// GetCurrentTarget returns the most recent target by effective_date
func (r *NutritionRepository) GetCurrentTarget(ctx context.Context, clientID uint) (*models.NutritionTarget, error) {
	var target models.NutritionTarget
	err := r.db.WithContext(ctx).
		Where("client_id = ? AND effective_date <= CURRENT_DATE", clientID).
		Order("effective_date DESC").
		First(&target).Error
	if err != nil {
		return nil, err
	}
	return &target, nil
}

func (r *NutritionRepository) ListTargets(ctx context.Context, clientID uint) ([]models.NutritionTarget, error) {
	var targets []models.NutritionTarget
	err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("effective_date DESC").
		Find(&targets).Error
	return targets, err
}

// --- Food Items ---

func (r *NutritionRepository) CreateFoodItem(ctx context.Context, item *models.FoodItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *NutritionRepository) GetFoodItem(ctx context.Context, id uint) (*models.FoodItem, error) {
	var item models.FoodItem
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *NutritionRepository) SearchFoodItems(ctx context.Context, query string, limit, offset int) ([]models.FoodItem, int64, error) {
	var items []models.FoodItem
	var total int64

	dbQuery := r.db.WithContext(ctx).
		Where("is_active = ? AND name ILIKE ?", true, "%"+query+"%")

	if err := dbQuery.Model(&models.FoodItem{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := dbQuery.
		Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&items).Error

	return items, total, err
}

func (r *NutritionRepository) GetByBarcode(ctx context.Context, barcode string) (*models.FoodItem, error) {
	var item models.FoodItem
	err := r.db.WithContext(ctx).
		Where("barcode = ?", barcode).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetByExternalID checks if we already cached a food item from Open Food Facts
func (r *NutritionRepository) GetByExternalID(ctx context.Context, source, externalID string) (*models.FoodItem, error) {
	var item models.FoodItem
	err := r.db.WithContext(ctx).
		Where("source = ? AND external_id = ?", source, externalID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// --- Food Logs ---

func (r *NutritionRepository) CreateFoodLog(ctx context.Context, entry *models.FoodLogEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *NutritionRepository) UpdateFoodLog(ctx context.Context, entry *models.FoodLogEntry) error {
	return r.db.WithContext(ctx).Save(entry).Error
}

func (r *NutritionRepository) DeleteFoodLog(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.FoodLogEntry{}, id).Error
}

func (r *NutritionRepository) ListFoodLogs(ctx context.Context, clientID uint, date string) ([]models.FoodLogEntry, error) {
	var entries []models.FoodLogEntry
	err := r.db.WithContext(ctx).
		Preload("FoodItem").
		Where("client_id = ? AND logged_date = ?", clientID, date).
		Order("created_at ASC").
		Find(&entries).Error
	return entries, err
}

// DailySummary holds aggregated macros for a single day
type DailySummary struct {
	Calories     int     `json:"calories"`
	ProteinGrams float64 `json:"protein_grams"`
	CarbsGrams   float64 `json:"carbs_grams"`
	FatGrams     float64 `json:"fat_grams"`
}

// GetDailySummary aggregates all food logs and quick macros for a day in a single query
func (r *NutritionRepository) GetDailySummary(ctx context.Context, clientID uint, date string) (*DailySummary, error) {
	var summary DailySummary

	// Aggregate food log entries
	var foodSummary DailySummary
	r.db.WithContext(ctx).
		Model(&models.FoodLogEntry{}).
		Select("COALESCE(SUM(calories), 0) as calories, COALESCE(SUM(protein_grams), 0) as protein_grams, COALESCE(SUM(carbs_grams), 0) as carbs_grams, COALESCE(SUM(fat_grams), 0) as fat_grams").
		Where("client_id = ? AND logged_date = ?", clientID, date).
		Scan(&foodSummary)

	// Aggregate quick macro entries
	var quickSummary DailySummary
	r.db.WithContext(ctx).
		Model(&models.QuickMacroEntry{}).
		Select("COALESCE(SUM(calories), 0) as calories, COALESCE(SUM(protein_grams), 0) as protein_grams, COALESCE(SUM(carbs_grams), 0) as carbs_grams, COALESCE(SUM(fat_grams), 0) as fat_grams").
		Where("client_id = ? AND logged_date = ?", clientID, date).
		Scan(&quickSummary)

	// Combine both sources
	summary.Calories = foodSummary.Calories + quickSummary.Calories
	summary.ProteinGrams = foodSummary.ProteinGrams + quickSummary.ProteinGrams
	summary.CarbsGrams = foodSummary.CarbsGrams + quickSummary.CarbsGrams
	summary.FatGrams = foodSummary.FatGrams + quickSummary.FatGrams

	return &summary, nil
}

// --- Quick Macros ---

func (r *NutritionRepository) CreateQuickMacro(ctx context.Context, entry *models.QuickMacroEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *NutritionRepository) UpdateQuickMacro(ctx context.Context, entry *models.QuickMacroEntry) error {
	return r.db.WithContext(ctx).Save(entry).Error
}

func (r *NutritionRepository) DeleteQuickMacro(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.QuickMacroEntry{}, id).Error
}

func (r *NutritionRepository) ListQuickMacros(ctx context.Context, clientID uint, date string) ([]models.QuickMacroEntry, error) {
	var entries []models.QuickMacroEntry
	err := r.db.WithContext(ctx).
		Where("client_id = ? AND logged_date = ?", clientID, date).
		Order("created_at ASC").
		Find(&entries).Error
	return entries, err
}
