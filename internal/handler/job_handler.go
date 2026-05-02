package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/dto"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/usecase"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/jwt"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/response"
	"gorm.io/gorm"
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
		if n, err := strconv.Atoi(pageStr); err == nil && n > 0 {
			page = n
		}
	}
	limit := 10
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
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

func (h *JobHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}
	job, cat, err := h.uc.GetByID(id)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidID) {
			response.Error(w, http.StatusBadRequest, "invalid job id")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "job not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	catDTO := dto.CategoryDTO{ID: job.CategoryID}
	if cat != nil {
		catDTO.Name = cat.Name
	}
	emp := dto.EmployerDTO{}
	if job.Employer != nil {
		emp.ID = job.Employer.ID.String()
		emp.CompanyName = job.Employer.CompanyName
		emp.LogoURL = job.Employer.LogoURL
		emp.Description = job.Employer.Description
	}
	schedules := make([]dto.JobScheduleDTO, 0, len(job.Schedules))
	for _, s := range job.Schedules {
		schedules = append(schedules, dto.JobScheduleDTO{Day: dayName(s.Day), StartTime: s.StartTime, EndTime: s.EndTime})
	}

	out := dto.JobResponse{
		ID:               job.ID.String(),
		Title:            job.Title,
		Description:      job.Description,
		Type:             string(job.Type),
		Status:           string(job.Status),
		Salary:           job.Salary,
		Location:         job.Location,
		Category:         catDTO,
		Employer:         emp,
		Schedules:        schedules,
		WorkHoursPerWeek: 0,
		CreatedAt:        job.CreatedAt.Format(time.RFC3339),
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": out})
}

func extractAuth(r *http.Request) (string, string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return "", "", fmt.Errorf("missing authorization token")
	}
	tkn := strings.TrimPrefix(auth, "Bearer ")
	claims, err := jwt.ParseToken(tkn)
	if err != nil {
		return "", "", fmt.Errorf("invalid token")
	}
	roleV, ok := claims["role"]
	if !ok {
		return "", "", fmt.Errorf("invalid token")
	}
	roleStr, _ := roleV.(string)
	sub, ok := claims["sub"]
	if !ok {
		return "", "", fmt.Errorf("invalid token")
	}
	subStr, _ := sub.(string)
	return subStr, roleStr, nil
}

func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "employer" {
		response.Error(w, http.StatusForbidden, "only employers can create jobs")
		return
	}
	subStr := userID

	var req dto.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Title == "" || req.CategoryID == 0 {
		response.Error(w, http.StatusBadRequest, "missing required fields")
		return
	}
	job, err := h.uc.Create(subStr, req)
	if err != nil {
		if strings.Contains(err.Error(), "category not found") {
			response.Error(w, http.StatusBadRequest, "category not found")
			return
		}
		if strings.Contains(err.Error(), "user is not employer") {
			response.Error(w, http.StatusForbidden, "only employers can create jobs")
			return
		}
		if strings.Contains(err.Error(), "invalid day") {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":         job.ID.String(),
			"title":      job.Title,
			"status":     string(job.Status),
			"created_at": job.CreatedAt.Format(time.RFC3339),
		},
		"message": "Job berhasil dibuat",
	})
}

func (h *JobHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "employer" {
		response.Error(w, http.StatusForbidden, "only employers can update jobs")
		return
	}
	subStr := userID

	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	var req dto.UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	job, err := h.uc.Update(subStr, id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "job not found")
			return
		}
		if strings.Contains(err.Error(), "user is not employer") {
			response.Error(w, http.StatusForbidden, "only employers can update jobs")
			return
		}
		if errors.Is(err, usecase.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "not allowed")
			return
		}
		if strings.Contains(err.Error(), "invalid day") {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":         job.ID.String(),
			"title":      job.Title,
			"updated_at": job.UpdatedAt.Format(time.RFC3339),
		},
		"message": "Job berhasil diupdate",
	})
}

func (h *JobHandler) GetEmployerJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "employer" {
		response.Error(w, http.StatusForbidden, "only employers can access this resource")
		return
	}
	q := r.URL.Query()
	status := strings.TrimSpace(q.Get("status"))
	page := 1
	if p := strings.TrimSpace(q.Get("page")); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	limit := 10
	if l := strings.TrimSpace(q.Get("limit")); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	jobs, total, err := h.uc.GetEmployerJobs(userID, usecase.JobFilter{Status: status, Page: page, Limit: limit})
	if err != nil {
		if strings.Contains(err.Error(), "user is not employer") {
			response.Error(w, http.StatusForbidden, "only employers can access this resource")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"data": jobs, "meta": map[string]any{"page": page, "limit": limit, "total": total}})
}

