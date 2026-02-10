package services

import (
	"chalk-api/pkg/events"
	"chalk-api/pkg/models"
	"chalk-api/pkg/repositories"
	"chalk-api/pkg/utils"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrCoachProfileNotFound = errors.New("coach profile not found")
	ErrInviteCodeNotFound   = errors.New("invite code not found")
	ErrInviteForbidden      = errors.New("invite does not belong to coach")
)

type UpsertCoachProfileInput struct {
	BusinessName        *string             `json:"business_name"`
	Bio                 *string             `json:"bio"`
	CoverPhotoURL       *string             `json:"cover_photo_url"`
	Specialties         *[]string           `json:"specialties"`
	YearsExperience     *int                `json:"years_experience"`
	Languages           *[]string           `json:"languages"`
	TrainingType        *string             `json:"training_type"`
	HourlyRate          *float64            `json:"hourly_rate"`
	HourlyRateCurrency  *string             `json:"hourly_rate_currency"`
	ShowRate            *bool               `json:"show_rate"`
	SocialLinks         *models.SocialLinks `json:"social_links"`
	OnboardingCompleted *bool               `json:"onboarding_completed"`
	IsAcceptingClients  *bool               `json:"is_accepting_clients"`
}

type CreateInviteCodeInput struct {
	ExpiresInDays int `json:"expires_in_days"`
}

