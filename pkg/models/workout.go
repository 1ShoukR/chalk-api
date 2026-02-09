package models

import "time"

// Workout - An assigned workout for a specific client on a specific date.
// Created as a deep copy from a template (if applicable) so template edits don't affect existing assignments.
type Workout struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	ClientID uint `gorm:"index;not null" json:"client_id"`
	CoachID  uint `gorm:"index;not null" json:"coach_id"`

	// Optional reference to source template (informational only, not a live link)
	TemplateID *uint `json:"template_id"`

	Name          string  `gorm:"not null" json:"name"`
	Description   *string `gorm:"type:text" json:"description"`
	ScheduledDate *string `gorm:"type:date;index" json:"scheduled_date"` // "2026-02-15"

	// Status flow: scheduled → in_progress → completed / skipped
	Status      string     `gorm:"default:'scheduled';index" json:"status"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`

	// Notes from both sides
	ClientNotes *string `gorm:"type:text" json:"client_notes"`
	CoachNotes  *string `gorm:"type:text" json:"coach_notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Client    ClientProfile     `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	Coach     CoachProfile      `gorm:"foreignKey:CoachID" json:"-"`
	Template  *WorkoutTemplate  `gorm:"foreignKey:TemplateID" json:"-"`
	Exercises []WorkoutExercise `gorm:"foreignKey:WorkoutID" json:"exercises,omitempty"`
}

func (Workout) TableName() string {
	return "workouts"
}

// WorkoutExercise - Exercise within an assigned workout with completion tracking.
// Mirrors template exercise structure but adds per-exercise completion status.
type WorkoutExercise struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	WorkoutID  uint `gorm:"index;not null" json:"workout_id"`
	ExerciseID uint `gorm:"not null" json:"exercise_id"`

	// Ordering and grouping (copied from template at assignment time)
	OrderIndex   int     `gorm:"not null" json:"order_index"`
	SectionLabel *string `json:"section_label"`

	SupersetGroup *int    `json:"superset_group"`
	GroupType     *string `json:"group_type"`

	// Prescribed values (copied from template)
	Sets        *int     `json:"sets"`
	RepsMin     *int     `json:"reps_min"`
	RepsMax     *int     `json:"reps_max"`
	WeightValue *float64 `json:"weight_value"`
	WeightUnit  *string  `json:"weight_unit"`

	PrescriptionNote *string `gorm:"type:text" json:"prescription_note"`
	RestSeconds      *int    `json:"rest_seconds"`
	Tempo            *string `json:"tempo"`
	Notes            *string `gorm:"type:text" json:"notes"`

	// Per-exercise completion tracking so coaches see partial progress
	IsCompleted bool `gorm:"default:false;index" json:"is_completed"`
	SkippedReason *string `json:"skipped_reason"` // why client skipped this exercise

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Workout  Workout      `gorm:"foreignKey:WorkoutID" json:"-"`
	Exercise Exercise     `gorm:"foreignKey:ExerciseID" json:"exercise,omitempty"`
	Logs     []WorkoutLog `gorm:"foreignKey:WorkoutExerciseID" json:"logs,omitempty"`
}

func (WorkoutExercise) TableName() string {
	return "workout_exercises"
}

// WorkoutLog - Actual performance data for a single set.
// One row per set enables granular progress tracking and analytics.
type WorkoutLog struct {
	ID                uint `gorm:"primaryKey" json:"id"`
	WorkoutExerciseID uint `gorm:"index;not null" json:"workout_exercise_id"`

	SetNumber     int      `gorm:"not null" json:"set_number"`
	RepsCompleted *int     `json:"reps_completed"`
	WeightUsed    *float64 `json:"weight_used"`
	WeightUnit    *string  `json:"weight_unit"` // "lbs", "kg"

	// Subjective effort tracking
	RPE   *int    `json:"rpe"` // Rate of Perceived Exertion 1-10
	Notes *string `json:"notes"`

	// For time/distance-based exercises
	DurationSeconds *int     `json:"duration_seconds"`
	Distance        *float64 `json:"distance"`
	DistanceUnit    *string  `json:"distance_unit"` // "miles", "km", "meters"

	CreatedAt time.Time `json:"created_at"`

	WorkoutExercise WorkoutExercise `gorm:"foreignKey:WorkoutExerciseID" json:"-"`
}

func (WorkoutLog) TableName() string {
	return "workout_logs"
}
