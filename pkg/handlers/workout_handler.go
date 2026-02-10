package handlers

import (
	"chalk-api/pkg/services"
	"chalk-api/pkg/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WorkoutHandler struct {
	workoutService *services.WorkoutService
}

func NewWorkoutHandler(workoutService *services.WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{workoutService: workoutService}
}

func (h *WorkoutHandler) CreateTemplate(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.CreateWorkoutTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	template, err := h.workoutService.CreateTemplate(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template"})
		}
		return
	}

	c.JSON(http.StatusCreated, template)
}

func (h *WorkoutHandler) ListMyTemplates(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := parseQueryInt(c.DefaultQuery("limit", "20"), 20)
	offset := parseQueryInt(c.DefaultQuery("offset", "0"), 0)

	templates, total, err := h.workoutService.ListMyTemplates(c.Request.Context(), userID, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list templates"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   templates,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *WorkoutHandler) GetMyTemplate(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	templateID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	template, err := h.workoutService.GetMyTemplate(c.Request.Context(), userID, templateID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound), errors.Is(err, services.ErrTemplateNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		case errors.Is(err, services.ErrTemplateForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "template does not belong to this coach"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch template"})
		}
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *WorkoutHandler) UpdateMyTemplate(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	templateID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	var input services.UpdateWorkoutTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	template, err := h.workoutService.UpdateMyTemplate(c.Request.Context(), userID, templateID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound), errors.Is(err, services.ErrTemplateNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		case errors.Is(err, services.ErrTemplateForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "template does not belong to this coach"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update template"})
		}
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *WorkoutHandler) AssignWorkout(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input services.AssignWorkoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	workout, err := h.workoutService.AssignTemplateToClient(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCoachProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "coach profile not found"})
		case errors.Is(err, services.ErrTemplateNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		case errors.Is(err, services.ErrTemplateForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "template does not belong to this coach"})
		case errors.Is(err, services.ErrClientProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "client profile not found"})
		case errors.Is(err, services.ErrClientProfileForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "client profile does not belong to this coach"})
		case errors.Is(err, services.ErrInvalidScheduledDate):
			c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_date must be YYYY-MM-DD"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign workout"})
		}
		return
	}

	c.JSON(http.StatusCreated, workout)
}

func (h *WorkoutHandler) ListMyWorkouts(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := parseQueryInt(c.DefaultQuery("limit", "20"), 20)
	offset := parseQueryInt(c.DefaultQuery("offset", "0"), 0)

	workouts, total, err := h.workoutService.ListMyWorkouts(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list workouts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   workouts,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *WorkoutHandler) GetMyWorkout(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	workoutID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
		return
	}

	workout, err := h.workoutService.GetMyWorkout(c.Request.Context(), userID, workoutID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrWorkoutNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		case errors.Is(err, services.ErrWorkoutForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "workout does not belong to this user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workout"})
		}
		return
	}

	c.JSON(http.StatusOK, workout)
}

func (h *WorkoutHandler) StartMyWorkout(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	workoutID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
		return
	}

	workout, err := h.workoutService.StartMyWorkout(c.Request.Context(), userID, workoutID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrWorkoutNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		case errors.Is(err, services.ErrWorkoutForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "workout does not belong to this user"})
		case errors.Is(err, services.ErrInvalidWorkoutState):
			c.JSON(http.StatusConflict, gin.H{"error": "workout is already finalized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start workout"})
		}
		return
	}

	c.JSON(http.StatusOK, workout)
}

func (h *WorkoutHandler) CompleteMyWorkout(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	workoutID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
		return
	}

	workout, err := h.workoutService.CompleteMyWorkout(c.Request.Context(), userID, workoutID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrWorkoutNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		case errors.Is(err, services.ErrWorkoutForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "workout does not belong to this user"})
		case errors.Is(err, services.ErrInvalidWorkoutState):
			c.JSON(http.StatusConflict, gin.H{"error": "workout is already finalized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete workout"})
		}
		return
	}

	c.JSON(http.StatusOK, workout)
}

func (h *WorkoutHandler) MarkExerciseCompleted(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	exerciseID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout exercise id"})
		return
	}

	exercise, err := h.workoutService.MarkMyExerciseCompleted(c.Request.Context(), userID, exerciseID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrWorkoutExerciseNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout exercise not found"})
		case errors.Is(err, services.ErrWorkoutNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		case errors.Is(err, services.ErrWorkoutForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "workout does not belong to this user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark exercise completed"})
		}
		return
	}

	c.JSON(http.StatusOK, exercise)
}

func (h *WorkoutHandler) SkipExercise(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	exerciseID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout exercise id"})
		return
	}

	var input services.SkipWorkoutExerciseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	exercise, err := h.workoutService.SkipMyExercise(c.Request.Context(), userID, exerciseID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrWorkoutExerciseNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout exercise not found"})
		case errors.Is(err, services.ErrWorkoutNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		case errors.Is(err, services.ErrWorkoutForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "workout does not belong to this user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to skip exercise"})
		}
		return
	}

	c.JSON(http.StatusOK, exercise)
}

func (h *WorkoutHandler) CreateExerciseLog(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	exerciseID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout exercise id"})
		return
	}

	var input services.CreateWorkoutLogInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	logEntry, err := h.workoutService.CreateMyExerciseLog(c.Request.Context(), userID, exerciseID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrWorkoutExerciseNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout exercise not found"})
		case errors.Is(err, services.ErrWorkoutNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		case errors.Is(err, services.ErrWorkoutForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "workout does not belong to this user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workout log"})
		}
		return
	}

	c.JSON(http.StatusCreated, logEntry)
}

func (h *WorkoutHandler) UpdateWorkoutLog(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	logID, valid := parseUintParam(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout log id"})
		return
	}

	var input services.UpdateWorkoutLogInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	logEntry, err := h.workoutService.UpdateMyWorkoutLog(c.Request.Context(), userID, logID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrWorkoutLogNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout log not found"})
		case errors.Is(err, services.ErrWorkoutExerciseNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout exercise not found"})
		case errors.Is(err, services.ErrWorkoutNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		case errors.Is(err, services.ErrWorkoutForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "workout does not belong to this user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update workout log"})
		}
		return
	}

	c.JSON(http.StatusOK, logEntry)
}

func parseUintParam(raw string) (uint, bool) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func parseQueryInt(raw string, fallback int) int {
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}
