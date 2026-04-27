package handler
import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/dto"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/middleware"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/usecase"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/response"
	"github.com/google/uuid"
)

type JobHandler struct {
	uc *usecase.JobUsecase
}
func NewJobHandler(uc *usecase.JobUsecase) *JobHandler { return &JobHandler{uc: uc} }

func dayName(n int) string {
	switch n {
	case 1:
		return "monday"
	case 2:
		return "tuesday"
	case 3:
		return "wednesday"
	case 4:
		return "thursday"
	case 5:
		return "friday"
	case 6:
		return "saturday"
	case 7:
		return "sunday"
	default:
		return "unknown"
	}
}

func mapDayToInt(day string) int {
	switch strings.ToLower(day) {
	case "monday":
		return 1
	case "tuesday":
		return 2
	case "wednesday":
		return 3
	case "thursday":
		return 4
	case "friday":
		return 5
	case "saturday":
		return 6
	case "sunday":
		return 7
	default:
		return 0
	}
}

func (h *JobHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	q := r.URL.Query()
	search := strings.TrimSpace(q.Get("search"))
	catIDStr := strings.TrimSpace(q.Get("category_id"))
	typ := strings.TrimSpace(q.Get("type"))
	location := strings.TrimSpace(q.Get("location"))
	status := strings.TrimSpace(q.Get("status"))
	pageStr := strings.TrimSpace(q.Get("page"))
	limitStr := strings.TrimSpace(q.Get("limit"))

	if status == "" {
		status = string("open")
	}
	var catID *int
	if catIDStr != "" {
		if n, err := strconv.Atoi(catIDStr); err == nil {
			catID = &n
		}
	}
	page := 1
	if pageStr != "" {
		if n, err := strconv.Atoi(pageStr); err == nil && n > 0 { page = n }
	}
	limit := 10
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 { limit = n }
	}

	jobs, total, err := h.uc.GetAll(usecase.JobFilter{
		Search:     search,
		CategoryID: catID,
		Type:       typ,
		Location:   location,
		Status:     status,
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]dto.JobResponse, 0, len(jobs))
	for _, j := range jobs {
		cat := dto.CategoryDTO{ID: j.CategoryID}
		emp := dto.EmployerDTO{}
		if j.Employer != nil {
			emp.ID = j.Employer.ID.String()
			emp.CompanyName = j.Employer.CompanyName
			emp.LogoURL = j.Employer.LogoURL
		}
		schedules := make([]dto.JobScheduleDTO, 0, len(j.Schedules))
		for _, s := range j.Schedules {
			schedules = append(schedules, dto.JobScheduleDTO{Day: dayName(s.Day), StartTime: s.StartTime, EndTime: s.EndTime})
		}
		out = append(out, dto.JobResponse{
			ID:        j.ID.String(),
			Title:     j.Title,
			Type:      string(j.Type),
			Status:    string(j.Status),
			Salary:    j.Salary,
			Location:  j.Location,
			Category:  cat,
			Employer:  emp,
			Schedules: schedules,
			CreatedAt: j.CreatedAt.Format(time.RFC3339),
		})
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": out,
		"meta": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// userID should be set in context by AuthMiddleware
	val := r.Context().Value(middleware.UserIDKey)
	if val == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userIDStr, ok := val.(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	employerID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	var req dto.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	job := &domain.Job{
		EmployerID:  employerID,
		CategoryID:  req.CategoryID,
		Title:       req.Title,
		Description: req.Description,
		Type:        domain.JobType(req.Type),
		Salary:      req.Salary,
		Location:    req.Location,
	}

	for _, s := range req.Schedules {
		day := mapDayToInt(s.Day)
		if day == 0 {
			response.Error(w, http.StatusBadRequest, "invalid day name: "+s.Day)
			return
		}
		job.Schedules = append(job.Schedules, domain.JobSchedule{
			Day:       day,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
		})
	}

	if err := h.uc.Create(job); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create job")
		return
	}

	res := dto.CreateJobResponse{
		Data: dto.JobDataResponse{
			ID:        job.ID,
			Title:     job.Title,
			Status:    string(job.Status),
			CreatedAt: job.CreatedAt,
		},
		Message: "Job berhasil dibuat",
	}

	response.JSON(w, http.StatusCreated, res)
}
