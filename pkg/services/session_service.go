package services

import (
	"chalk-api/pkg/events"
	"chalk-api/pkg/models"
	"chalk-api/pkg/repositories"
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrSessionTypeInvalid      = errors.New("invalid session type payload")
	ErrSessionTypeNotFound     = errors.New("session type not found")
	ErrSessionTypeForbidden    = errors.New("session type does not belong to this coach")
	ErrSessionTypeInactive     = errors.New("session type is inactive")
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionForbidden        = errors.New("session does not belong to this user")
	ErrSessionActionForbidden  = errors.New("session action is not allowed for this user")
	ErrSessionStateInvalid     = errors.New("invalid session state transition")
	ErrSessionConflict         = errors.New("requested time conflicts with an existing session")
	ErrOutsideAvailability     = errors.New("requested time is outside coach availability")
	ErrAvailabilitySlotInvalid = errors.New("invalid availability slot")
	ErrOverrideNotFound        = errors.New("availability override not found")
	ErrOverrideForbidden       = errors.New("availability override does not belong to this coach")
	ErrInvalidDateRange        = errors.New("invalid date range")
	ErrInvalidDateFormat       = errors.New("invalid date format, expected YYYY-MM-DD")
	ErrInvalidScheduledAt      = errors.New("invalid scheduled_at, expected RFC3339 datetime")
	ErrInvalidSessionDuration  = errors.New("invalid session duration")
)

const (
	defaultBookableRangeDays = 14
	defaultListRangeDays     = 30
	maxRangeDays             = 90
	slotStepMinutes          = 15
)

type AvailabilitySlotInput struct {
	DayOfWeek int    `json:"day_of_week" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	IsActive  *bool  `json:"is_active"`
}

type SetAvailabilityInput struct {
	Slots []AvailabilitySlotInput `json:"slots"`
}

type CreateAvailabilityOverrideInput struct {
	Date        string  `json:"date" binding:"required"`
	IsAvailable bool    `json:"is_available"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Reason      *string `json:"reason"`
}

type CreateSessionTypeInput struct {
	Name            string  `json:"name" binding:"required"`
	DurationMinutes int     `json:"duration_minutes" binding:"required"`
	Description     *string `json:"description"`
	Color           *string `json:"color"`
}

type UpdateSessionTypeInput struct {
	Name            *string `json:"name"`
	DurationMinutes *int    `json:"duration_minutes"`
	Description     *string `json:"description"`
	Color           *string `json:"color"`
	IsActive        *bool   `json:"is_active"`
}

type BookSessionInput struct {
	ClientProfileID uint    `json:"client_profile_id" binding:"required"`
	SessionTypeID   uint    `json:"session_type_id" binding:"required"`
	ScheduledAt     string  `json:"scheduled_at" binding:"required"` // RFC3339, converted to UTC
	Location        *string `json:"location"`
	Notes           *string `json:"notes"`
}

type CancelSessionInput struct {
	Reason *string `json:"reason"`
}

type BookableSlot struct {
	StartAt         time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	DurationMinutes int       `json:"duration_minutes"`
	CoachID         uint      `json:"coach_id"`
	SessionTypeID   *uint     `json:"session_type_id,omitempty"`
}

type SessionService struct {
	repos       *repositories.RepositoriesCollection
	coachRepo   *repositories.CoachRepository
	clientRepo  *repositories.ClientRepository
	sessionRepo *repositories.SessionRepository
	events      *events.Publisher
}

func NewSessionService(
	repos *repositories.RepositoriesCollection,
	eventsPublisher *events.Publisher,
) *SessionService {
	return &SessionService{
		repos:       repos,
		coachRepo:   repos.Coach,
		clientRepo:  repos.Client,
		sessionRepo: repos.Session,
		events:      eventsPublisher,
	}
}

func (s *SessionService) GetMyAvailability(ctx context.Context, userID uint) ([]models.CoachAvailability, error) {
	coach, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.sessionRepo.GetAvailability(ctx, coach.ID)
}

