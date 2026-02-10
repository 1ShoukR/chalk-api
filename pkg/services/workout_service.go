package services

import (
	"chalk-api/pkg/events"
	"chalk-api/pkg/models"
	"chalk-api/pkg/repositories"
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrTemplateNotFound        = errors.New("template not found")
	ErrTemplateForbidden       = errors.New("template does not belong to this coach")
	ErrWorkoutNotFound         = errors.New("workout not found")
	ErrWorkoutForbidden        = errors.New("workout does not belong to this user")
	ErrWorkoutExerciseNotFound = errors.New("workout exercise not found")
	ErrWorkoutLogNotFound      = errors.New("workout log not found")
	ErrClientProfileNotFound   = errors.New("client profile not found")
	ErrClientProfileForbidden  = errors.New("client profile does not belong to this coach")
	ErrInvalidWorkoutState     = errors.New("invalid workout state transition")
	ErrInvalidScheduledDate    = errors.New("scheduled date must be YYYY-MM-DD")
)

type TemplateExerciseInput struct {
	ExerciseID       uint     `json:"exercise_id" binding:"required"`
	OrderIndex       int      `json:"order_index"`
	SectionLabel     *string  `json:"section_label"`
	SupersetGroup    *int     `json:"superset_group"`
	GroupType        *string  `json:"group_type"`
	Sets             *int     `json:"sets"`
	RepsMin          *int     `json:"reps_min"`
	RepsMax          *int     `json:"reps_max"`
	WeightValue      *float64 `json:"weight_value"`
	WeightUnit       *string  `json:"weight_unit"`
	PrescriptionNote *string  `json:"prescription_note"`
	RestSeconds      *int     `json:"rest_seconds"`
	Tempo            *string  `json:"tempo"`
	Notes            *string  `json:"notes"`
}

type CreateWorkoutTemplateInput struct {
	Name             string                  `json:"name" binding:"required"`
	Description      *string                 `json:"description"`
	Category         *string                 `json:"category"`
	Tags             []string                `json:"tags"`
	EstimatedMinutes *int                    `json:"estimated_minutes"`
	Exercises        []TemplateExerciseInput `json:"exercises"`
}

type UpdateWorkoutTemplateInput struct {
	Name             *string                  `json:"name"`
	Description      *string                  `json:"description"`
	Category         *string                  `json:"category"`
	Tags             *[]string                `json:"tags"`
	EstimatedMinutes *int                     `json:"estimated_minutes"`
	IsActive         *bool                    `json:"is_active"`
	Exercises        *[]TemplateExerciseInput `json:"exercises"`
}

type AssignWorkoutInput struct {
	TemplateID      uint    `json:"template_id" binding:"required"`
	ClientProfileID uint    `json:"client_profile_id" binding:"required"`
	ScheduledDate   *string `json:"scheduled_date"` // YYYY-MM-DD
}

type SkipWorkoutExerciseInput struct {
	Reason string `json:"reason" binding:"required"`
}

type CreateWorkoutLogInput struct {
	SetNumber       int      `json:"set_number" binding:"required"`
	RepsCompleted   *int     `json:"reps_completed"`
	WeightUsed      *float64 `json:"weight_used"`
	WeightUnit      *string  `json:"weight_unit"`
	RPE             *int     `json:"rpe"`
	Notes           *string  `json:"notes"`
	DurationSeconds *int     `json:"duration_seconds"`
	Distance        *float64 `json:"distance"`
	DistanceUnit    *string  `json:"distance_unit"`
}

type UpdateWorkoutLogInput struct {
	SetNumber       *int     `json:"set_number"`
	RepsCompleted   *int     `json:"reps_completed"`
	WeightUsed      *float64 `json:"weight_used"`
	WeightUnit      *string  `json:"weight_unit"`
	RPE             *int     `json:"rpe"`
	Notes           *string  `json:"notes"`
	DurationSeconds *int     `json:"duration_seconds"`
	Distance        *float64 `json:"distance"`
	DistanceUnit    *string  `json:"distance_unit"`
}

