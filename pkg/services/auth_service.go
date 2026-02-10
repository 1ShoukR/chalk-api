package services

import (
	"chalk-api/pkg/models"
	"chalk-api/pkg/repositories"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserDisabled       = errors.New("user account is inactive or banned")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
)

type RegisterInput struct {
	Email     string  `json:"email" binding:"required,email"`
	Password  string  `json:"password" binding:"required,min=8"`
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name" binding:"required"`
	Phone     *string `json:"phone"`
	Timezone  string  `json:"timezone"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutInput struct {
	RefreshToken string `json:"refresh_token"`
	AllDevices   bool   `json:"all_devices"`
}

type AuthResult struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         *models.User `json:"user"`
}

type accessTokenClaims struct {
	UserID uint   `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo        *repositories.UserRepository
	authRepo        *repositories.AuthRepository
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(
	userRepo *repositories.UserRepository,
	authRepo *repositories.AuthRepository,
	jwtSecret string,
	jwtExpirationHours int,
) *AuthService {
	accessHours := jwtExpirationHours
	if accessHours <= 0 {
		accessHours = 24
	}

	return &AuthService{
		userRepo:       userRepo,
		authRepo:       authRepo,
		jwtSecret:      []byte(jwtSecret),
		accessTokenTTL: time.Duration(accessHours) * time.Hour,
		// Keep refresh tokens longer than access tokens for mobile/web session continuity.
		refreshTokenTTL: 30 * 24 * time.Hour,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput, userAgent, ipAddress string) (*AuthResult, error) {
	email := normalizeEmail(input.Email)
	if email == "" {
		return nil, ErrInvalidCredentials
	}

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	passwordHashStr := string(passwordHash)

	timezone := strings.TrimSpace(input.Timezone)
	if timezone == "" {
		timezone = "UTC"
	}

	user := &models.User{
		Email:        email,
		PasswordHash: &passwordHashStr,
		IsActive:     true,
		IsBanned:     false,
	}

	profile := &models.Profile{
		FirstName: strings.TrimSpace(input.FirstName),
		LastName:  strings.TrimSpace(input.LastName),
		Phone:     input.Phone,
		Timezone:  timezone,
	}

	if err := s.userRepo.Create(ctx, user, profile); err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	freshUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, freshUser, userAgent, ipAddress)
}

func (s *AuthService) Login(ctx context.Context, input LoginInput, userAgent, ipAddress string) (*AuthResult, error) {
	email := normalizeEmail(input.Email)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if user.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive || user.IsBanned {
		return nil, ErrUserDisabled
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	updatedUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, updatedUser, userAgent, ipAddress)
}

func (s *AuthService) Refresh(ctx context.Context, input RefreshInput, userAgent, ipAddress string) (*AuthResult, error) {
	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidRefresh
	}

	tokenHash := hashRefreshToken(refreshToken)
	storedToken, err := s.authRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidRefresh
		}
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive || user.IsBanned {
		return nil, ErrUserDisabled
	}

	if err := s.authRepo.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, userAgent, ipAddress)
}

func (s *AuthService) Logout(ctx context.Context, userID uint, input LogoutInput) error {
	if input.AllDevices || strings.TrimSpace(input.RefreshToken) == "" {
		return s.authRepo.RevokeAllUserTokens(ctx, userID)
	}

	tokenHash := hashRefreshToken(input.RefreshToken)
	token, err := s.authRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if token.UserID != userID {
		return ErrInvalidRefresh
	}

	return s.authRepo.RevokeRefreshToken(ctx, token.ID)
}

func (s *AuthService) issueTokens(ctx context.Context, user *models.User, userAgent, ipAddress string) (*AuthResult, error) {
	accessToken, expiresAt, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	tokenHash := hashRefreshToken(refreshToken)

	var deviceInfo *string
	if strings.TrimSpace(userAgent) != "" {
		ua := strings.TrimSpace(userAgent)
		deviceInfo = &ua
	}

	var ip *string
	if strings.TrimSpace(ipAddress) != "" {
		trimmed := strings.TrimSpace(ipAddress)
		ip = &trimmed
	}

	dbToken := &models.RefreshToken{
		UserID:     user.ID,
		Token:      tokenHash,
		ExpiresAt:  time.Now().UTC().Add(s.refreshTokenTTL),
		DeviceInfo: deviceInfo,
		IPAddress:  ip,
	}
	if err := s.authRepo.CreateRefreshToken(ctx, dbToken); err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    expiresAt,
		User:         user,
	}, nil
}

func (s *AuthService) generateAccessToken(user *models.User) (string, time.Time, error) {
	if len(s.jwtSecret) == 0 {
		return "", time.Time{}, fmt.Errorf("JWT_SECRET is not configured")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTokenTTL)

	jti, err := generateRefreshToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate token id: %w", err)
	}

	claims := accessTokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signedToken, expiresAt, nil
}

func ValidateAccessToken(tokenString string, jwtSecret string) (uint, error) {
	if strings.TrimSpace(tokenString) == "" {
		return 0, ErrInvalidCredentials
	}

	claims := &accessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || token == nil || !token.Valid {
		return 0, ErrInvalidCredentials
	}

	if claims.UserID == 0 {
		return 0, ErrInvalidCredentials
	}

	return claims.UserID, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func generateRefreshToken() (string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(random), nil
}

func hashRefreshToken(rawToken string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawToken)))
	return hex.EncodeToString(sum[:])
}