type InvitePreview struct {
	Code         string    `json:"code"`
	CoachID      uint      `json:"coach_id"`
	BusinessName *string   `json:"business_name"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type AcceptInviteInput struct {
	Code string `json:"code" binding:"required"`
}

type AcceptInviteResult struct {
	ClientProfile    *models.ClientProfile `json:"client_profile"`
	AlreadyConnected bool                  `json:"already_connected"`
}

type CoachService struct {
	repos           *repositories.RepositoriesCollection
	coachRepo       *repositories.CoachRepository
	clientRepo      *repositories.ClientRepository
	eventsPublisher *events.Publisher
}

func NewCoachService(
	repos *repositories.RepositoriesCollection,
	eventsPublisher *events.Publisher,
) *CoachService {
	return &CoachService{
		repos:           repos,
		coachRepo:       repos.Coach,
		clientRepo:      repos.Client,
		eventsPublisher: eventsPublisher,
	}
}

func (s *CoachService) GetMyProfile(ctx context.Context, userID uint) (*models.CoachProfile, error) {
	profile, err := s.coachRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCoachProfileNotFound
		}
		return nil, err
	}
	return profile, nil
}

func (s *CoachService) UpsertMyProfile(ctx context.Context, userID uint, input UpsertCoachProfileInput) (*models.CoachProfile, error) {
	profile, err := s.coachRepo.GetByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		profile = &models.CoachProfile{
			UserID:             userID,
			TrainingType:       "hybrid",
			HourlyRateCurrency: "USD",
			SubscriptionTier:   "free",
			IsAcceptingClients: true,
		}
		if input.Specialties != nil {
			profile.Specialties = *input.Specialties
		}
		if input.Languages != nil {
			profile.Languages = *input.Languages
		}
		if input.SocialLinks != nil {
			profile.SocialLinks = *input.SocialLinks
		}

		applyCoachProfileUpdates(profile, input)

		if err := s.coachRepo.Create(ctx, profile); err != nil {
			return nil, err
		}

		// Initialize coach stats row on profile creation.
		stats := &models.CoachStats{CoachID: profile.ID}
		if err := s.coachRepo.UpdateStats(ctx, stats); err != nil {
			return nil, err
		}

		return s.coachRepo.GetByID(ctx, profile.ID)
	}

	applyCoachProfileUpdates(profile, input)
	if err := s.coachRepo.Update(ctx, profile); err != nil {
		return nil, err
	}
	return s.coachRepo.GetByID(ctx, profile.ID)
}

func (s *CoachService) CreateInviteCode(ctx context.Context, userID uint, input CreateInviteCodeInput) (*models.InviteCode, error) {
	profile, err := s.coachRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCoachProfileNotFound
		}
		return nil, err
	}

	days := input.ExpiresInDays
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	var invite *models.InviteCode
	for i := 0; i < 5; i++ {
		code, codeErr := generateInviteCode(10)
		if codeErr != nil {
			return nil, codeErr
		}

		candidate := &models.InviteCode{
			CoachID:   profile.ID,
			Code:      code,
			ExpiresAt: time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour),
			IsActive:  true,
		}

		if err := s.clientRepo.CreateInviteCode(ctx, candidate); err != nil {
			// Retry on code collisions from unique constraint.
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			return nil, err
		}

		invite = candidate
		break
	}
	if invite == nil {
		return nil, fmt.Errorf("failed to generate unique invite code")
	}

	return invite, nil
}

func (s *CoachService) ListInviteCodes(ctx context.Context, userID uint) ([]models.InviteCode, error) {
	profile, err := s.coachRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCoachProfileNotFound
		}
		return nil, err
	}
	return s.clientRepo.ListInviteCodes(ctx, profile.ID)
}

func (s *CoachService) DeactivateInviteCode(ctx context.Context, userID, inviteID uint) error {
	profile, err := s.coachRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCoachProfileNotFound
		}
		return err
	}

	invite, err := s.clientRepo.GetInviteCodeByID(ctx, inviteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInviteCodeNotFound
		}
		return err
	}

	if invite.CoachID != profile.ID {
		return ErrInviteForbidden
	}

	return s.clientRepo.DeactivateInviteCode(ctx, inviteID)
}

func (s *CoachService) GetInvitePreview(ctx context.Context, code string) (*InvitePreview, error) {
	invite, err := s.clientRepo.GetInviteCode(ctx, strings.ToUpper(strings.TrimSpace(code)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInviteCodeNotFound
		}
		return nil, err
	}

	coach, err := s.coachRepo.GetByID(ctx, invite.CoachID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCoachProfileNotFound
		}
		return nil, err
	}

	return &InvitePreview{
		Code:         invite.Code,
		CoachID:      coach.ID,
		BusinessName: coach.BusinessName,
		ExpiresAt:    invite.ExpiresAt,
	}, nil
}

func (s *CoachService) AcceptInvite(ctx context.Context, userID uint, input AcceptInviteInput) (*AcceptInviteResult, error) {
	code := strings.ToUpper(strings.TrimSpace(input.Code))
	if code == "" {
		return nil, ErrInviteCodeNotFound
	}

	var result *AcceptInviteResult

	err := s.repos.WithTransaction(ctx, func(tx *gorm.DB, txRepos *repositories.RepositoriesCollection) error {
		invite, err := txRepos.Client.GetInviteCode(ctx, code)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInviteCodeNotFound
			}
			return err
		}

		clientProfile, alreadyConnected, err := txRepos.Client.AcceptInvite(ctx, invite, userID)
		if err != nil {
			return err
		}

		if !alreadyConnected {
			if err := txRepos.Coach.IncrementStat(ctx, invite.CoachID, "active_clients", 1); err != nil {
				return err
			}
			if err := txRepos.Coach.IncrementStat(ctx, invite.CoachID, "total_clients_all_time", 1); err != nil {
				return err
			}
		}

		if s.eventsPublisher != nil {
			payload := events.InviteAcceptedPayload{
				InviteCodeID:    invite.ID,
				CoachID:         invite.CoachID,
				ClientUserID:    userID,
				ClientProfileID: clientProfile.ID,
				Code:            invite.Code,
			}
			idempotencyKey := events.BuildIdempotencyKey(
				events.EventTypeInviteAccepted,
				strconv.FormatUint(uint64(invite.ID), 10),
				strconv.FormatUint(uint64(userID), 10),
			)
			if err := s.eventsPublisher.PublishInTx(
				ctx,
				tx,
				events.EventTypeInviteAccepted,
				"client_profile",
				strconv.FormatUint(uint64(clientProfile.ID), 10),
				idempotencyKey,
				payload,
			); err != nil {
				return err
			}
		}

		result = &AcceptInviteResult{
			ClientProfile:    clientProfile,
			AlreadyConnected: alreadyConnected,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func applyCoachProfileUpdates(profile *models.CoachProfile, input UpsertCoachProfileInput) {
	if input.BusinessName != nil {
		profile.BusinessName = input.BusinessName
	}
	if input.Bio != nil {
		profile.Bio = input.Bio
	}
	if input.CoverPhotoURL != nil {
		profile.CoverPhotoURL = input.CoverPhotoURL
	}
	if input.Specialties != nil {
		profile.Specialties = *input.Specialties
	}
	if input.YearsExperience != nil {
		profile.YearsExperience = input.YearsExperience
	}
	if input.Languages != nil {
		profile.Languages = *input.Languages
	}
	if input.TrainingType != nil && strings.TrimSpace(*input.TrainingType) != "" {
		profile.TrainingType = strings.TrimSpace(*input.TrainingType)
	}
	if input.HourlyRate != nil {
		profile.HourlyRate = input.HourlyRate
	}
	if input.HourlyRateCurrency != nil && strings.TrimSpace(*input.HourlyRateCurrency) != "" {
		profile.HourlyRateCurrency = strings.ToUpper(strings.TrimSpace(*input.HourlyRateCurrency))
	}
	if input.ShowRate != nil {
		profile.ShowRate = *input.ShowRate
	}
	if input.SocialLinks != nil {
		profile.SocialLinks = *input.SocialLinks
	}
	if input.OnboardingCompleted != nil {
		profile.OnboardingCompleted = *input.OnboardingCompleted
	}
	if input.IsAcceptingClients != nil {
		profile.IsAcceptingClients = *input.IsAcceptingClients
	}
}

func generateInviteCode(length int) (string, error) {
	if length <= 0 {
		length = 10
	}
	raw, err := utils.GenerateRandomString(length)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(raw), nil
}