func (s *SessionService) SetMyAvailability(ctx context.Context, userID uint, input SetAvailabilityInput) ([]models.CoachAvailability, error) {
	coach, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	slots, err := buildValidatedAvailabilitySlots(coach.ID, input.Slots)
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.SetAvailability(ctx, coach.ID, slots); err != nil {
		return nil, err
	}

	return s.sessionRepo.GetAvailability(ctx, coach.ID)
}

func (s *SessionService) CreateAvailabilityOverride(ctx context.Context, userID uint, input CreateAvailabilityOverrideInput) (*models.CoachAvailabilityOverride, error) {
	coach, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	date, err := parseDateOnly(input.Date)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	override := &models.CoachAvailabilityOverride{
		CoachID:     coach.ID,
		Date:        date.Format("2006-01-02"),
		IsAvailable: input.IsAvailable,
		StartTime:   nil,
		EndTime:     nil,
		Reason:      trimSessionPtr(input.Reason),
	}

	if input.IsAvailable {
		start, end, err := parseOptionalTimeRange(input.StartTime, input.EndTime)
		if err != nil {
			return nil, err
		}
		override.StartTime = &start
		override.EndTime = &end
	}

	if err := s.sessionRepo.CreateOverride(ctx, override); err != nil {
		return nil, err
	}

	return override, nil
}

func (s *SessionService) ListMyAvailabilityOverrides(ctx context.Context, userID uint, startDateRaw, endDateRaw string) ([]models.CoachAvailabilityOverride, error) {
	coach, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	startDate, endDate, err := parseDateRange(startDateRaw, endDateRaw, defaultBookableRangeDays)
	if err != nil {
		return nil, err
	}

	return s.sessionRepo.ListOverrides(
		ctx,
		coach.ID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)
}

func (s *SessionService) DeleteMyAvailabilityOverride(ctx context.Context, userID, overrideID uint) error {
	coach, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return err
	}

	override, err := s.sessionRepo.GetOverrideByID(ctx, overrideID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOverrideNotFound
		}
		return err
	}
	if override.CoachID != coach.ID {
		return ErrOverrideForbidden
	}

	return s.sessionRepo.DeleteOverride(ctx, overrideID)
}

func (s *SessionService) CreateMySessionType(ctx context.Context, userID uint, input CreateSessionTypeInput) (*models.SessionType, error) {
	coach, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrSessionTypeInvalid
	}
	if !isValidSessionDuration(input.DurationMinutes) {
		return nil, ErrInvalidSessionDuration
	}

	sessionType := &models.SessionType{
		CoachID:         coach.ID,
		Name:            name,
		DurationMinutes: input.DurationMinutes,
		Description:     trimSessionPtr(input.Description),
		Color:           trimSessionPtr(input.Color),
		IsActive:        true,
	}

	if err := s.sessionRepo.CreateSessionType(ctx, sessionType); err != nil {
		return nil, err
	}
	return sessionType, nil
}

func (s *SessionService) ListMySessionTypes(ctx context.Context, userID uint) ([]models.SessionType, error) {
	coach, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.sessionRepo.ListSessionTypes(ctx, coach.ID)
}

func (s *SessionService) UpdateMySessionType(ctx context.Context, userID, sessionTypeID uint, input UpdateSessionTypeInput) (*models.SessionType, error) {
	coach, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	sessionType, err := s.sessionRepo.GetSessionTypeByID(ctx, sessionTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionTypeNotFound
		}
		return nil, err
	}
	if sessionType.CoachID != coach.ID {
		return nil, ErrSessionTypeForbidden
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name != "" {
			sessionType.Name = name
		}
	}
	if input.DurationMinutes != nil {
		if !isValidSessionDuration(*input.DurationMinutes) {
			return nil, ErrInvalidSessionDuration
		}
		sessionType.DurationMinutes = *input.DurationMinutes
	}
	if input.Description != nil {
		sessionType.Description = trimSessionPtr(input.Description)
	}
	if input.Color != nil {
		sessionType.Color = trimSessionPtr(input.Color)
	}
	if input.IsActive != nil {
		sessionType.IsActive = *input.IsActive
	}

	if err := s.sessionRepo.UpdateSessionType(ctx, sessionType); err != nil {
		return nil, err
	}
	return sessionType, nil
}

