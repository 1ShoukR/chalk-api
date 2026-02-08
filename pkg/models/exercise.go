package models

import "time"

// Exercise - Fitness exercise database supporting system, third-party, and coach-custom exercises.
// Source field enables seamless migration from third-party APIs to our own content library.
type Exercise struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Name         string  `gorm:"not null;index" json:"name"`
	Description  *string `gorm:"type:text" json:"description"`
	Instructions *string `gorm:"type:text" json:"instructions"`

	// Visual aids - URLs can point to third-party CDNs now, our own S3 later
	GifURL       *string `json:"gif_url"`
	VideoURL     *string `json:"video_url"`
	ThumbnailURL *string `json:"thumbnail_url"`

	// Categorization
	Category            string   `gorm:"not null;index" json:"category"`                             // "strength", "cardio", "flexibility", "plyometric"
	PrimaryMuscleGroups []string `gorm:"type:text[];serializer:json;index" json:"primary_muscle_groups"` // main muscles targeted
	SecondaryMuscleGroups []string `gorm:"type:text[];serializer:json" json:"secondary_muscle_groups"`  // assisting muscles

	// Equipment - split into required vs nice-to-have
	PrimaryEquipment  []string `gorm:"type:text[];serializer:json;index" json:"primary_equipment"`
	OptionalEquipment []string `gorm:"type:text[];serializer:json" json:"optional_equipment"`

	Difficulty      *string `gorm:"index" json:"difficulty"`            // "beginner", "intermediate", "advanced"
	MeasurementType string  `gorm:"not null" json:"measurement_type"` // "reps", "time", "distance"

	// Coaching info for trainers
	CoachingCues   *string `gorm:"type:text" json:"coaching_cues"`
	CommonMistakes *string `gorm:"type:text" json:"common_mistakes"`

	// Related exercises for same muscle groups (exercise IDs)
	RelatedExercises []uint `gorm:"type:integer[];serializer:json" json:"related_exercises"`

	// Extra categorization for filtering
	Tags []string `gorm:"type:text[];serializer:json" json:"tags"` // ["compound", "push", "horizontal_press"]

	// Content source - enables switching from third-party to our own library
	// "exercisedb", "musclewiki", "chalk", "coach_custom"
	Source     string  `gorm:"not null;default:'chalk';index" json:"source"`
	ExternalID *string `gorm:"index" json:"external_id"` // ID from third-party API for syncing

	// Ownership - system exercises are visible to all, custom are coach-only
	IsSystem bool  `gorm:"default:false;index" json:"is_system"`
	CoachID  *uint `gorm:"index" json:"coach_id"`

	IsActive bool `gorm:"default:true;index" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Coach *CoachProfile `gorm:"foreignKey:CoachID" json:"coach,omitempty"`
}

func (Exercise) TableName() string {
	return "exercises"
}
