package models

import (
	"time"

	"gorm.io/gorm"
)

// User - Core user identity
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`

	// Email/Password auth (nullable for OAuth-only users)
	PasswordHash *string `gorm:"column:password_hash" json:"-"`

	// Email verification
	EmailVerified   bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`

	// Account status
	IsActive bool `gorm:"default:true" json:"is_active"`
	IsBanned bool `gorm:"default:false" json:"is_banned"`

	// Activity tracking
	LastLoginAt *time.Time `json:"last_login_at"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations (loaded with Preload)
	Profile        *Profile         `gorm:"foreignKey:UserID" json:"profile,omitempty"`
	CoachProfile   *CoachProfile    `gorm:"foreignKey:UserID" json:"coach_profile,omitempty"`
	ClientProfiles []ClientProfile  `gorm:"foreignKey:UserID" json:"client_profiles,omitempty"`
	Subscription   *Subscription    `gorm:"foreignKey:UserID" json:"subscription,omitempty"`
	OAuthProviders []OAuthProvider  `gorm:"foreignKey:UserID" json:"-"`
	RefreshTokens  []RefreshToken   `gorm:"foreignKey:UserID" json:"-"`
	DeviceTokens   []DeviceToken    `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// Profile - User profile information
type Profile struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	UserID    uint    `gorm:"uniqueIndex;not null" json:"user_id"`

	FirstName string  `gorm:"not null" json:"first_name"`
	LastName  string  `gorm:"not null" json:"last_name"`
	AvatarURL *string `json:"avatar_url"`
	Phone     *string `json:"phone"`
	Timezone  string  `gorm:"default:'UTC'" json:"timezone"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Profile) TableName() string {
	return "profiles"
}

// OAuthProvider - Third-party OAuth connections (Google, Facebook, Apple)
type OAuthProvider struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	UserID         uint   `gorm:"index;not null" json:"user_id"`
	Provider       string `gorm:"not null;index:idx_provider_user" json:"provider"` // "google", "facebook", "apple"
	ProviderUserID string `gorm:"not null;index:idx_provider_user" json:"-"`        // Their ID from provider

	// OAuth tokens (store encrypted in production)
	AccessToken    *string    `json:"-"`
	RefreshToken   *string    `json:"-"`
	TokenExpiresAt *time.Time `json:"-"`

	// Provider profile data (cached from OAuth provider)
	ProviderEmail  *string `json:"provider_email"` // Email from provider (may differ from user.email)
	ProviderName   *string `json:"provider_name"`
	ProviderAvatar *string `json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (OAuthProvider) TableName() string {
	return "oauth_providers"
}

// RefreshToken - JWT refresh tokens for session management
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null;size:512" json:"-"` // Hashed token
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	Revoked   bool      `gorm:"default:false;index" json:"revoked"`
	RevokedAt *time.Time `json:"revoked_at"`

	// Device/session tracking
	DeviceInfo *string `gorm:"type:text" json:"device_info"` // User agent
	IPAddress  *string `json:"ip_address"`

	// Last used (for cleanup of stale tokens)
	LastUsedAt *time.Time `json:"last_used_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// DeviceToken - Push notification tokens (Expo)
type DeviceToken struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	UserID   uint   `gorm:"index;not null" json:"user_id"`
	Token    string `gorm:"uniqueIndex;not null;size:512" json:"-"` // Expo push token
	Platform string `gorm:"not null" json:"platform"`               // "ios", "android"
	IsActive bool   `gorm:"default:true;index" json:"is_active"`

	// Device metadata
	DeviceName *string `json:"device_name"` // "John's iPhone"
	AppVersion *string `json:"app_version"` // "1.2.3"
	OSVersion  *string `json:"os_version"`  // "iOS 17.2"

	// Activity tracking
	LastUsedAt *time.Time `json:"last_used_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (DeviceToken) TableName() string {
	return "device_tokens"
}

// PasswordReset - Password reset request tokens
type PasswordReset struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"index;not null" json:"email"` // Not FK - works for non-existent users too
	Token     string    `gorm:"uniqueIndex;not null;size:512" json:"-"` // Hashed token
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	Used      bool      `gorm:"default:false;index" json:"used"`
	UsedAt    *time.Time `json:"used_at"`

	// Security tracking
	IPAddress *string `json:"ip_address"`
	UserAgent *string `gorm:"type:text" json:"-"`

	CreatedAt time.Time `json:"created_at"`
}

func (PasswordReset) TableName() string {
	return "password_resets"
}

// EmailVerification - Email verification tokens
type EmailVerification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"index;not null" json:"email"`
	Token     string    `gorm:"uniqueIndex;not null;size:512" json:"-"` // Hashed token
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	Used      bool      `gorm:"default:false;index" json:"used"`
	UsedAt    *time.Time `json:"used_at"`

	CreatedAt time.Time `json:"created_at"`
}

func (EmailVerification) TableName() string {
	return "email_verifications"
}

// MagicLink - Passwordless login tokens (future use)
type MagicLink struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"index;not null" json:"email"`
	Token     string    `gorm:"uniqueIndex;not null;size:512" json:"-"` // Hashed token
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	Used      bool      `gorm:"default:false;index" json:"used"`
	UsedAt    *time.Time `json:"used_at"`

	// Device tracking
	IPAddress *string `json:"ip_address"`
	UserAgent *string `gorm:"type:text" json:"-"`

	CreatedAt time.Time `json:"created_at"`
}

func (MagicLink) TableName() string {
	return "magic_links"
}