type WorkoutService struct {
	repos        *repositories.RepositoriesCollection
	templateRepo *repositories.TemplateRepository
	workoutRepo  *repositories.WorkoutRepository
	coachRepo    *repositories.CoachRepository
	clientRepo   *repositories.ClientRepository
	events       *events.Publisher
}

func NewWorkoutService(
	repos *repositories.RepositoriesCollection,
	eventsPublisher *events.Publisher,
) *WorkoutService {
	return &WorkoutService{
		repos:        repos,
		templateRepo: repos.Template,
		workoutRepo:  repos.Workout,
		coachRepo:    repos.Coach,
		clientRepo:   repos.Client,
		events:       eventsPublisher,
	}
}

func (s *WorkoutService) CreateTemplate(ctx context.Context, userID uint, input CreateWorkoutTemplateInput) (*models.WorkoutTemplate, error) {
	coachProfile, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrTemplateNotFound
	}

	template := &models.WorkoutTemplate{
		CoachID:          coachProfile.ID,
		Name:             name,
		Description:      input.Description,
		Category:         input.Category,
		Tags:             input.Tags,
		EstimatedMinutes: input.EstimatedMinutes,
		IsActive:         true,
	}

	template.Exercises = buildTemplateExercises(input.Exercises)

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	return s.templateRepo.GetByID(ctx, template.ID)
}

func (s *WorkoutService) ListMyTemplates(ctx context.Context, userID uint, limit, offset int) ([]models.WorkoutTemplate, int64, error) {
	coachProfile, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.templateRepo.ListByCoach(ctx, coachProfile.ID, limit, offset)
}

func (s *WorkoutService) GetMyTemplate(ctx context.Context, userID, templateID uint) (*models.WorkoutTemplate, error) {
	coachProfile, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	if template.CoachID != coachProfile.ID {
		return nil, ErrTemplateForbidden
	}

	return template, nil
}

func (s *WorkoutService) UpdateMyTemplate(ctx context.Context, userID, templateID uint, input UpdateWorkoutTemplateInput) (*models.WorkoutTemplate, error) {
	template, err := s.GetMyTemplate(ctx, userID, templateID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed != "" {
			template.Name = trimmed
		}
	}
	if input.Description != nil {
		template.Description = input.Description
	}
	if input.Category != nil {
		template.Category = input.Category
	}
	if input.Tags != nil {
		template.Tags = *input.Tags
	}
	if input.EstimatedMinutes != nil {
		template.EstimatedMinutes = input.EstimatedMinutes
	}
	if input.IsActive != nil {
		template.IsActive = *input.IsActive
	}

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, err
	}

	if input.Exercises != nil {
		exercises := buildTemplateExercises(*input.Exercises)
		if err := s.templateRepo.ReplaceExercises(ctx, template.ID, exercises); err != nil {
			return nil, err
		}
	}

	return s.templateRepo.GetByID(ctx, template.ID)
}

