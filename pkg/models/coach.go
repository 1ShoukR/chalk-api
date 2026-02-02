package models

import "time"

// SocialLinks - Flexible social media links structure
type SocialLinks struct {
	Instagram string            `json:"instagram,omitempty"`
	YouTube   string            `json:"youtube,omitempty"`
	TikTok    string            `json:"tiktok,omitempty"`
	Website   string            `json:"website,omitempty"`
	LinkedIn  string            `json:"linkedin,omitempty"`
	Facebook  string            `json:"facebook,omitempty"`
	Twitter   string            `json:"twitter,omitempty"`
	Other     map[string]string `json:"other,omitempty"` // Catch-all for future platforms
}

// CoachProfile - Coach-specific profile data
type CoachProfile struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"uniqueIndex;not null" json:"user_id"`

	// Business Info
	BusinessName *string `json:"business_name"`
	Bio          *string `gorm:"type:text" json:"bio"`
	CoverPhotoURL *string `json:"cover_photo_url"`

	// Expertise
	Specialties      []string `gorm:"type:text[];serializer:json" json:"specialties"` // ["strength", "weight loss", "bodybuilding"]
	YearsExperience  *int     `json:"years_experience"`
	Languages        []string `gorm:"type:text[];serializer:json" json:"languages"` // ["English", "Spanish"]

	// Service Details
	TrainingType string `gorm:"default:'hybrid'" json:"training_type"` // "in_person", "online", "hybrid"

	// Pricing (optional - coaches can choose to display)
	HourlyRate         *float64 `json:"hourly_rate"`
	HourlyRateCurrency string   `gorm:"default:'USD'" json:"hourly_rate_currency"`
	ShowRate           bool     `gorm:"default:false" json:"-"` // Privacy control

	// Social/Marketing
	SocialLinks SocialLinks `gorm:"type:jsonb;serializer:json" json:"social_links"`

	// Subscription (RevenueCat integration)
	SubscriptionTier      string     `gorm:"default:'free'" json:"subscription_tier"` // "free", "pro", "enterprise"
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at"`

	// Onboarding & Status
	OnboardingCompleted bool `gorm:"default:false" json:"onboarding_completed"`
	IsAcceptingClients  bool `gorm:"default:true" json:"is_accepting_clients"`

	// Activity
	LastActiveAt *time.Time `json:"last_active_at"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User           User            `gorm:"foreignKey:UserID" json:"-"`
	Certifications []Certification `gorm:"foreignKey:CoachID" json:"certifications,omitempty"`
	Locations      []CoachLocation `gorm:"foreignKey:CoachID" json:"locations,omitempty"`
	Stats          *CoachStats     `gorm:"foreignKey:CoachID" json:"stats,omitempty"`
	Clients        []ClientProfile `gorm:"foreignKey:CoachID" json:"clients,omitempty"`
}

func (CoachProfile) TableName() string {
	return "coach_profiles"
}

// Certification - Coach certifications with document upload
type Certification struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	CoachID uint `gorm:"index;not null" json:"coach_id"`

	Name        string  `gorm:"not null" json:"name"`               // "NASM-CPT"
	IssuingOrg  string  `json:"issuing_org"`                        // "National Academy of Sports Medicine"
	Description *string `gorm:"type:text" json:"description"`

	// Document upload
	CertificateURL *string `json:"certificate_url"` // S3/R2 link to PDF/image

	// Validity
	IssuedDate *string `gorm:"type:date" json:"issued_date"` // "2022-01-15"
	ExpiryDate *string `gorm:"type:date" json:"expiry_date"` // "2025-01-15"
	IsVerified bool    `gorm:"default:false" json:"is_verified"` // Admin verification

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Coach CoachProfile `gorm:"foreignKey:CoachID" json:"-"`
}

func (Certification) TableName() string {
	return "certifications"
}

// CoachLocation - Training locations (coaches can have multiple)
type CoachLocation struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	CoachID uint `gorm:"index;not null" json:"coach_id"`

	Name string `gorm:"not null" json:"name"` // "Gold's Gym Downtown"
	Type string `gorm:"not null" json:"type"` // "gym", "studio", "outdoor", "home", "online"

	// Address
	Address *string `json:"address"`
	City    *string `json:"city"`
	State   *string `json:"state"`
	ZipCode *string `json:"zip_code"`
	Country string  `gorm:"default:'US'" json:"country"`

	// Geolocation (for "find coaches near me" feature)
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`

	// Details
	IsPrimary bool    `gorm:"default:false;index" json:"is_primary"` // Main location
	IsActive  bool    `gorm:"default:true" json:"is_active"`
	Notes     *string `gorm:"type:text" json:"notes"` // "Access through side entrance"

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Coach CoachProfile `gorm:"foreignKey:CoachID" json:"-"`
}

func (CoachLocation) TableName() string {
	return "coach_locations"
}

// CoachStats - Analytics and metrics (updated by background jobs)
type CoachStats struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	CoachID uint `gorm:"uniqueIndex;not null" json:"coach_id"`

	// Client metrics
	ActiveClients       int `gorm:"default:0" json:"active_clients"`
	TotalClientsAllTime int `gorm:"default:0" json:"total_clients_all_time"`
	ClientsThisMonth    int `gorm:"default:0" json:"clients_this_month"`

	// Workout metrics
	WorkoutsAssignedTotal  int `gorm:"default:0" json:"workouts_assigned_total"`
	WorkoutsCompletedTotal int `gorm:"default:0" json:"workouts_completed_total"`
	WorkoutsThisWeek       int `gorm:"default:0" json:"workouts_this_week"`

	// Session metrics
	SessionsCompletedTotal int `gorm:"default:0" json:"sessions_completed_total"`
	SessionsThisMonth      int `gorm:"default:0" json:"sessions_this_month"`

	// Engagement
	MessagesThisWeek       int  `gorm:"default:0" json:"messages_this_week"`
	AvgResponseTimeMinutes *int `json:"avg_response_time_minutes"`

	// Revenue tracking (future)
	TotalRevenueThisMonth *float64 `json:"total_revenue_this_month"`

	UpdatedAt time.Time `json:"updated_at"`

	Coach CoachProfile `gorm:"foreignKey:CoachID" json:"-"`
}

func (CoachStats) TableName() string {
	return "coach_stats"
}