func (h *JobHandler) GetWorkerApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can access this resource")
		return
	}

	q := r.URL.Query()
	status := strings.TrimSpace(q.Get("status"))
	page := 1
	if p := strings.TrimSpace(q.Get("page")); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	limit := 10
	if l := strings.TrimSpace(q.Get("limit")); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	apps, cvCats, total, err := h.uc.ListWorkerApplications(userID, usecase.ApplicationFilter{Status: status, Page: page, Limit: limit})
	if err != nil {
		if strings.Contains(err.Error(), "user is not worker") {
			response.Error(w, http.StatusForbidden, "only workers can access this resource")
			return
		}
		if strings.Contains(err.Error(), "worker profile not found") {
			response.Error(w, http.StatusBadRequest, "worker profile not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]dto.WorkerApplicationItemDTO, 0, len(apps))
	for _, a := range apps {
		item := dto.WorkerApplicationItemDTO{
			ID:        a.ID.String(),
			Status:    a.Status,
			CoverNote: a.CoverNote,
			AppliedAt: a.AppliedAt.Format(time.RFC3339),
		}
		if a.Job != nil {
			item.Job.ID = a.Job.ID.String()
			item.Job.Title = a.Job.Title
			item.Job.Type = string(a.Job.Type)
			item.Job.Location = a.Job.Location
			if a.Job.Employer != nil {
				item.Job.Employer.CompanyName = a.Job.Employer.CompanyName
				item.Job.Employer.LogoURL = a.Job.Employer.LogoURL
			}
		}
		if a.CV != nil {
			cat := cvCats[a.CV.CategoryID]
			item.CV = &dto.WorkerApplicationsCVDTo{
				ID:      a.CV.ID.String(),
				FileURL: a.CV.FileURL,
				Category: dto.WorkerApplicationsCVCategoryDTO{
					ID:   a.CV.CategoryID,
					Name: cat.Name,
				},
			}
		}
		out = append(out, item)
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": out,
		"meta": map[string]any{"page": page, "limit": limit, "total": total},
	})
}

func (h *JobHandler) GetEmployerApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "employer" {
		response.Error(w, http.StatusForbidden, "only employers can access this resource")
		return
	}

	q := r.URL.Query()
	status := strings.TrimSpace(q.Get("status"))
	page := 1
	if p := strings.TrimSpace(q.Get("page")); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	limit := 10
	if l := strings.TrimSpace(q.Get("limit")); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	apps, cvCats, total, err := h.uc.ListEmployerApplications(userID, usecase.ApplicationFilter{Status: status, Page: page, Limit: limit})
	if err != nil {
		if strings.Contains(err.Error(), "user is not employer") {
			response.Error(w, http.StatusForbidden, "only employers can access this resource")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]dto.EmployerApplicationItemDTO, 0, len(apps))
	for _, a := range apps {
		item := dto.EmployerApplicationItemDTO{
			ID:        a.ID.String(),
			Status:    a.Status,
			AppliedAt: a.AppliedAt.Format(time.RFC3339),
		}
		if a.Job != nil {
			item.Job.ID = a.Job.ID.String()
			item.Job.Title = a.Job.Title
		}
		if a.Worker != nil {
			item.Worker.ID = a.Worker.ID.String()
			item.Worker.FullName = a.Worker.FullName
			item.Worker.PhotoURL = a.Worker.PhotoURL
		}
		if a.CV != nil {
			cat := cvCats[a.CV.CategoryID]
			item.CV = &dto.EmployerApplicationsCVDTO{
				ID:      a.CV.ID.String(),
				FileURL: a.CV.FileURL,
				Category: dto.EmployerApplicationsCVCategoryDTO{
					ID:   a.CV.CategoryID,
					Name: cat.Name,
				},
			}
		}
		out = append(out, item)
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": out,
		"meta": map[string]any{"page": page, "limit": limit, "total": total},
	})
}

func (h *JobHandler) GetEmployerJobApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "employer" {
		response.Error(w, http.StatusForbidden, "only employers can access this resource")
		return
	}

	p := strings.TrimPrefix(r.URL.Path, "/employer/jobs/")
	p = strings.Trim(p, "/")
	parts := strings.Split(p, "/")
	if len(parts) != 2 || parts[1] != "applications" || parts[0] == "" {
		response.Error(w, http.StatusBadRequest, "invalid path")
		return
	}
	jobID := parts[0]

	q := r.URL.Query()
	status := strings.TrimSpace(q.Get("status"))
	page := 1
	if p := strings.TrimSpace(q.Get("page")); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	limit := 10
	if l := strings.TrimSpace(q.Get("limit")); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	apps, cvCats, total, err := h.uc.ListEmployerJobApplications(userID, jobID, usecase.ApplicationFilter{Status: status, Page: page, Limit: limit})
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidID) {
			response.Error(w, http.StatusBadRequest, "invalid job id")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "job not found")
			return
		}
		if errors.Is(err, usecase.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "not allowed")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]dto.EmployerJobApplicationItemDTO, 0, len(apps))
	for _, a := range apps {
		item := dto.EmployerJobApplicationItemDTO{
			ID:        a.ID.String(),
			Status:    a.Status,
			CoverNote: a.CoverNote,
			AppliedAt: a.AppliedAt.Format(time.RFC3339),
		}
		if a.Worker != nil {
			item.Worker.ID = a.Worker.ID.String()
			item.Worker.FullName = a.Worker.FullName
			item.Worker.PhotoURL = a.Worker.PhotoURL
			item.Worker.Bio = a.Worker.Bio
			item.Worker.Skills = a.Worker.Skills
		}
		if a.CV != nil {
			cat := cvCats[a.CV.CategoryID]
			item.CV = &dto.EmployerApplicationsCVDTO{
				ID:      a.CV.ID.String(),
				FileURL: a.CV.FileURL,
				Category: dto.EmployerApplicationsCVCategoryDTO{
					ID:   a.CV.CategoryID,
					Name: cat.Name,
				},
			}
		}
		out = append(out, item)
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": out,
		"meta": map[string]any{"page": page, "limit": limit, "total": total},
	})
}

