package models

import "time"

// CoachAvailability - Recurring weekly availability slots.
// Business logic computes bookable time slots from these ranges based on session duration.
// All times stored in UTC; convert using coach's timezone from Profile.
type CoachAvailability struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	CoachID uint `gorm:"index;not null" json:"coach_id"`

	DayOfWeek int    `gorm:"not null" json:"day_of_week"` // 0=Sunday, 6=Saturday
	StartTime string `gorm:"not null" json:"start_time"`  // "09:00" (UTC)
	EndTime   string `gorm:"not null" json:"end_time"`    // "17:00" (UTC)
	IsActive  bool   `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Coach CoachProfile `gorm:"foreignKey:CoachID" json:"-"`
}

func (CoachAvailability) TableName() string {
	return "coach_availabilities"
}

// CoachAvailabilityOverride - Date-specific exceptions to recurring availability.
// Used to block off days (vacation) or add extra availability (working a Saturday).
type CoachAvailabilityOverride struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	CoachID uint `gorm:"index;not null" json:"coach_id"`

	Date        string `gorm:"type:date;not null;index" json:"date"` // "2026-03-15"
	IsAvailable bool   `gorm:"default:false" json:"is_available"`    // false = blocked off, true = extra availability

	// Only needed when adding extra availability (IsAvailable=true)
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`

	Reason *string `json:"reason"` // "Vacation", "Holiday", "Special event"

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Coach CoachProfile `gorm:"foreignKey:CoachID" json:"-"`
}

func (CoachAvailabilityOverride) TableName() string {
	return "coach_availability_overrides"
}

// SessionType - Types of sessions a coach offers with defined durations.
// Enables future per-type pricing and consistent booking experience.
type SessionType struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	CoachID uint `gorm:"index;not null" json:"coach_id"`

	Name            string  `gorm:"not null" json:"name"` // "1-on-1 Training", "Quick Check-in"
	DurationMinutes int     `gorm:"not null" json:"duration_minutes"`
	Description     *string `gorm:"type:text" json:"description"`
	Color           *string `json:"color"`                       // hex color for calendar display
	IsActive        bool    `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Coach CoachProfile `gorm:"foreignKey:CoachID" json:"-"`
}

func (SessionType) TableName() string {
	return "session_types"
}

// Session - A booked session between a coach and client.
// Tracks full lifecycle from scheduled through completion or cancellation.
type Session struct {
	ID            uint `gorm:"primaryKey" json:"id"`
	CoachID       uint `gorm:"index;not null" json:"coach_id"`
	ClientID      uint `gorm:"index;not null" json:"client_id"`
	SessionTypeID uint `gorm:"not null" json:"session_type_id"`

	ScheduledAt     time.Time `gorm:"not null;index" json:"scheduled_at"` // UTC
	DurationMinutes int       `gorm:"not null" json:"duration_minutes"`

	// Status flow: scheduled → completed / cancelled / no_show
	Status   string  `gorm:"default:'scheduled';index" json:"status"`
	Location *string `json:"location"`
	Notes    *string `gorm:"type:text" json:"notes"`

	// Cancellation tracking - who cancelled and why
	CancelledAt        *time.Time `json:"cancelled_at"`
	CancelledBy        *string    `json:"cancelled_by"`         // "coach" or "client"
	CancellationReason *string    `gorm:"type:text" json:"cancellation_reason"`

	CompletedAt *time.Time `json:"completed_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Coach       CoachProfile  `gorm:"foreignKey:CoachID" json:"coach,omitempty"`
	Client      ClientProfile `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	SessionType SessionType   `gorm:"foreignKey:SessionTypeID" json:"session_type,omitempty"`
}

func (Session) TableName() string {
	return "sessions"
}