func (s *WorkoutService) AssignTemplateToClient(ctx context.Context, userID uint, input AssignWorkoutInput) (*models.Workout, error) {
	coachProfile, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	template, err := s.templateRepo.GetByID(ctx, input.TemplateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	if template.CoachID != coachProfile.ID {
		return nil, ErrTemplateForbidden
	}
	if !template.IsActive {
		return nil, ErrTemplateNotFound
	}

	clientProfile, err := s.clientRepo.GetByID(ctx, input.ClientProfileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClientProfileNotFound
		}
		return nil, err
	}
	if clientProfile.CoachID != coachProfile.ID {
		return nil, ErrClientProfileForbidden
	}

	scheduledDate, err := normalizeScheduledDate(input.ScheduledDate)
	if err != nil {
		return nil, err
	}

	workout := &models.Workout{
		ClientID:      clientProfile.ID,
		CoachID:       coachProfile.ID,
		TemplateID:    &template.ID,
		Name:          template.Name,
		Description:   template.Description,
		ScheduledDate: scheduledDate,
		Status:        "scheduled",
	}
	workout.Exercises = buildWorkoutExercisesFromTemplate(template.Exercises)

	if err := s.repos.WithTransaction(ctx, func(tx *gorm.DB, txRepos *repositories.RepositoriesCollection) error {
		if err := txRepos.Workout.Create(ctx, workout); err != nil {
			return err
		}

		if s.events != nil {
			payload := events.WorkoutAssignedPayload{
				WorkoutID:      workout.ID,
				CoachID:        workout.CoachID,
				ClientID:       workout.ClientID,
				ScheduledDate:  safeString(workout.ScheduledDate),
				WorkoutName:    workout.Name,
				AssignedByUser: userID,
			}
			idempotencyKey := events.BuildIdempotencyKey(
				events.EventTypeWorkoutAssigned,
				strconv.FormatUint(uint64(workout.ID), 10),
			)
			if err := s.events.PublishInTx(
				ctx,
				tx,
				events.EventTypeWorkoutAssigned,
				"workout",
				strconv.FormatUint(uint64(workout.ID), 10),
				idempotencyKey,
				payload,
			); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return s.workoutRepo.GetByID(ctx, workout.ID)
}

func (s *WorkoutService) ListMyWorkouts(ctx context.Context, userID uint, limit, offset int) ([]models.Workout, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	clientProfiles, err := s.clientRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if len(clientProfiles) == 0 {
		return []models.Workout{}, 0, nil
	}

	clientIDs := make([]uint, 0, len(clientProfiles))
	for i := range clientProfiles {
		clientIDs = append(clientIDs, clientProfiles[i].ID)
	}

	return s.workoutRepo.ListByClients(ctx, clientIDs, limit, offset)
}

func (s *WorkoutService) GetMyWorkout(ctx context.Context, userID, workoutID uint) (*models.Workout, error) {
	workout, err := s.workoutRepo.GetByID(ctx, workoutID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	if err := s.ensureWorkoutOwnedByUser(ctx, userID, workout); err != nil {
		return nil, err
	}
	return workout, nil
}

func (s *WorkoutService) StartMyWorkout(ctx context.Context, userID, workoutID uint) (*models.Workout, error) {
	workout, err := s.GetMyWorkout(ctx, userID, workoutID)
	if err != nil {
		return nil, err
	}

	if workout.Status == "completed" || workout.Status == "skipped" {
		return nil, ErrInvalidWorkoutState
	}

	if err := s.workoutRepo.StartWorkout(ctx, workoutID); err != nil {
		return nil, err
	}

	return s.workoutRepo.GetByID(ctx, workoutID)
}

func (s *WorkoutService) CompleteMyWorkout(ctx context.Context, userID, workoutID uint) (*models.Workout, error) {
	workout, err := s.GetMyWorkout(ctx, userID, workoutID)
	if err != nil {
		return nil, err
	}

	if workout.Status == "completed" || workout.Status == "skipped" {
		return nil, ErrInvalidWorkoutState
	}

	completedAt := time.Now().UTC()
	if err := s.repos.WithTransaction(ctx, func(tx *gorm.DB, txRepos *repositories.RepositoriesCollection) error {
		if err := txRepos.Workout.CompleteWorkout(ctx, workoutID); err != nil {
			return err
		}

		if s.events != nil {
			payload := events.WorkoutCompletedPayload{
				WorkoutID:   workout.ID,
				CoachID:     workout.CoachID,
				ClientID:    workout.ClientID,
				CompletedAt: completedAt,
			}
			idempotencyKey := events.BuildIdempotencyKey(
				events.EventTypeWorkoutCompleted,
				strconv.FormatUint(uint64(workout.ID), 10),
			)
			if err := s.events.PublishInTx(
				ctx,
				tx,
				events.EventTypeWorkoutCompleted,
				"workout",
				strconv.FormatUint(uint64(workout.ID), 10),
				idempotencyKey,
				payload,
			); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return s.workoutRepo.GetByID(ctx, workoutID)
}

func (s *WorkoutService) MarkMyExerciseCompleted(ctx context.Context, userID, workoutExerciseID uint) (*models.WorkoutExercise, error) {
	exercise, err := s.workoutRepo.GetExerciseByID(ctx, workoutExerciseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutExerciseNotFound
		}
		return nil, err
	}

	if err := s.ensureWorkoutOwnershipByID(ctx, userID, exercise.WorkoutID); err != nil {
		return nil, err
	}

	if err := s.workoutRepo.MarkExerciseCompleted(ctx, workoutExerciseID); err != nil {
		return nil, err
	}
	return s.workoutRepo.GetExerciseByID(ctx, workoutExerciseID)
}

func (s *WorkoutService) SkipMyExercise(ctx context.Context, userID, workoutExerciseID uint, input SkipWorkoutExerciseInput) (*models.WorkoutExercise, error) {
	exercise, err := s.workoutRepo.GetExerciseByID(ctx, workoutExerciseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutExerciseNotFound
		}
		return nil, err
	}
	if err := s.ensureWorkoutOwnershipByID(ctx, userID, exercise.WorkoutID); err != nil {
		return nil, err
	}

	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = "skipped"
	}
	if err := s.workoutRepo.SkipExercise(ctx, workoutExerciseID, reason); err != nil {
		return nil, err
	}
	return s.workoutRepo.GetExerciseByID(ctx, workoutExerciseID)
}

func (s *WorkoutService) CreateMyExerciseLog(ctx context.Context, userID, workoutExerciseID uint, input CreateWorkoutLogInput) (*models.WorkoutLog, error) {
	exercise, err := s.workoutRepo.GetExerciseByID(ctx, workoutExerciseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutExerciseNotFound
		}
		return nil, err
	}
	if err := s.ensureWorkoutOwnershipByID(ctx, userID, exercise.WorkoutID); err != nil {
		return nil, err
	}

	log := &models.WorkoutLog{
		WorkoutExerciseID: workoutExerciseID,
		SetNumber:         input.SetNumber,
		RepsCompleted:     input.RepsCompleted,
		WeightUsed:        input.WeightUsed,
		WeightUnit:        input.WeightUnit,
		RPE:               input.RPE,
		Notes:             input.Notes,
		DurationSeconds:   input.DurationSeconds,
		Distance:          input.Distance,
		DistanceUnit:      input.DistanceUnit,
	}
	if err := s.workoutRepo.CreateLog(ctx, log); err != nil {
		return nil, err
	}

	return s.workoutRepo.GetLogByID(ctx, log.ID)
}

func (s *WorkoutService) UpdateMyWorkoutLog(ctx context.Context, userID, workoutLogID uint, input UpdateWorkoutLogInput) (*models.WorkoutLog, error) {
	logEntry, err := s.workoutRepo.GetLogByID(ctx, workoutLogID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutLogNotFound
		}
		return nil, err
	}

	if err := s.ensureWorkoutOwnershipByExerciseID(ctx, userID, logEntry.WorkoutExerciseID); err != nil {
		return nil, err
	}

	if input.SetNumber != nil {
		logEntry.SetNumber = *input.SetNumber
	}
	if input.RepsCompleted != nil {
		logEntry.RepsCompleted = input.RepsCompleted
	}
	if input.WeightUsed != nil {
		logEntry.WeightUsed = input.WeightUsed
	}
	if input.WeightUnit != nil {
		logEntry.WeightUnit = input.WeightUnit
	}
	if input.RPE != nil {
		logEntry.RPE = input.RPE
	}
	if input.Notes != nil {
		logEntry.Notes = input.Notes
	}
	if input.DurationSeconds != nil {
		logEntry.DurationSeconds = input.DurationSeconds
	}
	if input.Distance != nil {
		logEntry.Distance = input.Distance
	}
	if input.DistanceUnit != nil {
		logEntry.DistanceUnit = input.DistanceUnit
	}

	if err := s.workoutRepo.UpdateLog(ctx, logEntry); err != nil {
		return nil, err
	}

	return s.workoutRepo.GetLogByID(ctx, logEntry.ID)
}

func (s *WorkoutService) getCoachProfile(ctx context.Context, userID uint) (*models.CoachProfile, error) {
	profile, err := s.coachRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCoachProfileNotFound
		}
		return nil, err
	}
	return profile, nil
}

