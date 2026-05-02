package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/dto"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/usecase"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var dayToInt = map[string]int{
	"monday":    1,
	"tuesday":   2,
	"wednesday": 3,
	"thursday":  4,
	"friday":    5,
	"saturday":  6,
	"sunday":    7,
}

func (h *JobHandler) GetWorkerProfile(w http.ResponseWriter, r *http.Request) {
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can access this resource")
		return
	}

	wp, cats, err := h.uc.GetWorkerProfile(userID)
	if err != nil {
		if strings.Contains(err.Error(), "user is not worker") {
			response.Error(w, http.StatusForbidden, "only workers can access this resource")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	avs := make([]dto.WorkerAvailabilityDTO, 0, len(wp.Availabilities))
	for _, a := range wp.Availabilities {
		avs = append(avs, dto.WorkerAvailabilityDTO{Day: dayName(a.Day), StartTime: a.StartTime, EndTime: a.EndTime})
	}
	cvs := make([]dto.WorkerCVItemDTO, 0, len(wp.CVs))
	for _, cv := range wp.CVs {
		cat := cats[cv.CategoryID]
		cvs = append(cvs, dto.WorkerCVItemDTO{
			ID:      cv.ID.String(),
			FileURL: cv.FileURL,
			Category: dto.CategoryDTO{
				ID:   cv.CategoryID,
				Name: cat.Name,
			},
		})
	}

	out := dto.WorkerProfileDTO{
		ID:             wp.ID.String(),
		FullName:       wp.FullName,
		PhoneNumber:    wp.PhoneNumber,
		Bio:            wp.Bio,
		Skills:         wp.Skills,
		PhotoURL:       wp.PhotoURL,
		Availabilities: avs,
		CVs:            cvs,
	}
	response.JSON(w, http.StatusOK, map[string]any{"data": out})
}

func (h *JobHandler) PutWorkerProfile(w http.ResponseWriter, r *http.Request) {
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can access this resource")
		return
	}

	var req dto.UpdateWorkerProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	wp, err := h.uc.UpdateWorkerProfile(userID, usecase.WorkerProfileUpdate{
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Bio:         req.Bio,
		Skills:      req.Skills,
		PhotoURL:    req.PhotoURL,
	})
	if err != nil {
		if strings.Contains(err.Error(), "user is not worker") {
			response.Error(w, http.StatusForbidden, "only workers can access this resource")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "Profil berhasil diupdate",
		"data": map[string]any{
			"id":         wp.ID.String(),
			"full_name":  wp.FullName,
			"updated_at": wp.UpdatedAt.Format(time.RFC3339),
		},
	})
}

func (h *JobHandler) GetWorkerAvailability(w http.ResponseWriter, r *http.Request) {
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can access this resource")
		return
	}

	avs, err := h.uc.ListWorkerAvailabilities(userID)
	if err != nil {
		if strings.Contains(err.Error(), "user is not worker") {
			response.Error(w, http.StatusForbidden, "only workers can access this resource")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]dto.WorkerAvailabilityDTO, 0, len(avs))
	for _, a := range avs {
		out = append(out, dto.WorkerAvailabilityDTO{Day: dayName(a.Day), StartTime: a.StartTime, EndTime: a.EndTime})
	}
	response.JSON(w, http.StatusOK, map[string]any{"data": out})
}

func (h *JobHandler) PutWorkerAvailability(w http.ResponseWriter, r *http.Request) {
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can access this resource")
		return
	}

	var req dto.UpdateWorkerAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	items := make([]usecase.AvailabilityItem, 0, len(req.Availabilities))
	for _, it := range req.Availabilities {
		d := dayToInt[strings.ToLower(strings.TrimSpace(it.Day))]
		if d == 0 {
			response.Error(w, http.StatusBadRequest, "invalid day")
			return
		}
		items = append(items, usecase.AvailabilityItem{Day: d, StartTime: it.StartTime, EndTime: it.EndTime})
	}

	avs, err := h.uc.UpdateWorkerAvailabilities(userID, items)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidAvailability) {
			response.Error(w, http.StatusBadRequest, "invalid availability")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]dto.WorkerAvailabilityDTO, 0, len(avs))
	for _, a := range avs {
		out = append(out, dto.WorkerAvailabilityDTO{Day: dayName(a.Day), StartTime: a.StartTime, EndTime: a.EndTime})
	}
	response.JSON(w, http.StatusOK, map[string]any{"message": "Availability berhasil diupdate", "data": out})
}