func (s *SessionService) GetBookableSlots(
	ctx context.Context,
	coachID uint,
	startDateRaw string,
	endDateRaw string,
	sessionTypeID *uint,
	durationMinutes *int,
) ([]BookableSlot, error) {
	if _, err := s.coachRepo.GetByID(ctx, coachID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCoachProfileNotFound
		}
		return nil, err
	}

	resolvedDuration, err := s.resolveBookableDuration(ctx, coachID, sessionTypeID, durationMinutes)
	if err != nil {
		return nil, err
	}

	startDate, endDate, err := parseDateRange(startDateRaw, endDateRaw, defaultBookableRangeDays)
	if err != nil {
		return nil, err
	}

	availability, err := s.sessionRepo.GetAvailability(ctx, coachID)
	if err != nil {
		return nil, err
	}
	overrides, err := s.sessionRepo.ListOverrides(ctx, coachID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	sessions, err := s.sessionRepo.ListSessions(ctx, coachID, 0, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return buildBookableSlots(startDate, endDate, coachID, sessionTypeID, resolvedDuration, availability, overrides, sessions), nil
}

func (s *SessionService) BookSession(ctx context.Context, userID uint, input BookSessionInput) (*models.Session, error) {
	if input.ClientProfileID == 0 {
		return nil, ErrClientProfileNotFound
	}
	if input.SessionTypeID == 0 {
		return nil, ErrSessionTypeNotFound
	}

	scheduledAt, err := time.Parse(time.RFC3339, strings.TrimSpace(input.ScheduledAt))
	if err != nil {
		return nil, ErrInvalidScheduledAt
	}
	scheduledAt = scheduledAt.UTC()
	if scheduledAt.Before(time.Now().UTC().Add(-1 * time.Minute)) {
		return nil, ErrInvalidScheduledAt
	}

	clientProfile, err := s.clientRepo.GetByID(ctx, input.ClientProfileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClientProfileNotFound
		}
		return nil, err
	}

	sessionType, err := s.sessionRepo.GetSessionTypeByID(ctx, input.SessionTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionTypeNotFound
		}
		return nil, err
	}
	if sessionType.CoachID != clientProfile.CoachID {
		return nil, ErrSessionTypeForbidden
	}
	if !sessionType.IsActive {
		return nil, ErrSessionTypeInactive
	}

	bookedBy, err := s.resolveBookedBy(ctx, userID, clientProfile.CoachID, clientProfile.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.assertSlotBookable(ctx, clientProfile.CoachID, scheduledAt, sessionType.DurationMinutes); err != nil {
		return nil, err
	}

	session := &models.Session{
		CoachID:         clientProfile.CoachID,
		ClientID:        clientProfile.ID,
		SessionTypeID:   sessionType.ID,
		ScheduledAt:     scheduledAt,
		DurationMinutes: sessionType.DurationMinutes,
		Status:          "scheduled",
		Location:        trimSessionPtr(input.Location),
		Notes:           trimSessionPtr(input.Notes),
	}

	if err := s.repos.WithTransaction(ctx, func(tx *gorm.DB, txRepos *repositories.RepositoriesCollection) error {
		if conflict, err := txRepos.Session.HasCoachConflict(
			ctx,
			session.CoachID,
			session.ScheduledAt,
			session.ScheduledAt.Add(time.Duration(session.DurationMinutes)*time.Minute),
			nil,
		); err != nil {
			return err
		} else if conflict {
			return ErrSessionConflict
		}

		if err := txRepos.Session.CreateSession(ctx, session); err != nil {
			return err
		}

		if s.events != nil {
			payload := events.SessionBookedPayload{
				SessionID:   session.ID,
				CoachID:     session.CoachID,
				ClientID:    session.ClientID,
				ScheduledAt: session.ScheduledAt,
				BookedBy:    bookedBy,
			}
			idempotencyKey := events.BuildIdempotencyKey(events.EventTypeSessionBooked, strconv.FormatUint(uint64(session.ID), 10))
			if err := s.events.PublishInTx(
				ctx,
				tx,
				events.EventTypeSessionBooked,
				"session",
				strconv.FormatUint(uint64(session.ID), 10),
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

	return s.sessionRepo.GetSession(ctx, session.ID)
}

func (s *SessionService) ListMySessions(ctx context.Context, userID uint, startDateRaw, endDateRaw string) ([]models.Session, error) {
	startDate, endDate, err := parseDateRange(startDateRaw, endDateRaw, defaultListRangeDays)
	if err != nil {
		return nil, err
	}

	clientProfiles, err := s.clientRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(clientProfiles) == 0 {
		return []models.Session{}, nil
	}

	clientIDs := make([]uint, 0, len(clientProfiles))
	for i := range clientProfiles {
		clientIDs = append(clientIDs, clientProfiles[i].ID)
	}

	return s.sessionRepo.ListSessionsByClients(ctx, clientIDs, startDate, endDate)
}

func (s *SessionService) ListCoachSessions(ctx context.Context, userID uint, startDateRaw, endDateRaw string) ([]models.Session, error) {
	coach, err := s.getCoachProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	startDate, endDate, err := parseDateRange(startDateRaw, endDateRaw, defaultListRangeDays)
	if err != nil {
		return nil, err
	}

	return s.sessionRepo.ListSessions(ctx, coach.ID, 0, startDate, endDate)
}

func (s *SessionService) CancelSession(ctx context.Context, userID, sessionID uint, input CancelSessionInput) (*models.Session, error) {
	session, err := s.getSessionForUser(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	actor := resolveSessionActor(session, userID)
	if actor == "" {
		return nil, ErrSessionForbidden
	}
	if session.Status != "scheduled" {
		return nil, ErrSessionStateInvalid
	}

	reason := "cancelled"
	if input.Reason != nil && strings.TrimSpace(*input.Reason) != "" {
		reason = strings.TrimSpace(*input.Reason)
	}

	if err := s.sessionRepo.CancelSession(ctx, session.ID, actor, reason); err != nil {
		return nil, err
	}

	return s.sessionRepo.GetSession(ctx, session.ID)
}

func (s *SessionService) CompleteSession(ctx context.Context, userID, sessionID uint) (*models.Session, error) {
	session, err := s.getSessionForUser(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if resolveSessionActor(session, userID) != "coach" {
		return nil, ErrSessionActionForbidden
	}
	if session.Status != "scheduled" {
		return nil, ErrSessionStateInvalid
	}

	if err := s.sessionRepo.CompleteSession(ctx, session.ID); err != nil {
		return nil, err
	}
	return s.sessionRepo.GetSession(ctx, session.ID)
}

func (s *SessionService) MarkNoShow(ctx context.Context, userID, sessionID uint) (*models.Session, error) {
	session, err := s.getSessionForUser(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if resolveSessionActor(session, userID) != "coach" {
		return nil, ErrSessionActionForbidden
	}
	if session.Status != "scheduled" {
		return nil, ErrSessionStateInvalid
	}

	if err := s.sessionRepo.MarkNoShow(ctx, session.ID); err != nil {
		return nil, err
	}
	return s.sessionRepo.GetSession(ctx, session.ID)
}

func (s *SessionService) resolveBookableDuration(ctx context.Context, coachID uint, sessionTypeID *uint, durationMinutes *int) (int, error) {
	if sessionTypeID != nil && *sessionTypeID > 0 {
		sessionType, err := s.sessionRepo.GetSessionTypeByID(ctx, *sessionTypeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, ErrSessionTypeNotFound
			}
			return 0, err
		}
		if sessionType.CoachID != coachID {
			return 0, ErrSessionTypeForbidden
		}
		if !sessionType.IsActive {
			return 0, ErrSessionTypeInactive
		}
		return sessionType.DurationMinutes, nil
	}

	if durationMinutes != nil {
		if !isValidSessionDuration(*durationMinutes) {
			return 0, ErrInvalidSessionDuration
		}
		return *durationMinutes, nil
	}

	return 60, nil
}

func (s *SessionService) assertSlotBookable(ctx context.Context, coachID uint, scheduledAt time.Time, durationMinutes int) error {
	if !isValidSessionDuration(durationMinutes) {
		return ErrInvalidSessionDuration
	}

	dateStart := time.Date(scheduledAt.Year(), scheduledAt.Month(), scheduledAt.Day(), 0, 0, 0, 0, time.UTC)
	dateEnd := dateStart.Add(24*time.Hour - time.Nanosecond)

	availability, err := s.sessionRepo.GetAvailability(ctx, coachID)
	if err != nil {
		return err
	}
	overrides, err := s.sessionRepo.ListOverrides(ctx, coachID, dateStart.Format("2006-01-02"), dateStart.Format("2006-01-02"))
	if err != nil {
		return err
	}

	if !isWithinAvailabilityWindow(scheduledAt, durationMinutes, availability, overrides) {
		return ErrOutsideAvailability
	}

	endsAt := scheduledAt.Add(time.Duration(durationMinutes) * time.Minute)
	conflict, err := s.sessionRepo.HasCoachConflict(ctx, coachID, scheduledAt, endsAt, nil)
	if err != nil {
		return err
	}
	if conflict {
		return ErrSessionConflict
	}

	// Ensure the requested slot lies inside the requested booking day range.
	if scheduledAt.Before(dateStart) || scheduledAt.After(dateEnd) {
		return ErrOutsideAvailability
	}

	return nil
}

func (s *SessionService) resolveBookedBy(ctx context.Context, userID, coachID, clientUserID uint) (string, error) {
	if userID == clientUserID {
		return "client", nil
	}

	coachProfile, err := s.coachRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrSessionForbidden
		}
		return "", err
	}
	if coachProfile.ID != coachID {
		return "", ErrSessionForbidden
	}
	return "coach", nil
}

func (s *SessionService) getSessionForUser(ctx context.Context, userID, sessionID uint) (*models.Session, error) {
	session, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	if resolveSessionActor(session, userID) == "" {
		return nil, ErrSessionForbidden
	}
	return session, nil
}

func (s *SessionService) getCoachProfile(ctx context.Context, userID uint) (*models.CoachProfile, error) {
	coach, err := s.coachRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCoachProfileNotFound
		}
		return nil, err
	}
	return coach, nil
}

func buildValidatedAvailabilitySlots(coachID uint, inputs []AvailabilitySlotInput) ([]models.CoachAvailability, error) {
	slots := make([]models.CoachAvailability, 0, len(inputs))
	type minuteWindow struct {
		start int
		end   int
	}
	dayWindows := map[int][]minuteWindow{}

	for i := range inputs {
		if inputs[i].DayOfWeek < 0 || inputs[i].DayOfWeek > 6 {
			return nil, ErrAvailabilitySlotInvalid
		}

		startMin, endMin, err := parseTimeRange(inputs[i].StartTime, inputs[i].EndTime)
		if err != nil {
			return nil, err
		}

		active := true
		if inputs[i].IsActive != nil {
			active = *inputs[i].IsActive
		}

		if active {
			candidate := minuteWindow{start: startMin, end: endMin}
			for _, existing := range dayWindows[inputs[i].DayOfWeek] {
				if rangesOverlap(existing.start, existing.end, candidate.start, candidate.end) {
					return nil, ErrAvailabilitySlotInvalid
				}
			}
			dayWindows[inputs[i].DayOfWeek] = append(dayWindows[inputs[i].DayOfWeek], candidate)
		}

		slots = append(slots, models.CoachAvailability{
			CoachID:   coachID,
			DayOfWeek: inputs[i].DayOfWeek,
			StartTime: normalizeHHMM(inputs[i].StartTime),
			EndTime:   normalizeHHMM(inputs[i].EndTime),
			IsActive:  active,
		})
	}

	return slots, nil
}

func buildBookableSlots(
	startDate time.Time,
	endDate time.Time,
	coachID uint,
	sessionTypeID *uint,
	durationMinutes int,
	availability []models.CoachAvailability,
	overrides []models.CoachAvailabilityOverride,
	sessions []models.Session,
) []BookableSlot {
	overrideByDate := map[string][]models.CoachAvailabilityOverride{}
	for i := range overrides {
		overrideByDate[overrides[i].Date] = append(overrideByDate[overrides[i].Date], overrides[i])
	}

	busyByDate := map[string][]timeRange{}
	for i := range sessions {
		if sessions[i].Status != "scheduled" {
			continue
		}
		start := sessions[i].ScheduledAt.UTC()
		end := start.Add(time.Duration(sessions[i].DurationMinutes) * time.Minute)
		key := start.Format("2006-01-02")
		busyByDate[key] = append(busyByDate[key], timeRange{start: start, end: end})
	}

	for key := range busyByDate {
		sort.Slice(busyByDate[key], func(i, j int) bool {
			return busyByDate[key][i].start.Before(busyByDate[key][j].start)
		})
	}

	nowUTC := time.Now().UTC()
	var slots []BookableSlot

	for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
		windows := windowsForDate(current, availability, overrideByDate[current.Format("2006-01-02")])
		if len(windows) == 0 {
			continue
		}

		dayBusy := busyByDate[current.Format("2006-01-02")]
		for _, window := range windows {
			for minute := window.start; minute+durationMinutes <= window.end; minute += slotStepMinutes {
				startAt := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC).Add(time.Duration(minute) * time.Minute)
				endAt := startAt.Add(time.Duration(durationMinutes) * time.Minute)

				if endAt.Before(nowUTC) {
					continue
				}
				if hasBusyConflict(startAt, endAt, dayBusy) {
					continue
				}

				slots = append(slots, BookableSlot{
					StartAt:         startAt,
					EndAt:           endAt,
					DurationMinutes: durationMinutes,
					CoachID:         coachID,
					SessionTypeID:   sessionTypeID,
				})
			}
		}
	}

	return slots
}

