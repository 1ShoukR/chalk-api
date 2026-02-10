package services

import (
	"chalk-api/pkg/models"
	"chalk-api/pkg/repositories"
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type UpdateMeInput struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Phone     *string `json:"phone"`
	AvatarURL *string `json:"avatar_url"`
	Timezone  *string `json:"timezone"`
}

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetMe(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateMe(ctx context.Context, userID uint, input UpdateMeInput) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.Profile == nil {
		return nil, ErrUserNotFound
	}

	if input.FirstName != nil {
		user.Profile.FirstName = strings.TrimSpace(*input.FirstName)
	}
	if input.LastName != nil {
		user.Profile.LastName = strings.TrimSpace(*input.LastName)
	}
	if input.Phone != nil {
		user.Profile.Phone = input.Phone
	}
	if input.AvatarURL != nil {
		user.Profile.AvatarURL = input.AvatarURL
	}
	if input.Timezone != nil && strings.TrimSpace(*input.Timezone) != "" {
		user.Profile.Timezone = strings.TrimSpace(*input.Timezone)
	}

	if err := s.userRepo.UpdateProfile(ctx, user.Profile); err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, userID)
}