func (h *JobHandler) PatchEmployerApplicationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "employer" {
		response.Error(w, http.StatusForbidden, "only employers can access this resource")
		return
	}

	p := strings.TrimPrefix(r.URL.Path, "/employer/applications/")
	p = strings.Trim(p, "/")
	parts := strings.Split(p, "/")
	if len(parts) != 2 || parts[1] != "status" || parts[0] == "" {
		response.Error(w, http.StatusBadRequest, "invalid path")
		return
	}
	appID := parts[0]

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	app, err := h.uc.UpdateEmployerApplicationStatus(userID, appID, req.Status)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidStatus) {
			response.Error(w, http.StatusBadRequest, "invalid status")
			return
		}
		if errors.Is(err, usecase.ErrInvalidID) {
			response.Error(w, http.StatusBadRequest, "invalid application id")
			return
		}
		if errors.Is(err, usecase.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "not allowed")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "application not found")
			return
		}
		if strings.Contains(err.Error(), "user is not employer") {
			response.Error(w, http.StatusForbidden, "only employers can access this resource")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "Status lamaran berhasil diubah",
		"data": map[string]any{
			"id":         app.ID.String(),
			"status":     app.Status,
			"updated_at": app.UpdatedAt.Format(time.RFC3339),
		},
	})
}

func (h *JobHandler) DeleteWorkerApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can access this resource")
		return
	}

	appID := strings.TrimPrefix(r.URL.Path, "/worker/applications/")
	appID = strings.Trim(appID, "/")
	if appID == "" || strings.Contains(appID, "/") {
		response.Error(w, http.StatusBadRequest, "invalid path")
		return
	}

	if err := h.uc.WithdrawWorkerApplication(userID, appID); err != nil {
		if errors.Is(err, usecase.ErrInvalidID) {
			response.Error(w, http.StatusBadRequest, "invalid application id")
			return
		}
		if errors.Is(err, usecase.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "not allowed")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "application not found")
			return
		}
		if strings.Contains(err.Error(), "worker profile not found") {
			response.Error(w, http.StatusBadRequest, "worker profile not found")
			return
		}
		if strings.Contains(err.Error(), "user is not worker") {
			response.Error(w, http.StatusForbidden, "only workers can access this resource")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"message": "Lamaran berhasil ditarik"})
}

func (h *JobHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "employer" {
		response.Error(w, http.StatusForbidden, "only employers can delete jobs")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}
	if err := h.uc.Delete(userID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "job not found")
			return
		}
		if errors.Is(err, usecase.ErrForbidden) || strings.Contains(err.Error(), "user is not employer") {
			response.Error(w, http.StatusForbidden, "not allowed")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"message": "Job berhasil dihapus"})
}

func (h *JobHandler) Apply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can apply to jobs")
		return
	}

	p := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if !strings.HasSuffix(p, "/apply") {
		response.Error(w, http.StatusBadRequest, "invalid path")
		return
	}
	id := strings.TrimSuffix(p, "/apply")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	var req struct {
		CVID      string `json:"cv_id"`
		CoverNote string `json:"cover_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.CVID == "" {
		response.Error(w, http.StatusBadRequest, "missing cv_id")
		return
	}

	app, err := h.uc.Apply(userID, id, req.CVID, req.CoverNote)
	if err != nil {
		if strings.Contains(err.Error(), "invalid cv id") || strings.Contains(err.Error(), "cv not found") || strings.Contains(err.Error(), "cv does not belong") {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "already applied") {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "worker profile not found") {
			response.Error(w, http.StatusBadRequest, "worker profile not found")
			return
		}
		if strings.Contains(err.Error(), "user is not worker") {
			response.Error(w, http.StatusForbidden, "only workers can apply to jobs")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, usecase.ErrInvalidID) {
			response.Error(w, http.StatusNotFound, "job not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"message": "Lamaran berhasil dikirim",
		"data": map[string]any{
			"id":         app.ID.String(),
			"job_id":     app.JobID.String(),
			"status":     app.Status,
			"applied_at": app.AppliedAt.Format(time.RFC3339),
		},
	})
}