func isWithinAvailabilityWindow(
	scheduledAt time.Time,
	durationMinutes int,
	availability []models.CoachAvailability,
	overrides []models.CoachAvailabilityOverride,
) bool {
	date := time.Date(scheduledAt.Year(), scheduledAt.Month(), scheduledAt.Day(), 0, 0, 0, 0, time.UTC)
	windows := windowsForDate(date, availability, overrides)
	if len(windows) == 0 {
		return false
	}

	startMinute := scheduledAt.Hour()*60 + scheduledAt.Minute()
	endMinute := startMinute + durationMinutes
	for _, window := range windows {
		if startMinute >= window.start && endMinute <= window.end {
			return true
		}
	}
	return false
}

type minuteWindow struct {
	start int
	end   int
}

type timeRange struct {
	start time.Time
	end   time.Time
}

func windowsForDate(
	date time.Time,
	availability []models.CoachAvailability,
	overrides []models.CoachAvailabilityOverride,
) []minuteWindow {
	if len(overrides) > 0 {
		blocksDate := false
		windows := make([]minuteWindow, 0, len(overrides))
		for i := range overrides {
			if !overrides[i].IsAvailable {
				blocksDate = true
				continue
			}
			if overrides[i].StartTime == nil || overrides[i].EndTime == nil {
				continue
			}
			start, end, err := parseTimeRange(*overrides[i].StartTime, *overrides[i].EndTime)
			if err != nil {
				continue
			}
			windows = append(windows, minuteWindow{start: start, end: end})
		}
		if blocksDate {
			return nil
		}
		return mergeWindows(windows)
	}

	dayOfWeek := int(date.Weekday())
	windows := make([]minuteWindow, 0)
	for i := range availability {
		if !availability[i].IsActive || availability[i].DayOfWeek != dayOfWeek {
			continue
		}
		start, end, err := parseTimeRange(availability[i].StartTime, availability[i].EndTime)
		if err != nil {
			continue
		}
		windows = append(windows, minuteWindow{start: start, end: end})
	}
	return mergeWindows(windows)
}

