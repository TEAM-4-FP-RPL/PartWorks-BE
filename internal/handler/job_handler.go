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

func (h *JobHandler) DeleteJob(c *gin.Context) {
    idParam := c.Param("id")
    jobID, err := uuid.Parse(idParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "format ID pekerjaan tidak valid"})
        return
    }

    val, _ := c.Get("user_id")
    employerID, _ := uuid.Parse(val.(string))

    err = h.usecase.DeleteJob(c.Request.Context(), jobID, employerID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus pekerjaan: " + err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "pekerjaan berhasil dihapus"})
}