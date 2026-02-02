package models

import "time"

// ClientProfile - Relationship between a user (client) and their coach
type ClientProfile struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	UserID  uint `gorm:"index;not null" json:"user_id"` // The client
	CoachID uint `gorm:"index;not null" json:"coach_id"` // Their coach

	// Relationship - Auto-approved when using invite code
	Status string `gorm:"default:'active'" json:"status"` // "active", "paused", "archived"

	// Program Details (set by coach)
	Goals           *string `gorm:"type:text" json:"goals"`
	ProgramType     *string `json:"program_type"` // "strength", "weight_loss", "general_fitness"
	SessionsPerWeek *int    `json:"sessions_per_week"`

	// Organization (coach-only)
	Tags         []string `gorm:"type:text[];serializer:json" json:"tags"` // ["priority", "beginner"]
	PrivateNotes *string  `gorm:"type:text" json:"-"`                      // NEVER sent to client

	// Tracking
	LastContactAt *time.Time `json:"last_contact_at"` // Last message/session

	// Timestamps
	InvitedAt *time.Time `json:"invited_at"` // When coach created the invite
	JoinedAt  *time.Time `json:"joined_at"`  // When client accepted invite

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User       User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Coach      CoachProfile      `gorm:"foreignKey:CoachID" json:"coach,omitempty"`
	IntakeForm *ClientIntakeForm `gorm:"foreignKey:ClientID" json:"intake_form,omitempty"`
}

func (ClientProfile) TableName() string {
	return "client_profiles"
}

// InviteCode - Coach invitation system with unique codes
type InviteCode struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	CoachID uint   `gorm:"index;not null" json:"coach_id"`
	Code    string `gorm:"uniqueIndex;not null;size:20" json:"code"` // URL-safe code e.g., "ABC123XYZ"

	// Expiration (always set, e.g., 7 days from creation)
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`

	// Usage tracking (single-use)
	UsedBy *uint      `gorm:"index" json:"used_by"` // Which UserID used it (null if unused)
	UsedAt *time.Time `json:"used_at"`

	// Status
	IsActive bool `gorm:"default:true;index" json:"is_active"` // Coach can manually deactivate

	CreatedAt time.Time `json:"created_at"`

	// Relations
	Coach CoachProfile `gorm:"foreignKey:CoachID" json:"-"`
	User  *User        `gorm:"foreignKey:UsedBy" json:"used_by_user,omitempty"`
}

func (InviteCode) TableName() string {
	return "invite_codes"
}

// ClientIntakeForm - Initial client assessment filled out once when joining a coach
type ClientIntakeForm struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	ClientID uint `gorm:"uniqueIndex;not null" json:"client_id"` // FK to ClientProfile

	// Fitness Background
	FitnessLevel       string  `json:"fitness_level"` // "beginner", "intermediate", "advanced"
	YearsTraining      *int    `json:"years_training"`
	PreviousExperience *string `gorm:"type:text" json:"previous_experience"` // "Played football in high school..."

	// Goals & Motivation
	PrimaryGoal     string  `json:"primary_goal"` // "weight_loss", "muscle_gain", "strength", "athletic_performance", "general_fitness"
	SpecificGoals   *string `gorm:"type:text" json:"specific_goals"` // Free text details
	MotivationLevel *int    `json:"motivation_level"`                // 1-10 scale
	WhyHireCoach    *string `gorm:"type:text" json:"why_hire_coach"` // "Accountability, expert guidance..."

	// Limitations & Health
	Injuries         *string `gorm:"type:text" json:"injuries"`          // "Bad left knee, shoulder surgery 2020"
	HealthConditions *string `gorm:"type:text" json:"health_conditions"` // "Asthma, high blood pressure"
	Medications      *string `gorm:"type:text" json:"medications"`
	DoctorClearance  bool    `gorm:"default:false" json:"doctor_clearance"` // Cleared to train?

	// Availability & Preferences
	AvailableDays      []string `gorm:"type:text[];serializer:json" json:"available_days"` // ["Monday", "Wednesday", "Friday"]
	PreferredTimeOfDay string   `json:"preferred_time_of_day"`                             // "morning", "afternoon", "evening", "flexible"
	SessionDuration    *int     `json:"session_duration"`                                  // Preferred minutes per session

	// Equipment & Location
	TrainingLocation   string  `json:"training_location"`                       // "gym", "home", "outdoor", "flexible"
	EquipmentAvailable *string `gorm:"type:text" json:"equipment_available"`    // "Dumbbells, resistance bands, pull-up bar"
	GymMembership      *string `json:"gym_membership"`                          // Which gym they belong to

	// Lifestyle
	OccupationType     *string `json:"occupation_type"`                      // "sedentary", "active", "very_active"
	SleepHours         *int    `json:"sleep_hours"`                          // Average per night
	StressLevel        *int    `json:"stress_level"`                         // 1-10 scale
	DietaryPreferences *string `gorm:"type:text" json:"dietary_preferences"` // "Vegetarian, allergic to nuts"

	// Additional Notes
	AdditionalInfo *string `gorm:"type:text" json:"additional_info"` // Anything else client wants to share

	// Completion
	CompletedAt *time.Time `json:"completed_at"` // When client submitted the form

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	ClientProfile ClientProfile `gorm:"foreignKey:ClientID" json:"-"`
}

func (ClientIntakeForm) TableName() string {
	return "client_intake_forms"
}