func mergeWindows(windows []minuteWindow) []minuteWindow {
	if len(windows) <= 1 {
		return windows
	}

	sort.Slice(windows, func(i, j int) bool {
		if windows[i].start == windows[j].start {
			return windows[i].end < windows[j].end
		}
		return windows[i].start < windows[j].start
	})

	merged := make([]minuteWindow, 0, len(windows))
	current := windows[0]
	for i := 1; i < len(windows); i++ {
		if windows[i].start <= current.end {
			if windows[i].end > current.end {
				current.end = windows[i].end
			}
			continue
		}
		merged = append(merged, current)
		current = windows[i]
	}
	merged = append(merged, current)
	return merged
}

func hasBusyConflict(startAt, endAt time.Time, busy []timeRange) bool {
	for i := range busy {
		if startAt.Before(busy[i].end) && busy[i].start.Before(endAt) {
			return true
		}
	}
	return false
}

func resolveSessionActor(session *models.Session, userID uint) string {
	if session.Coach.UserID == userID {
		return "coach"
	}
	if session.Client.UserID == userID {
		return "client"
	}
	return ""
}

func parseDateRange(startRaw, endRaw string, defaultDays int) (time.Time, time.Time, error) {
	var (
		startDate time.Time
		endDate   time.Time
		err       error
	)

	if strings.TrimSpace(startRaw) == "" {
		now := time.Now().UTC()
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		startDate, err = parseDateOnly(startRaw)
		if err != nil {
			return time.Time{}, time.Time{}, ErrInvalidDateFormat
		}
	}

	if strings.TrimSpace(endRaw) == "" {
		endDate = startDate.AddDate(0, 0, defaultDays)
	} else {
		endDate, err = parseDateOnly(endRaw)
		if err != nil {
			return time.Time{}, time.Time{}, ErrInvalidDateFormat
		}
	}

	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}

	rangeDays := int(endDate.Sub(startDate).Hours() / 24)
	if rangeDays > maxRangeDays {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}

	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	return startDate, endDate, nil
}

