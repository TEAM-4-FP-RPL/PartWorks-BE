package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path"
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

var (
	errFileTooLarge = errors.New("file too large")
	errNotPDF       = errors.New("file must be pdf")
	errNotImage     = errors.New("file must be png or jpg")
)

const maxUploadSize = 20 << 20

func isHex64(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			continue
		case r >= 'a' && r <= 'f':
			continue
		case r >= 'A' && r <= 'F':
			continue
		default:
			return false
		}
	}
	return true
}

func extractSHA256FromFileURL(fileURL string) string {
	u := strings.TrimSpace(fileURL)
	if u == "" {
		return ""
	}
	if q := strings.IndexByte(u, '?'); q >= 0 {
		u = u[:q]
	}
	base := path.Base(u)
	base = strings.TrimSuffix(base, ".pdf")
	if isHex64(base) {
		return strings.ToLower(base)
	}
	return ""
}

func readPDFAndHash(f multipart.File) ([]byte, string, error) {
	defer f.Close()
	lr := io.LimitReader(f, maxUploadSize+1)
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > maxUploadSize {
		return nil, "", errFileTooLarge
	}
	if len(data) < 4 || string(data[:4]) != "%PDF" {
		return nil, "", errNotPDF
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), nil
}

func readImageAndHash(f multipart.File) (data []byte, hash string, contentType string, ext string, err error) {
	defer f.Close()
	lr := io.LimitReader(f, maxUploadSize+1)
	data, err = io.ReadAll(lr)
	if err != nil {
		return nil, "", "", "", err
	}
	if int64(len(data)) > maxUploadSize {
		return nil, "", "", "", errFileTooLarge
	}
	if len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		contentType = "image/png"
		ext = "png"
	} else if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		contentType = "image/jpeg"
		ext = "jpg"
	} else {
		return nil, "", "", "", errNotImage
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), contentType, ext, nil
}