func (s *WorkoutService) ensureWorkoutOwnershipByID(ctx context.Context, userID, workoutID uint) error {
	workout, err := s.workoutRepo.GetByID(ctx, workoutID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkoutNotFound
		}
		return err
	}
	return s.ensureWorkoutOwnedByUser(ctx, userID, workout)
}

func (s *WorkoutService) ensureWorkoutOwnershipByExerciseID(ctx context.Context, userID, workoutExerciseID uint) error {
	exercise, err := s.workoutRepo.GetExerciseByID(ctx, workoutExerciseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkoutExerciseNotFound
		}
		return err
	}
	return s.ensureWorkoutOwnershipByID(ctx, userID, exercise.WorkoutID)
}

func (s *WorkoutService) ensureWorkoutOwnedByUser(ctx context.Context, userID uint, workout *models.Workout) error {
	clientProfile, err := s.clientRepo.GetByID(ctx, workout.ClientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkoutForbidden
		}
		return err
	}

	if clientProfile.UserID != userID {
		return ErrWorkoutForbidden
	}

	return nil
}

func buildTemplateExercises(inputs []TemplateExerciseInput) []models.WorkoutTemplateExercise {
	exercises := make([]models.WorkoutTemplateExercise, 0, len(inputs))
	for i := range inputs {
		order := inputs[i].OrderIndex
		if order <= 0 {
			order = i + 1
		}

		exercises = append(exercises, models.WorkoutTemplateExercise{
			ExerciseID:       inputs[i].ExerciseID,
			OrderIndex:       order,
			SectionLabel:     inputs[i].SectionLabel,
			SupersetGroup:    inputs[i].SupersetGroup,
			GroupType:        inputs[i].GroupType,
			Sets:             inputs[i].Sets,
			RepsMin:          inputs[i].RepsMin,
			RepsMax:          inputs[i].RepsMax,
			WeightValue:      inputs[i].WeightValue,
			WeightUnit:       inputs[i].WeightUnit,
			PrescriptionNote: inputs[i].PrescriptionNote,
			RestSeconds:      inputs[i].RestSeconds,
			Tempo:            inputs[i].Tempo,
			Notes:            inputs[i].Notes,
		})
	}
	return exercises
}

