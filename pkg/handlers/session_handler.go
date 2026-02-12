package handlers

import (
	"chalk-api/pkg/services"
	"chalk-api/pkg/utils"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	sessionService *services.SessionService
}

func NewSessionHandler(sessionService *services.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

func (h *SessionHandler) GetMyAvailability(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slots, err := h.sessionService.GetMyAvailability(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch availability"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": slots})
}

func (h *SessionHandler) SetMyAvailability(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.SetAvailabilityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	slots, err := h.sessionService.SetMyAvailability(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		case errors.Is(err, services.ErrAvailabilitySlotInvalid):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid availability slot payload"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save availability"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": slots})
}

func (h *SessionHandler) CreateAvailabilityOverride(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.CreateAvailabilityOverrideInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	override, err := h.sessionService.CreateAvailabilityOverride(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		case errors.Is(err, services.ErrInvalidDateFormat), errors.Is(err, services.ErrAvailabilitySlotInvalid):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid override payload"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create override"})
		}
		return
	}

	c.JSON(http.StatusCreated, override)
}

func (h *SessionHandler) ListAvailabilityOverrides(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	overrides, err := h.sessionService.ListMyAvailabilityOverrides(
		c.Request.Context(),
		userID,
		c.Query("start"),
		c.Query("end"),
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		case errors.Is(err, services.ErrInvalidDateRange), errors.Is(err, services.ErrInvalidDateFormat):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date range"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch overrides"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": overrides})
}

func (h *SessionHandler) DeleteAvailabilityOverride(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	overrideID, valid := parseUintPathParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid override id"})
		return
	}

	if err := h.sessionService.DeleteMyAvailabilityOverride(c.Request.Context(), userID, overrideID); err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		case errors.Is(err, services.ErrOverrideNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "availability override not found"})
		case errors.Is(err, services.ErrOverrideForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "override does not belong to this coach"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete override"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "availability override deleted"})
}

func (h *SessionHandler) CreateSessionType(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.CreateSessionTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	sessionType, err := h.sessionService.CreateMySessionType(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		case errors.Is(err, services.ErrSessionTypeInvalid):
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		case errors.Is(err, services.ErrInvalidSessionDuration):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid duration_minutes"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session type"})
		}
		return
	}

	c.JSON(http.StatusCreated, sessionType)
}

func (h *SessionHandler) ListSessionTypes(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionTypes, err := h.sessionService.ListMySessionTypes(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch session types"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sessionTypes})
}

func (h *SessionHandler) UpdateSessionType(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionTypeID, valid := parseUintPathParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session type id"})
		return
	}

	var input services.UpdateSessionTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	sessionType, err := h.sessionService.UpdateMySessionType(c.Request.Context(), userID, sessionTypeID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		case errors.Is(err, services.ErrSessionTypeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "session type not found"})
		case errors.Is(err, services.ErrSessionTypeForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "session type does not belong to this coach"})
		case errors.Is(err, services.ErrInvalidSessionDuration):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid duration_minutes"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session type"})
		}
		return
	}

	c.JSON(http.StatusOK, sessionType)
}

func (h *SessionHandler) GetBookableSlots(c *gin.Context) {
	// Keep this protected for now (clients/coaches in app), but no ownership restriction.
	if _, ok := utils.GetUserIDFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	coachID, valid := parseUintPathParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coach id"})
		return
	}

	sessionTypeID, hasSessionType, err := parseOptionalUintQuery(c.Query("session_type_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_type_id"})
		return
	}
	duration, hasDuration, err := parseOptionalIntQuery(c.Query("duration_minutes"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid duration_minutes"})
		return
	}

	var sessionTypeRef *uint
	if hasSessionType {
		sessionTypeRef = &sessionTypeID
	}
	var durationRef *int
	if hasDuration {
		durationRef = &duration
	}

	slots, serviceErr := h.sessionService.GetBookableSlots(
		c.Request.Context(),
		coachID,
		c.Query("start"),
		c.Query("end"),
		sessionTypeRef,
		durationRef,
	)
	if serviceErr != nil {
		switch {
		case errors.Is(serviceErr, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		case errors.Is(serviceErr, services.ErrSessionTypeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "session type not found"})
		case errors.Is(serviceErr, services.ErrSessionTypeForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "session type does not belong to this coach"})
		case errors.Is(serviceErr, services.ErrSessionTypeInactive):
			c.JSON(http.StatusConflict, gin.H{"error": "session type is inactive"})
		case errors.Is(serviceErr, services.ErrInvalidDateRange), errors.Is(serviceErr, services.ErrInvalidDateFormat), errors.Is(serviceErr, services.ErrInvalidSessionDuration):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build bookable slots"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  slots,
		"total": len(slots),
	})
}

