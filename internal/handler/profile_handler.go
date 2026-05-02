package handler

import (
	"crypto/sha256"
	"encoding/hex"
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

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
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

func fileSHA256Hex(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
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

func (h *JobHandler) PatchWorkerProfile(w http.ResponseWriter, r *http.Request) {
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

func (h *JobHandler) PatchWorkerAvailability(w http.ResponseWriter, r *http.Request) {
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

	type slot struct {
		Idx      int
		CatID    int
		FileKey  string
		CatKey   string
		Local    string
		FileURL  string
		FileID   uuid.UUID
		FileHash string
	}

	slots := make([]slot, 0, 3)
	bulkDetected := false
	for i := 1; i <= 3; i++ {
		catKey := fmt.Sprintf("category_id_%d", i)
		fileKey := fmt.Sprintf("file_%d", i)
		catVal := strings.TrimSpace(r.FormValue(catKey))
		_, filePresent := r.MultipartForm.File[fileKey]
		if catVal != "" || filePresent {
			bulkDetected = true
			if catVal == "" || !filePresent {
				response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s dan %s harus berpasangan", catKey, fileKey))
				return
			}
			catID, err := strconv.Atoi(catVal)
			if err != nil || catID <= 0 {
				response.Error(w, http.StatusBadRequest, fmt.Sprintf("invalid %s", catKey))
				return
			}
			slots = append(slots, slot{Idx: i, CatID: catID, FileKey: fileKey, CatKey: catKey})
		}
	}
	if !bulkDetected {
		catIDStr := strings.TrimSpace(r.FormValue("category_id"))
		catID, err := strconv.Atoi(catIDStr)
		if err != nil || catID <= 0 {
			response.Error(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		if _, ok := r.MultipartForm.File["file"]; !ok {
			response.Error(w, http.StatusBadRequest, "missing file")
			return
		}
		slots = append(slots, slot{Idx: 1, CatID: catID, FileKey: "file", CatKey: "category_id"})
	}
	if len(slots) == 0 {
		response.Error(w, http.StatusBadRequest, "missing file")
		return
	}

	existingCVs, _, err := h.uc.ListWorkerCVs(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if len(existingCVs)+len(slots) > 3 {
		response.Error(w, http.StatusBadRequest, "maksimal 3 CV")
		return
	}

	cats, err := h.uc.ListCategories()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	catOK := make(map[int]struct{}, len(cats))
	for i := range cats {
		catOK[cats[i].ID] = struct{}{}
	}
	for i := range slots {
		if _, ok := catOK[slots[i].CatID]; !ok {
			response.Error(w, http.StatusBadRequest, "category not found")
			return
		}
	}

	existingHashes := make(map[string]struct{})
	for _, cv := range existingCVs {
		if !strings.HasPrefix(cv.FileURL, "/uploads/") {
			continue
		}
		oldPath := strings.TrimPrefix(cv.FileURL, "/")
		oldHash, err := fileSHA256Hex(oldPath)
		if err != nil {
			continue
		}
		existingHashes[oldHash] = struct{}{}
	}

	if err := os.MkdirAll(filepath.Join("uploads", "cvs"), 0o755); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	newHashes := make(map[string]int)
	cleanup := func() {
		for i := range slots {
			if slots[i].Local != "" {
				_ = os.Remove(slots[i].Local)
			}
		}
	}

	createItems := make([]usecase.WorkerCVCreateItem, 0, len(slots))
	for i := range slots {
		f, _, err := r.FormFile(slots[i].FileKey)
		if err != nil {
			cleanup()
			response.Error(w, http.StatusBadRequest, fmt.Sprintf("missing %s", slots[i].FileKey))
			return
		}

		buf := make([]byte, 4)
		n, _ := io.ReadFull(f, buf)
		if n < 4 || string(buf) != "%PDF" {
			_ = f.Close()
			cleanup()
			response.Error(w, http.StatusBadRequest, "file must be pdf")
			return
		}

		cvID := uuid.New()
		name := fmt.Sprintf("%s.pdf", cvID.String())
		localPath := filepath.Join("uploads", "cvs", name)
		outFile, err := os.Create(localPath)
		if err != nil {
			_ = f.Close()
			cleanup()
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}

		hasher := sha256.New()
		mw := io.MultiWriter(outFile, hasher)
		if _, err := mw.Write(buf); err != nil {
			_ = outFile.Close()
			_ = f.Close()
			cleanup()
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}
		if _, err := io.Copy(mw, f); err != nil {
			_ = outFile.Close()
			_ = f.Close()
			cleanup()
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}
		_ = outFile.Close()
		_ = f.Close()

		newHash := hex.EncodeToString(hasher.Sum(nil))
		if _, ok := existingHashes[newHash]; ok {
			slots[i].Local = localPath
			cleanup()
			response.Error(w, http.StatusBadRequest, "cv yang sama sudah pernah diupload")
			return
		}
		if prevIdx, ok := newHashes[newHash]; ok {
			slots[i].Local = localPath
			cleanup()
			response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s duplikat dengan file_%d", slots[i].FileKey, prevIdx))
			return
		}
		newHashes[newHash] = slots[i].Idx

		slots[i].FileID = cvID
		slots[i].Local = localPath
		slots[i].FileURL = "/uploads/cvs/" + name
		slots[i].FileHash = newHash
		createItems = append(createItems, usecase.WorkerCVCreateItem{ID: cvID, CategoryID: slots[i].CatID, FileURL: slots[i].FileURL})
	}

	created, catMap, err := h.uc.CreateWorkerCVsBulk(userID, createItems)
	if err != nil {
		cleanup()
		if errors.Is(err, usecase.ErrCVLimitReached) {
			response.Error(w, http.StatusBadRequest, "maksimal 3 CV")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusBadRequest, "category not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]dto.WorkerCVItemDTO, 0, len(created))
	for i := range created {
		cat := catMap[created[i].CategoryID]
		out = append(out, dto.WorkerCVItemDTO{
			ID:        created[i].ID.String(),
			FileURL:   created[i].FileURL,
			CreatedAt: created[i].CreatedAt.Format(time.RFC3339),
			Category:  dto.CategoryDTO{ID: created[i].CategoryID, Name: cat.Name},
		})
	}

	if len(out) == 1 {
		response.JSON(w, http.StatusCreated, map[string]any{"message": "CV berhasil diupload", "data": out[0]})
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"message": "CV berhasil diupload", "data": out})
}

func (h *JobHandler) DeleteWorkerCVs(w http.ResponseWriter, r *http.Request) {
	userID, role, err := extractAuth(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "worker" {
		response.Error(w, http.StatusForbidden, "only workers can access this resource")
		return
	}

	var req struct {
		IDs   []string `json:"ids"`
		CVIDs []string `json:"cv_ids"`
		ID1   string   `json:"id_1"`
		ID2   string   `json:"id_2"`
		ID3   string   `json:"id_3"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	ids := make([]string, 0)
	if len(req.IDs) > 0 {
		ids = append(ids, req.IDs...)
	} else if len(req.CVIDs) > 0 {
		ids = append(ids, req.CVIDs...)
	} else {
		if strings.TrimSpace(req.ID1) != "" {
			ids = append(ids, req.ID1)
		}
		if strings.TrimSpace(req.ID2) != "" {
			ids = append(ids, req.ID2)
		}
		if strings.TrimSpace(req.ID3) != "" {
			ids = append(ids, req.ID3)
		}
	}
	if len(ids) == 0 {
		response.Error(w, http.StatusBadRequest, "missing ids")
		return
	}
	if len(ids) > 3 {
		response.Error(w, http.StatusBadRequest, "maksimal 3 CV")
		return
	}

	cvs, err := h.uc.DeleteWorkerCVsBulk(userID, ids)
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

	for i := range cvs {
		if strings.HasPrefix(cvs[i].FileURL, "/uploads/") {
			_ = os.Remove(strings.TrimPrefix(cvs[i].FileURL, "/"))
		}
	}

	response.JSON(w, http.StatusOK, map[string]any{"message": "CV berhasil dihapus"})
}

func (h *JobHandler) PatchWorkerCVs(w http.ResponseWriter, r *http.Request) {
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

	type slot struct {
		Idx        int
		CVKey      string
		CVID       uuid.UUID
		CatID      *int
		FileKey    string
		LocalPath  string
		NewFileURL *string
		OldFileURL string
	}

	slots := make([]slot, 0, 3)
	for i := 1; i <= 3; i++ {
		cvKey := fmt.Sprintf("cv_id_%d", i)
		catKey := fmt.Sprintf("category_id_%d", i)
		fileKey := fmt.Sprintf("file_%d", i)
		cvVal := strings.TrimSpace(r.FormValue(cvKey))
		catVal := strings.TrimSpace(r.FormValue(catKey))
		_, filePresent := r.MultipartForm.File[fileKey]
		if cvVal == "" && catVal == "" && !filePresent {
			continue
		}
		if cvVal == "" {
			response.Error(w, http.StatusBadRequest, fmt.Sprintf("missing %s", cvKey))
			return
		}
		cvID, err := uuid.Parse(cvVal)
		if err != nil {
			response.Error(w, http.StatusBadRequest, fmt.Sprintf("invalid %s", cvKey))
			return
		}
		var catIDPtr *int
		if catVal != "" {
			cid, err := strconv.Atoi(catVal)
			if err != nil || cid <= 0 {
				response.Error(w, http.StatusBadRequest, fmt.Sprintf("invalid %s", catKey))
				return
			}
			catIDPtr = &cid
		}
		if catIDPtr == nil && !filePresent {
			response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s harus mengirim category_id atau file", cvKey))
			return
		}
		slots = append(slots, slot{Idx: i, CVKey: cvKey, CVID: cvID, CatID: catIDPtr, FileKey: fileKey})
	}
	if len(slots) == 0 {
		response.Error(w, http.StatusBadRequest, "missing updates")
		return
	}

	cleanup := func() {
		for i := range slots {
			if slots[i].LocalPath != "" {
				_ = os.Remove(slots[i].LocalPath)
			}
		}
	}

	existingCVs, _, err := h.uc.ListWorkerCVs(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	cvByID := make(map[uuid.UUID]domain.WorkerCV, len(existingCVs))
	for i := range existingCVs {
		cvByID[existingCVs[i].ID] = existingCVs[i]
	}
	for i := range slots {
		cv, ok := cvByID[slots[i].CVID]
		if !ok {
			response.Error(w, http.StatusNotFound, fmt.Sprintf("%s not found", slots[i].CVKey))
			return
		}
		slots[i].OldFileURL = cv.FileURL
	}

	existingHashes := make(map[string]struct{})
	for _, cv := range existingCVs {
		if !strings.HasPrefix(cv.FileURL, "/uploads/") {
			continue
		}
		hash, err := fileSHA256Hex(strings.TrimPrefix(cv.FileURL, "/"))
		if err != nil {
			continue
		}
		existingHashes[hash] = struct{}{}
	}

	if err := os.MkdirAll(filepath.Join("uploads", "cvs"), 0o755); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	newHashes := make(map[string]int)
	for i := range slots {
		_, filePresent := r.MultipartForm.File[slots[i].FileKey]
		if !filePresent {
			continue
		}
		f, _, err := r.FormFile(slots[i].FileKey)
		if err != nil {
			cleanup()
			response.Error(w, http.StatusBadRequest, fmt.Sprintf("missing %s", slots[i].FileKey))
			return
		}
		buf := make([]byte, 4)
		n, _ := io.ReadFull(f, buf)
		if n < 4 || string(buf) != "%PDF" {
			_ = f.Close()
			cleanup()
			response.Error(w, http.StatusBadRequest, "file must be pdf")
			return
		}
		newFileID := uuid.New()
		name := fmt.Sprintf("%s.pdf", newFileID.String())
		localPath := filepath.Join("uploads", "cvs", name)
		outFile, err := os.Create(localPath)
		if err != nil {
			_ = f.Close()
			cleanup()
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}
		hasher := sha256.New()
		mw := io.MultiWriter(outFile, hasher)
		if _, err := mw.Write(buf); err != nil {
			_ = outFile.Close()
			_ = f.Close()
			cleanup()
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}
		if _, err := io.Copy(mw, f); err != nil {
			_ = outFile.Close()
			_ = f.Close()
			cleanup()
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}
		_ = outFile.Close()
		_ = f.Close()

		newHash := hex.EncodeToString(hasher.Sum(nil))
		if _, ok := existingHashes[newHash]; ok {
			slots[i].LocalPath = localPath
			cleanup()
			response.Error(w, http.StatusBadRequest, "cv yang sama sudah pernah diupload")
			return
		}
		if prevIdx, ok := newHashes[newHash]; ok {
			slots[i].LocalPath = localPath
			cleanup()
			response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s duplikat dengan file_%d", slots[i].FileKey, prevIdx))
			return
		}
		newHashes[newHash] = slots[i].Idx

		url := "/uploads/cvs/" + name
		slots[i].LocalPath = localPath
		slots[i].NewFileURL = &url
	}

	items := make([]usecase.WorkerCVUpdateItem, 0, len(slots))
	for i := range slots {
		items = append(items, usecase.WorkerCVUpdateItem{ID: slots[i].CVID, CategoryID: slots[i].CatID, FileURL: slots[i].NewFileURL})
	}

	updated, catMap, err := h.uc.UpdateWorkerCVsBulk(userID, items)
	if err != nil {
		cleanup()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "cv not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	for i := range slots {
		if slots[i].NewFileURL != nil && strings.HasPrefix(slots[i].OldFileURL, "/uploads/") {
			_ = os.Remove(strings.TrimPrefix(slots[i].OldFileURL, "/"))
		}
	}

	out := make([]dto.WorkerCVItemDTO, 0, len(updated))
	for i := range updated {
		cat := catMap[updated[i].CategoryID]
		out = append(out, dto.WorkerCVItemDTO{ID: updated[i].ID.String(), FileURL: updated[i].FileURL, CreatedAt: updated[i].CreatedAt.Format(time.RFC3339), Category: dto.CategoryDTO{ID: updated[i].CategoryID, Name: cat.Name}})
	}
	response.JSON(w, http.StatusOK, map[string]any{"message": "CV berhasil diupdate", "data": out})
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

func (h *JobHandler) PatchEmployerProfile(w http.ResponseWriter, r *http.Request) {
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
