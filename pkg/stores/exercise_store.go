package stores

import (
	"chalk-api/pkg/models"
	"time"
)

// ExerciseStore handles exercise library caching
// System exercises are cached aggressively since they rarely change
type ExerciseStore struct {
	redis *RedisClient
}

const (
	// System exercises rarely change, cache for 24 hours
	SystemExerciseTTL = 24 * time.Hour
	// Coach exercises might be updated more frequently
	CoachExerciseTTL  = 1 * time.Hour
	// Exercise list cache
	ExerciseListTTL   = 30 * time.Minute
)

// NewExerciseStore creates a new exercise store
func NewExerciseStore(redis *RedisClient) *ExerciseStore {
	return &ExerciseStore{redis: redis}
}

// CachedExercise is a lightweight cache representation
// Mirrors model pointer types to avoid unnecessary conversions
type CachedExercise struct {
	ID                    uint     `json:"id"`
	Name                  string   `json:"name"`
	Description           *string  `json:"description,omitempty"`
	Instructions          *string  `json:"instructions,omitempty"`
	GifURL                *string  `json:"gif_url,omitempty"`
	VideoURL              *string  `json:"video_url,omitempty"`
	ThumbnailURL          *string  `json:"thumbnail_url,omitempty"`
	PrimaryMuscleGroups   []string `json:"primary_muscle_groups,omitempty"`
	SecondaryMuscleGroups []string `json:"secondary_muscle_groups,omitempty"`
	PrimaryEquipment      []string `json:"primary_equipment,omitempty"`
	Difficulty            *string  `json:"difficulty,omitempty"`
	MeasurementType       string   `json:"measurement_type"`
	Tags                  []string `json:"tags,omitempty"`
	IsSystem              bool     `json:"is_system"`
	Source                string   `json:"source,omitempty"`
	ExternalID            *string  `json:"external_id,omitempty"`
}

// ToCachedExercise converts a models.Exercise to cached version
func ToCachedExercise(e *models.Exercise) *CachedExercise {
	if e == nil {
		return nil
	}
	return &CachedExercise{
		ID:                    e.ID,
		Name:                  e.Name,
		Description:           e.Description,
		Instructions:          e.Instructions,
		GifURL:                e.GifURL,
		VideoURL:              e.VideoURL,
		ThumbnailURL:          e.ThumbnailURL,
		PrimaryMuscleGroups:   e.PrimaryMuscleGroups,
		SecondaryMuscleGroups: e.SecondaryMuscleGroups,
		PrimaryEquipment:      e.PrimaryEquipment,
		Difficulty:            e.Difficulty,
		MeasurementType:       e.MeasurementType,
		Tags:                  e.Tags,
		IsSystem:              e.IsSystem,
		Source:                e.Source,
		ExternalID:            e.ExternalID,
	}
}

// Get retrieves a cached exercise by ID
func (s *ExerciseStore) Get(exerciseID uint) (*CachedExercise, bool) {
	if !s.redis.IsAvailable() {
		return nil, false
	}

	var exercise CachedExercise
	if s.redis.GetJSON(KeyExercise(exerciseID), &exercise) {
		return &exercise, true
	}
	return nil, false
}

// Set caches an exercise
func (s *ExerciseStore) Set(exercise *models.Exercise) {
	if !s.redis.IsAvailable() || exercise == nil {
		return
	}

	ttl := CoachExerciseTTL
	if exercise.IsSystem {
		ttl = SystemExerciseTTL
	}

	cached := ToCachedExercise(exercise)
	s.redis.SetJSON(KeyExercise(exercise.ID), cached, ttl)
}

// SetMany caches multiple exercises (useful for bulk operations)
func (s *ExerciseStore) SetMany(exercises []models.Exercise) {
	if !s.redis.IsAvailable() {
		return
	}

	for i := range exercises {
		s.Set(&exercises[i])
	}
}

// GetList retrieves a cached exercise list
func (s *ExerciseStore) GetList(coachID uint, page int) ([]CachedExercise, bool) {
	if !s.redis.IsAvailable() {
		return nil, false
	}

	var exercises []CachedExercise
	if s.redis.GetJSON(KeyExerciseList(coachID, page), &exercises) {
		return exercises, true
	}
	return nil, false
}

// SetList caches an exercise list
func (s *ExerciseStore) SetList(coachID uint, page int, exercises []models.Exercise) {
	if !s.redis.IsAvailable() {
		return
	}

	cached := make([]CachedExercise, len(exercises))
	for i := range exercises {
		cached[i] = *ToCachedExercise(&exercises[i])
	}

	s.redis.SetJSON(KeyExerciseList(coachID, page), cached, ExerciseListTTL)
}

// GetSystemList retrieves cached system exercises
func (s *ExerciseStore) GetSystemList(page int) ([]CachedExercise, bool) {
	if !s.redis.IsAvailable() {
		return nil, false
	}

	var exercises []CachedExercise
	if s.redis.GetJSON(KeySystemExercises(page), &exercises) {
		return exercises, true
	}
	return nil, false
}

// SetSystemList caches system exercises
func (s *ExerciseStore) SetSystemList(page int, exercises []models.Exercise) {
	if !s.redis.IsAvailable() {
		return
	}

	cached := make([]CachedExercise, len(exercises))
	for i := range exercises {
		cached[i] = *ToCachedExercise(&exercises[i])
	}

	s.redis.SetJSON(KeySystemExercises(page), cached, SystemExerciseTTL)
}

// Invalidate removes an exercise from cache
func (s *ExerciseStore) Invalidate(exerciseID uint) {
	if s.redis.IsAvailable() {
		s.redis.Delete(KeyExercise(exerciseID))
	}
}

// InvalidateLists removes all exercise list caches for a coach
// Call this when exercises are added/removed/modified
func (s *ExerciseStore) InvalidateLists(coachID uint) {
	if s.redis.IsAvailable() {
		// Delete pattern for coach's exercise lists
		s.redis.DeletePattern(KeyExerciseList(coachID, 0)[:len(KeyExerciseList(coachID, 0))-1] + "*")
	}
}

// InvalidateSystemLists removes all system exercise list caches
func (s *ExerciseStore) InvalidateSystemLists() {
	if s.redis.IsAvailable() {
		s.redis.DeletePattern("exercise:system:*")
	}
}
