package models

import "time"

// WorkoutTemplate - Reusable workout blueprint that coaches create once and assign to multiple clients.
// When assigned, a copy is made as a Workout so edits to the template don't affect existing assignments.
type WorkoutTemplate struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	CoachID uint `gorm:"index;not null" json:"coach_id"`

	Name        string  `gorm:"not null" json:"name"`
	Description *string `gorm:"type:text" json:"description"`

	// Categorization for coach's template library
	Category *string  `json:"category"` // "upper_body", "lower_body", "full_body", "cardio", "recovery"
	Tags     []string `gorm:"type:text[];serializer:json" json:"tags"`

	// Estimated duration helps with scheduling
	EstimatedMinutes *int `json:"estimated_minutes"`

	IsActive bool `gorm:"default:true;index" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Coach     CoachProfile              `gorm:"foreignKey:CoachID" json:"-"`
	Exercises []WorkoutTemplateExercise `gorm:"foreignKey:TemplateID" json:"exercises,omitempty"`
}

func (WorkoutTemplate) TableName() string {
	return "workout_templates"
}

// WorkoutTemplateExercise - Individual exercise within a template with prescribed sets/reps/weight.
// Uses structured fields for progress tracking with a free-text fallback for unusual prescriptions.
type WorkoutTemplateExercise struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	TemplateID uint `gorm:"index;not null" json:"template_id"`
	ExerciseID uint `gorm:"not null" json:"exercise_id"`

	// Ordering and grouping
	OrderIndex   int     `gorm:"not null" json:"order_index"`
	SectionLabel *string `json:"section_label"` // "Warm-up", "Main Lift", "Accessories", "Cooldown" - null means ungrouped

	// Superset/circuit grouping - exercises sharing the same group number are performed together
	SupersetGroup *int    `json:"superset_group"`
	GroupType     *string `json:"group_type"` // "superset", "circuit", "drop_set"

	// Structured prescription for tracking and analytics
	Sets      *int     `json:"sets"`
	RepsMin   *int     `json:"reps_min"`
	RepsMax   *int     `json:"reps_max"` // null if exact reps (use reps_min only)
	WeightValue *float64 `json:"weight_value"`
	WeightUnit  *string  `json:"weight_unit"` // "lbs", "kg", "percent_1rm", "rpe", "bodyweight"

	// Free-text override for anything the structured fields can't capture
	PrescriptionNote *string `gorm:"type:text" json:"prescription_note"` // "AMRAP", "work up to heavy single", etc.

	RestSeconds *int    `json:"rest_seconds"`
	Tempo       *string `json:"tempo"` // "3-1-2-0" (eccentric-pause-concentric-pause)
	Notes       *string `gorm:"type:text" json:"notes"` // Coach notes specific to this exercise in this template

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Template WorkoutTemplate `gorm:"foreignKey:TemplateID" json:"-"`
	Exercise Exercise        `gorm:"foreignKey:ExerciseID" json:"exercise,omitempty"`
}

func (WorkoutTemplateExercise) TableName() string {
	return "workout_template_exercises"
}