func (h *JobHandler) ListWorkerCVs(w http.ResponseWriter, r *http.Request) {
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can access this resource")
		return
	}

	cvs, cats, err := h.uc.ListWorkerCVs(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]dto.WorkerCVItemDTO, 0, len(cvs))
	for _, cv := range cvs {
		cat := cats[cv.CategoryID]
		out = append(out, dto.WorkerCVItemDTO{
			ID:        cv.ID.String(),
			FileURL:   cv.FileURL,
			CreatedAt: cv.CreatedAt.Format(time.RFC3339),
			Category:  dto.CategoryDTO{ID: cv.CategoryID, Name: cat.Name},
		})
	}
	response.JSON(w, http.StatusOK, map[string]any{"data": out})
}

func (h *JobHandler) UploadWorkerCV(w http.ResponseWriter, r *http.Request) {
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can access this resource")
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid multipart")
		return
	}
	catIDStr := strings.TrimSpace(r.FormValue("category_id"))
	catID, err := strconv.Atoi(catIDStr)
	if err != nil || catID <= 0 {
		response.Error(w, http.StatusBadRequest, "invalid category_id")
		return
	}
	f, fh, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "missing file")
		return
	}
	defer f.Close()
	_ = fh

	buf := make([]byte, 4)
	n, _ := io.ReadFull(f, buf)
	if n < 4 || string(buf) != "%PDF" {
		response.Error(w, http.StatusBadRequest, "file must be pdf")
		return
	}

	cvID := uuid.New()
	if err := os.MkdirAll(filepath.Join("uploads", "cvs"), 0o755); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	name := fmt.Sprintf("%s.pdf", cvID.String())
	localPath := filepath.Join("uploads", "cvs", name)
	outFile, err := os.Create(localPath)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer outFile.Close()
	if _, err := outFile.Write(buf); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if _, err := io.Copy(outFile, f); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	fileURL := "/uploads/cvs/" + name
	cv, cat, err := h.uc.CreateWorkerCV(userID, cvID, catID, fileURL)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusBadRequest, "category not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"message": "CV berhasil diupload",
		"data": dto.WorkerCVItemDTO{
			ID:        cv.ID.String(),
			FileURL:   cv.FileURL,
			CreatedAt: cv.CreatedAt.Format(time.RFC3339),
			Category:  dto.CategoryDTO{ID: cv.CategoryID, Name: cat.Name},
		},
	})
}

func (h *JobHandler) DeleteWorkerCV(w http.ResponseWriter, r *http.Request) {
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can access this resource")
		return
	}

	cvID := strings.TrimPrefix(r.URL.Path, "/worker/cvs/")
	cvID = strings.Trim(cvID, "/")
	if cvID == "" || strings.Contains(cvID, "/") {
		response.Error(w, http.StatusBadRequest, "invalid path")
		return
	}

	cv, err := h.uc.DeleteWorkerCV(userID, cvID)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidID) {
			response.Error(w, http.StatusBadRequest, "invalid cv id")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "cv not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	if strings.HasPrefix(cv.FileURL, "/uploads/") {
		_ = os.Remove(strings.TrimPrefix(cv.FileURL, "/"))
	}
	response.JSON(w, http.StatusOK, map[string]any{"message": "CV berhasil dihapus"})
}

func (h *JobHandler) GetEmployerProfile(w http.ResponseWriter, r *http.Request) {
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "employer" {
		response.Error(w, http.StatusForbidden, "only employers can access this resource")
		return
	}

	emp, err := h.uc.GetEmployerProfile(userID)
	if err != nil {
		if strings.Contains(err.Error(), "user is not employer") {
			response.Error(w, http.StatusForbidden, "only employers can access this resource")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": dto.EmployerProfileDTO{ID: emp.ID.String(), CompanyName: emp.CompanyName, Description: emp.Description, LogoURL: emp.LogoURL},
	})
}

func (h *JobHandler) PutEmployerProfile(w http.ResponseWriter, r *http.Request) {
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "employer" {
		response.Error(w, http.StatusForbidden, "only employers can access this resource")
		return
	}

	var req dto.UpdateEmployerProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	emp, err := h.uc.UpdateEmployerProfile(userID, usecase.EmployerProfileUpdate{CompanyName: req.CompanyName, Description: req.Description, LogoURL: req.LogoURL})
	if err != nil {
		if strings.Contains(err.Error(), "user is not employer") {
			response.Error(w, http.StatusForbidden, "only employers can access this resource")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "Profil berhasil diupdate",
		"data":    map[string]any{"id": emp.ID.String(), "company_name": emp.CompanyName, "updated_at": emp.UpdatedAt.Format(time.RFC3339)},
	})
}

func (h *JobHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.uc.ListCategories()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]dto.CategoryDTO, 0, len(cats))
	for _, c := range cats {
		out = append(out, dto.CategoryDTO{ID: c.ID, Name: c.Name})
	}
	response.JSON(w, http.StatusOK, map[string]any{"data": out})
}