func buildWorkoutExercisesFromTemplate(templateExercises []models.WorkoutTemplateExercise) []models.WorkoutExercise {
	result := make([]models.WorkoutExercise, 0, len(templateExercises))
	for i := range templateExercises {
		templateExercise := templateExercises[i]
		result = append(result, models.WorkoutExercise{
			ExerciseID:       templateExercise.ExerciseID,
			OrderIndex:       templateExercise.OrderIndex,
			SectionLabel:     templateExercise.SectionLabel,
			SupersetGroup:    templateExercise.SupersetGroup,
			GroupType:        templateExercise.GroupType,
			Sets:             templateExercise.Sets,
			RepsMin:          templateExercise.RepsMin,
			RepsMax:          templateExercise.RepsMax,
			WeightValue:      templateExercise.WeightValue,
			WeightUnit:       templateExercise.WeightUnit,
			PrescriptionNote: templateExercise.PrescriptionNote,
			RestSeconds:      templateExercise.RestSeconds,
			Tempo:            templateExercise.Tempo,
			Notes:            templateExercise.Notes,
		})
	}
	return result
}

func normalizeScheduledDate(scheduledDate *string) (*string, error) {
	if scheduledDate == nil {
		return nil, nil
	}

	value := strings.TrimSpace(*scheduledDate)
	if value == "" {
		return nil, nil
	}

	if _, err := time.Parse("2006-01-02", value); err != nil {
		return nil, ErrInvalidScheduledDate
	}

	return &value, nil
}

func safeString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