func parseDateOnly(raw string) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(raw), time.UTC)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
}

func parseOptionalTimeRange(startRaw *string, endRaw *string) (string, string, error) {
	if startRaw == nil || endRaw == nil {
		return "", "", ErrAvailabilitySlotInvalid
	}
	start, end, err := parseTimeRange(*startRaw, *endRaw)
	if err != nil {
		return "", "", err
	}
	return formatMinuteToHHMM(start), formatMinuteToHHMM(end), nil
}

func parseTimeRange(startRaw, endRaw string) (int, int, error) {
	start, err := parseHHMM(startRaw)
	if err != nil {
		return 0, 0, ErrAvailabilitySlotInvalid
	}
	end, err := parseHHMM(endRaw)
	if err != nil {
		return 0, 0, ErrAvailabilitySlotInvalid
	}
	if end <= start {
		return 0, 0, ErrAvailabilitySlotInvalid
	}
	return start, end, nil
}

func parseHHMM(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := time.Parse("15:04", raw)
	if err != nil {
		return 0, err
	}
	return parsed.Hour()*60 + parsed.Minute(), nil
}

func normalizeHHMM(raw string) string {
	minute, err := parseHHMM(raw)
	if err != nil {
		return raw
	}
	return formatMinuteToHHMM(minute)
}

func formatMinuteToHHMM(minute int) string {
	hour := minute / 60
	min := minute % 60
	return strconv.Itoa(hour/10) + strconv.Itoa(hour%10) + ":" + strconv.Itoa(min/10) + strconv.Itoa(min%10)
}

func isValidSessionDuration(minutes int) bool {
	if minutes < 15 || minutes > 240 {
		return false
	}
	return minutes%5 == 0
}

func rangesOverlap(startA, endA, startB, endB int) bool {
	return startA < endB && startB < endA
}

func trimSessionPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
