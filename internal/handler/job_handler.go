package handler

import (
	"net/http"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JobHandler struct {
	usecase usecase.JobUsecase
}

func NewJobHandler(u usecase.JobUsecase) *JobHandler {
	return &JobHandler{usecase: u}
}

func (h *JobHandler) GetEmployerJobs(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	employerID, err := uuid.Parse(val.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format"})
		return
	}

	jobs, err := h.usecase.GetEmployerJobs(c.Request.Context(), employerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jobs)
}