func (h *SessionHandler) BookSession(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.BookSessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	session, err := h.sessionService.BookSession(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrClientProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "client profile not found"})
		case errors.Is(err, services.ErrSessionTypeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "session type not found"})
		case errors.Is(err, services.ErrSessionTypeForbidden), errors.Is(err, services.ErrSessionForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "booking is not allowed for this user"})
		case errors.Is(err, services.ErrSessionTypeInactive):
			c.JSON(http.StatusConflict, gin.H{"error": "session type is inactive"})
		case errors.Is(err, services.ErrInvalidScheduledAt), errors.Is(err, services.ErrInvalidSessionDuration):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking payload"})
		case errors.Is(err, services.ErrOutsideAvailability):
			c.JSON(http.StatusConflict, gin.H{"error": "requested time is outside coach availability"})
		case errors.Is(err, services.ErrSessionConflict):
			c.JSON(http.StatusConflict, gin.H{"error": "requested time conflicts with another session"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to book session"})
		}
		return
	}

	c.JSON(http.StatusCreated, session)
}

func (h *SessionHandler) ListMySessions(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessions, err := h.sessionService.ListMySessions(c.Request.Context(), userID, c.Query("start"), c.Query("end"))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidDateRange), errors.Is(err, services.ErrInvalidDateFormat):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date range"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sessions"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

func (h *SessionHandler) ListCoachSessions(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessions, err := h.sessionService.ListCoachSessions(c.Request.Context(), userID, c.Query("start"), c.Query("end"))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		case errors.Is(err, services.ErrInvalidDateRange), errors.Is(err, services.ErrInvalidDateFormat):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date range"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sessions"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

func (h *SessionHandler) CancelSession(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID, valid := parseUintPathParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	var input services.CancelSessionInput
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	session, err := h.sessionService.CancelSession(c.Request.Context(), userID, sessionID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		case errors.Is(err, services.ErrSessionForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "session does not belong to this user"})
		case errors.Is(err, services.ErrSessionStateInvalid):
			c.JSON(http.StatusConflict, gin.H{"error": "session can no longer be cancelled"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel session"})
		}
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *SessionHandler) CompleteSession(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID, valid := parseUintPathParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	session, err := h.sessionService.CompleteSession(c.Request.Context(), userID, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		case errors.Is(err, services.ErrSessionForbidden), errors.Is(err, services.ErrSessionActionForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "only coach can complete this session"})
		case errors.Is(err, services.ErrSessionStateInvalid):
			c.JSON(http.StatusConflict, gin.H{"error": "session is not in a completable state"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete session"})
		}
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *SessionHandler) MarkNoShow(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionID, valid := parseUintPathParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	session, err := h.sessionService.MarkNoShow(c.Request.Context(), userID, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		case errors.Is(err, services.ErrSessionForbidden), errors.Is(err, services.ErrSessionActionForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "only coach can mark no-show"})
		case errors.Is(err, services.ErrSessionStateInvalid):
			c.JSON(http.StatusConflict, gin.H{"error": "session is not in a no-show state"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark no-show"})
		}
		return
	}

	c.JSON(http.StatusOK, session)
}

func parseUintPathParam(raw string) (uint, bool) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func parseOptionalUintQuery(raw string) (uint, bool, error) {
	if raw == "" {
		return 0, false, nil
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		return 0, false, errors.New("invalid unsigned integer")
	}
	return uint(value), true, nil
}

func parseOptionalIntQuery(raw string) (int, bool, error) {
	if raw == "" {
		return 0, false, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, false, errors.New("invalid integer")
	}
	return value, true, nil
}