func multipartValue(r *http.Request, key string) (string, bool) {
	if r.MultipartForm == nil {
		return "", false
	}
	vals, ok := r.MultipartForm.Value[key]
	if !ok {
		return "", false
	}
	if len(vals) == 0 {
		return "", true
	}
	return vals[0], true
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

	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if h.storage == nil {
			log.Printf("PatchWorkerProfile: storage is nil")
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			log.Printf("PatchWorkerProfile: parse multipart: %v", err)
			response.Error(w, http.StatusBadRequest, "invalid multipart")
			return
		}

		old, _, err := h.uc.GetWorkerProfile(userID)
		if err != nil {
			log.Printf("PatchWorkerProfile: get old profile: %v", err)
			if strings.Contains(err.Error(), "user is not worker") {
				response.Error(w, http.StatusForbidden, "only workers can access this resource")
				return
			}
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}

		upd := usecase.WorkerProfileUpdate{}
		if v, ok := multipartValue(r, "full_name"); ok {
			upd.FullName = &v
		}
		if v, ok := multipartValue(r, "phone_number"); ok {
			upd.PhoneNumber = &v
		}
		if v, ok := multipartValue(r, "bio"); ok {
			upd.Bio = &v
		}
		if v, ok := multipartValue(r, "skills"); ok {
			upd.Skills = &v
		}

		uploadedKey := ""
		if r.MultipartForm != nil {
			files := r.MultipartForm.File["photo"]
			if len(files) > 0 {
				f, err := files[0].Open()
				if err != nil {
					response.Error(w, http.StatusBadRequest, "invalid photo")
					return
				}
				data, hash, ct, ext, err := readImageAndHash(f)
				if err != nil {
					if errors.Is(err, errNotImage) {
						response.Error(w, http.StatusBadRequest, "file must be png or jpg")
						return
					}
					if errors.Is(err, errFileTooLarge) {
						response.Error(w, http.StatusBadRequest, "file too large")
						return
					}
					response.Error(w, http.StatusInternalServerError, "internal error")
					return
				}

				safeUserID := strings.ReplaceAll(userID, "/", "_")
				key := fmt.Sprintf("photos/workers/%s/%s.%s", safeUserID, hash, ext)

				if oldKey, ok := h.storage.KeyFromURL(old.PhotoURL); ok && oldKey == key {
					url := h.storage.PublicURL(key)
					upd.PhotoURL = &url
				} else {
					url, err := h.storage.Put(r.Context(), key, bytes.NewReader(data), ct, int64(len(data)))
					if err != nil {
						log.Printf("PatchWorkerProfile: storage put key=%s err=%v", key, err)
						response.Error(w, http.StatusInternalServerError, "internal error")
						return
					}
					uploadedKey = key
					upd.PhotoURL = &url
				}
			}
		}

		wp, err := h.uc.UpdateWorkerProfile(userID, upd)
		if err != nil {
			log.Printf("PatchWorkerProfile: update profile: %v", err)
			if uploadedKey != "" {
				if err2 := h.storage.Delete(r.Context(), uploadedKey); err2 != nil {
					log.Printf("PatchWorkerProfile: rollback delete key=%s err=%v", uploadedKey, err2)
				}
			}
			if strings.Contains(err.Error(), "user is not worker") {
				response.Error(w, http.StatusForbidden, "only workers can access this resource")
				return
			}
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}

		if uploadedKey != "" {
			if oldKey, ok := h.storage.KeyFromURL(old.PhotoURL); ok && oldKey != "" && oldKey != uploadedKey {
				if err := h.storage.Delete(r.Context(), oldKey); err != nil {
					log.Printf("PatchWorkerProfile: delete old key=%s err=%v", oldKey, err)
				}
			}
		}

		response.JSON(w, http.StatusOK, map[string]any{
			"message": "Profil berhasil diupdate",
			"data": map[string]any{
				"id":         wp.ID.String(),
				"full_name":  wp.FullName,
				"updated_at": wp.UpdatedAt.Format(time.RFC3339),
			},
		})
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
		if errors.Is(err, usecase.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "only workers can access this resource")
			return
		}
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

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid multipart")
		return
	}

	type slot struct {
		Idx      int
		CatID    int
		FileKey  string
		CatKey   string
		FileURL  string
		FileID   uuid.UUID
		FileHash string
		ObjKey   string
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
		if h := extractSHA256FromFileURL(cv.FileURL); h != "" {
			existingHashes[h] = struct{}{}
			continue
		}
	}

	newHashes := make(map[string]int)
	uploaded := make([]string, 0, len(slots))
	cleanup := func() {
		for _, k := range uploaded {
			_ = h.storage.Delete(r.Context(), k)
		}
	}

	safeUserID := strings.ReplaceAll(userID, "/", "_")
	createItems := make([]usecase.WorkerCVCreateItem, 0, len(slots))
	for i := range slots {
		f, _, err := r.FormFile(slots[i].FileKey)
		if err != nil {
			cleanup()
			response.Error(w, http.StatusBadRequest, fmt.Sprintf("missing %s", slots[i].FileKey))
			return
		}

		data, newHash, err := readPDFAndHash(f)
		if err != nil {
			cleanup()
			if errors.Is(err, errNotPDF) {
				response.Error(w, http.StatusBadRequest, "file must be pdf")
				return
			}
			if errors.Is(err, errFileTooLarge) {
				response.Error(w, http.StatusBadRequest, "file too large")
				return
			}
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}

		if _, ok := existingHashes[newHash]; ok {
			cleanup()
			response.Error(w, http.StatusBadRequest, "cv yang sama sudah pernah diupload")
			return
		}
		if prevIdx, ok := newHashes[newHash]; ok {
			cleanup()
			response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s duplikat dengan file_%d", slots[i].FileKey, prevIdx))
			return
		}
		newHashes[newHash] = slots[i].Idx

		objKey := fmt.Sprintf("cvs/%s/%s.pdf", safeUserID, newHash)
		url, err := h.storage.Put(r.Context(), objKey, bytes.NewReader(data), "application/pdf", int64(len(data)))
		if err != nil {
			cleanup()
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}
		uploaded = append(uploaded, objKey)

		cvID := uuid.New()
		slots[i].FileID = cvID
		slots[i].FileURL = url
		slots[i].FileHash = newHash
		slots[i].ObjKey = objKey
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
		if key, ok := h.storage.KeyFromURL(cvs[i].FileURL); ok {
			_ = h.storage.Delete(r.Context(), key)
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

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid multipart")
		return
	}

	type slot struct {
		Idx        int
		CVKey      string
		CVID       uuid.UUID
		CatID      *int
		FileKey    string
		NewFileURL *string
		OldFileURL string
		NewObjKey  string
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
		if h := extractSHA256FromFileURL(cv.FileURL); h != "" {
			existingHashes[h] = struct{}{}
			continue
		}
	}

	newHashes := make(map[string]int)
	uploaded := make([]string, 0, len(slots))
	cleanup := func() {
		for _, k := range uploaded {
			_ = h.storage.Delete(r.Context(), k)
		}
	}

	safeUserID := strings.ReplaceAll(userID, "/", "_")
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

		data, newHash, err := readPDFAndHash(f)
		if err != nil {
			cleanup()
			if errors.Is(err, errNotPDF) {
				response.Error(w, http.StatusBadRequest, "file must be pdf")
				return
			}
			if errors.Is(err, errFileTooLarge) {
				response.Error(w, http.StatusBadRequest, "file too large")
				return
			}
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}

		oldHash := extractSHA256FromFileURL(slots[i].OldFileURL)
		if oldHash != "" && oldHash == newHash {
			continue
		}

		if _, ok := existingHashes[newHash]; ok {
			cleanup()
			response.Error(w, http.StatusBadRequest, "cv yang sama sudah pernah diupload")
			return
		}
		if prevIdx, ok := newHashes[newHash]; ok {
			cleanup()
			response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s duplikat dengan file_%d", slots[i].FileKey, prevIdx))
			return
		}
		newHashes[newHash] = slots[i].Idx

		objKey := fmt.Sprintf("cvs/%s/%s.pdf", safeUserID, newHash)
		url, err := h.storage.Put(r.Context(), objKey, bytes.NewReader(data), "application/pdf", int64(len(data)))
		if err != nil {
			cleanup()
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}
		uploaded = append(uploaded, objKey)

		slots[i].NewObjKey = objKey
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
		if slots[i].NewFileURL == nil || *slots[i].NewFileURL == slots[i].OldFileURL {
			continue
		}
		if key, ok := h.storage.KeyFromURL(slots[i].OldFileURL); ok {
			_ = h.storage.Delete(r.Context(), key)
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

	if key, ok := h.storage.KeyFromURL(cv.FileURL); ok {
		_ = h.storage.Delete(r.Context(), key)
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

	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if h.storage == nil {
			log.Printf("PatchEmployerProfile: storage is nil")
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			log.Printf("PatchEmployerProfile: parse multipart: %v", err)
			response.Error(w, http.StatusBadRequest, "invalid multipart")
			return
		}

		old, err := h.uc.GetEmployerProfile(userID)
		if err != nil {
			log.Printf("PatchEmployerProfile: get old profile: %v", err)
			if strings.Contains(err.Error(), "user is not employer") {
				response.Error(w, http.StatusForbidden, "only employers can access this resource")
				return
			}
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}

		upd := usecase.EmployerProfileUpdate{}
		if v, ok := multipartValue(r, "company_name"); ok {
			upd.CompanyName = &v
		}
		if v, ok := multipartValue(r, "description"); ok {
			upd.Description = &v
		}

		uploadedKey := ""
		if r.MultipartForm != nil {
			files := r.MultipartForm.File["logo"]
			if len(files) > 0 {
				f, err := files[0].Open()
				if err != nil {
					response.Error(w, http.StatusBadRequest, "invalid logo")
					return
				}
				data, hash, ct, ext, err := readImageAndHash(f)
				if err != nil {
					if errors.Is(err, errNotImage) {
						response.Error(w, http.StatusBadRequest, "file must be png or jpg")
						return
					}
					if errors.Is(err, errFileTooLarge) {
						response.Error(w, http.StatusBadRequest, "file too large")
						return
					}
					response.Error(w, http.StatusInternalServerError, "internal error")
					return
				}

				safeUserID := strings.ReplaceAll(userID, "/", "_")
				key := fmt.Sprintf("logos/employers/%s/%s.%s", safeUserID, hash, ext)

				if oldKey, ok := h.storage.KeyFromURL(old.LogoURL); ok && oldKey == key {
					url := h.storage.PublicURL(key)
					upd.LogoURL = &url
				} else {
					url, err := h.storage.Put(r.Context(), key, bytes.NewReader(data), ct, int64(len(data)))
					if err != nil {
						log.Printf("PatchEmployerProfile: storage put key=%s err=%v", key, err)
						response.Error(w, http.StatusInternalServerError, "internal error")
						return
					}
					uploadedKey = key
					upd.LogoURL = &url
				}
			}
		}

		emp, err := h.uc.UpdateEmployerProfile(userID, upd)
		if err != nil {
			log.Printf("PatchEmployerProfile: update profile: %v", err)
			if uploadedKey != "" {
				if err2 := h.storage.Delete(r.Context(), uploadedKey); err2 != nil {
					log.Printf("PatchEmployerProfile: rollback delete key=%s err=%v", uploadedKey, err2)
				}
			}
			if strings.Contains(err.Error(), "user is not employer") {
				response.Error(w, http.StatusForbidden, "only employers can access this resource")
				return
			}
			response.Error(w, http.StatusInternalServerError, "internal error")
			return
		}

		if uploadedKey != "" {
			if oldKey, ok := h.storage.KeyFromURL(old.LogoURL); ok && oldKey != "" && oldKey != uploadedKey {
				if err := h.storage.Delete(r.Context(), oldKey); err != nil {
					log.Printf("PatchEmployerProfile: delete old key=%s err=%v", oldKey, err)
				}
			}
		}

		response.JSON(w, http.StatusOK, map[string]any{
			"message": "Profil berhasil diupdate",
			"data":    map[string]any{"id": emp.ID.String(), "company_name": emp.CompanyName, "updated_at": emp.UpdatedAt.Format(time.RFC3339)},
		})
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
