package models

import "time"

// BodyMetric - Tracks client measurements over time using a narrow type+value pattern.
// New metric types can be added without migrations (just use a new MetricType string).
type BodyMetric struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	ClientID uint `gorm:"index;not null" json:"client_id"`

	MetricType string  `gorm:"not null;index" json:"metric_type"` // "weight", "body_fat", "waist", "chest", "hips", "bicep", "thigh", etc.
	Value      float64 `gorm:"not null" json:"value"`
	Unit       *string `json:"unit"` // "lbs", "kg", "%", "inches", "cm"

	RecordedAt time.Time `gorm:"not null;index" json:"recorded_at"`
	Notes      *string   `json:"notes"`

	CreatedAt time.Time `json:"created_at"`

	Client ClientProfile `gorm:"foreignKey:ClientID" json:"-"`
}

func (BodyMetric) TableName() string {
	return "body_metrics"
}

// ProgressPhoto - Client progress photos stored in S3/Railway volume.
// Tagged with photo type for UI grouping; group by date for before/after comparisons.
type ProgressPhoto struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	ClientID uint `gorm:"index;not null" json:"client_id"`

	PhotoURL  string  `gorm:"not null" json:"photo_url"`
	PhotoType *string `json:"photo_type"` // "front", "side", "back", "other"
	Notes     *string `json:"notes"`

	TakenAt time.Time `gorm:"not null;index" json:"taken_at"`

	CreatedAt time.Time `json:"created_at"`

	Client ClientProfile `gorm:"foreignKey:ClientID" json:"-"`
}

func (ProgressPhoto) TableName() string {
	return "progress_photos"
}
