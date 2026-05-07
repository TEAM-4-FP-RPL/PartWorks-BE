package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrInvalidAvailability = errors.New("invalid availability")
var ErrCVLimitReached = errors.New("cv limit reached")

const maxWorkerCVs = 3

type WorkerProfileUpdate struct {
	FullName    *string
	PhoneNumber *string
	Bio         *string
	Skills      *string
	PhotoURL    *string
}

type EmployerProfileUpdate struct {
	CompanyName *string
	Description *string
	LogoURL     *string
}

type AvailabilityItem struct {
	Day       int
	StartTime string
	EndTime   string
}

type WorkerCVCreateItem struct {
	ID         uuid.UUID
	CategoryID int
	FileURL    string
}

type WorkerCVUpdateItem struct {
	ID         uuid.UUID
	CategoryID *int
	FileURL    *string
}

func (uc *JobUsecase) GetWorkerProfile(userIDStr string) (*domain.WorkerProfile, map[int]domain.Category, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleWorker {
		return nil, nil, fmt.Errorf("user is not worker")
	}

	wp, err := uc.repo.GetWorkerProfileByUserID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if _, err2 := uc.repo.CreateWorkerProfile(uid); err2 != nil {
				return nil, nil, fmt.Errorf("create worker profile: %w", err2)
			}
			wp, err = uc.repo.GetWorkerProfileByUserID(uid)
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("get worker profile: %w", err)
	}

	idsUniq := make(map[int]struct{})
	ids := make([]int, 0)
	for i := range wp.CVs {
		cid := wp.CVs[i].CategoryID
		if _, ok := idsUniq[cid]; !ok {
			idsUniq[cid] = struct{}{}
			ids = append(ids, cid)
		}
	}
	cats, err := uc.repo.GetCategoriesByIDs(ids)
	if err != nil {
		return nil, nil, fmt.Errorf("get categories: %w", err)
	}
	return wp, cats, nil
}

func (uc *JobUsecase) UpdateWorkerProfile(userIDStr string, upd WorkerProfileUpdate) (*domain.WorkerProfile, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleWorker {
		return nil, fmt.Errorf("user is not worker")
	}

	wp, err := uc.repo.GetWorkerProfileByUserID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wp, err = uc.repo.CreateWorkerProfile(uid)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("get worker profile: %w", err)
	}

	if upd.FullName != nil {
		wp.FullName = strings.TrimSpace(*upd.FullName)
	}
	if upd.PhoneNumber != nil {
		wp.PhoneNumber = strings.TrimSpace(*upd.PhoneNumber)
	}
	if upd.Bio != nil {
		wp.Bio = strings.TrimSpace(*upd.Bio)
	}
	if upd.Skills != nil {
		wp.Skills = strings.TrimSpace(*upd.Skills)
	}
	if upd.PhotoURL != nil {
		wp.PhotoURL = strings.TrimSpace(*upd.PhotoURL)
	}
	wp.UpdatedAt = time.Now()

	if err := uc.repo.UpdateWorkerProfile(wp); err != nil {
		return nil, fmt.Errorf("update worker profile: %w", err)
	}
	return wp, nil
}

func (uc *JobUsecase) ListWorkerAvailabilities(userIDStr string) ([]domain.Availability, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleWorker {
		return nil, fmt.Errorf("user is not worker")
	}
	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		return nil, fmt.Errorf("get worker: %w", err)
	}
	avs, err := uc.repo.ListWorkerAvailabilities(wp.ID)
	if err != nil {
		return nil, fmt.Errorf("list availabilities: %w", err)
	}
	return avs, nil
}

func (uc *JobUsecase) UpdateWorkerAvailabilities(userIDStr string, items []AvailabilityItem) ([]domain.Availability, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleWorker {
		return nil, fmt.Errorf("user is not worker")
	}
	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		return nil, fmt.Errorf("get worker: %w", err)
	}

	avs := make([]domain.Availability, 0, len(items))
	for _, it := range items {
		if it.Day < 1 || it.Day > 7 {
			return nil, ErrInvalidAvailability
		}
		avs = append(avs, domain.Availability{Day: it.Day, StartTime: it.StartTime, EndTime: it.EndTime})
	}
	out, err := uc.repo.ReplaceWorkerAvailabilities(wp.ID, avs)
	if err != nil {
		return nil, fmt.Errorf("replace availabilities: %w", err)
	}
	return out, nil
}

func (uc *JobUsecase) ListWorkerCVs(userIDStr string) ([]domain.WorkerCV, map[int]domain.Category, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleWorker {
		return nil, nil, fmt.Errorf("user is not worker")
	}
	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		return nil, nil, fmt.Errorf("get worker: %w", err)
	}

	cvs, err := uc.repo.ListWorkerCVs(wp.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("list cvs: %w", err)
	}

	uniq := make(map[int]struct{})
	ids := make([]int, 0)
	for i := range cvs {
		cid := cvs[i].CategoryID
		if _, ok := uniq[cid]; !ok {
			uniq[cid] = struct{}{}
			ids = append(ids, cid)
		}
	}
	cats, err := uc.repo.GetCategoriesByIDs(ids)
	if err != nil {
		return nil, nil, fmt.Errorf("get categories: %w", err)
	}
	return cvs, cats, nil
}

func (uc *JobUsecase) CreateWorkerCV(userIDStr string, cvID uuid.UUID, categoryID int, fileURL string) (*domain.WorkerCV, *domain.Category, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleWorker {
		return nil, nil, fmt.Errorf("user is not worker")
	}
	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		return nil, nil, fmt.Errorf("get worker: %w", err)
	}

	cnt, err := uc.repo.CountWorkerCVs(wp.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("count cvs: %w", err)
	}
	if cnt >= maxWorkerCVs {
		return nil, nil, ErrCVLimitReached
	}

	cat, err := uc.repo.GetCategoryByID(categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, gorm.ErrRecordNotFound
		}
		return nil, nil, fmt.Errorf("get category: %w", err)
	}

	cv := &domain.WorkerCV{ID: cvID, WorkerID: wp.ID, CategoryID: categoryID, FileURL: fileURL, CreatedAt: time.Now()}
	if err := uc.repo.CreateWorkerCV(cv); err != nil {
		return nil, nil, fmt.Errorf("create cv: %w", err)
	}
	return cv, cat, nil
}

func (uc *JobUsecase) CreateWorkerCVsBulk(userIDStr string, items []WorkerCVCreateItem) ([]domain.WorkerCV, map[int]domain.Category, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleWorker {
		return nil, nil, fmt.Errorf("user is not worker")
	}
	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		return nil, nil, fmt.Errorf("get worker: %w", err)
	}

	cnt, err := uc.repo.CountWorkerCVs(wp.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("count cvs: %w", err)
	}
	if cnt+int64(len(items)) > maxWorkerCVs {
		return nil, nil, ErrCVLimitReached
	}

	uniq := make(map[int]struct{})
	ids := make([]int, 0)
	for i := range items {
		cid := items[i].CategoryID
		if _, ok := uniq[cid]; !ok {
			uniq[cid] = struct{}{}
			ids = append(ids, cid)
		}
	}
	cats, err := uc.repo.GetCategoriesByIDs(ids)
	if err != nil {
		return nil, nil, fmt.Errorf("get categories: %w", err)
	}
	for i := range items {
		if _, ok := cats[items[i].CategoryID]; !ok {
			return nil, nil, gorm.ErrRecordNotFound
		}
	}

	cvs := make([]domain.WorkerCV, 0, len(items))
	now := time.Now()
	for i := range items {
		cvs = append(cvs, domain.WorkerCV{ID: items[i].ID, WorkerID: wp.ID, CategoryID: items[i].CategoryID, FileURL: items[i].FileURL, CreatedAt: now})
	}
	if err := uc.repo.CreateWorkerCVs(cvs); err != nil {
		return nil, nil, fmt.Errorf("create cvs: %w", err)
	}
	return cvs, cats, nil
}

func (uc *JobUsecase) DeleteWorkerCV(userIDStr, cvIDStr string) (*domain.WorkerCV, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleWorker {
		return nil, fmt.Errorf("user is not worker")
	}
	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		return nil, fmt.Errorf("get worker: %w", err)
	}
	cvID, err := uuid.Parse(cvIDStr)
	if err != nil {
		return nil, ErrInvalidID
	}
	cv, err := uc.repo.DeleteWorkerCV(cvID, wp.ID)
	if err != nil {
		return nil, err
	}
	return cv, nil
}

func (uc *JobUsecase) DeleteWorkerCVsBulk(userIDStr string, cvIDStrs []string) ([]domain.WorkerCV, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleWorker {
		return nil, fmt.Errorf("user is not worker")
	}
	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		return nil, fmt.Errorf("get worker: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(cvIDStrs))
	for i := range cvIDStrs {
		id, err := uuid.Parse(strings.TrimSpace(cvIDStrs[i]))
		if err != nil {
			return nil, ErrInvalidID
		}
		ids = append(ids, id)
	}
	cvs, err := uc.repo.DeleteWorkerCVs(ids, wp.ID)
	if err != nil {
		return nil, err
	}
	return cvs, nil
}

func (uc *JobUsecase) UpdateWorkerCVsBulk(userIDStr string, items []WorkerCVUpdateItem) ([]domain.WorkerCV, map[int]domain.Category, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleWorker {
		return nil, nil, fmt.Errorf("user is not worker")
	}
	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		return nil, nil, fmt.Errorf("get worker: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}
	current, err := uc.repo.GetWorkerCVsByIDs(ids, wp.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("get cvs: %w", err)
	}
	if len(current) != len(items) {
		return nil, nil, gorm.ErrRecordNotFound
	}
	byID := make(map[uuid.UUID]domain.WorkerCV, len(current))
	for i := range current {
		byID[current[i].ID] = current[i]
	}

	needCats := make(map[int]struct{})
	for i := range items {
		cv := byID[items[i].ID]
		if items[i].CategoryID != nil {
			cv.CategoryID = *items[i].CategoryID
		}
		if items[i].FileURL != nil {
			cv.FileURL = *items[i].FileURL
		}
		byID[items[i].ID] = cv
		needCats[cv.CategoryID] = struct{}{}
	}

	catIDs := make([]int, 0, len(needCats))
	for id := range needCats {
		catIDs = append(catIDs, id)
	}
	cats, err := uc.repo.GetCategoriesByIDs(catIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("get categories: %w", err)
	}
	for id := range needCats {
		if _, ok := cats[id]; !ok {
			return nil, nil, gorm.ErrRecordNotFound
		}
	}

	out := make([]domain.WorkerCV, 0, len(items))
	for i := range items {
		out = append(out, byID[items[i].ID])
	}
	if err := uc.repo.UpdateWorkerCVs(out); err != nil {
		return nil, nil, fmt.Errorf("update cvs: %w", err)
	}
	return out, cats, nil
}

func (uc *JobUsecase) GetEmployerProfile(userIDStr string) (*domain.EmployerProfile, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleEmployer {
		return nil, fmt.Errorf("user is not employer")
	}
	p, err := uc.repo.GetEmployerProfileByUserID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			p, err = uc.repo.CreateEmployerProfile(uid)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("get employer profile: %w", err)
	}
	return p, nil
}

func (uc *JobUsecase) UpdateEmployerProfile(userIDStr string, upd EmployerProfileUpdate) (*domain.EmployerProfile, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Role != domain.RoleEmployer {
		return nil, fmt.Errorf("user is not employer")
	}
	p, err := uc.repo.GetEmployerProfileByUserID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			p, err = uc.repo.CreateEmployerProfile(uid)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("get employer profile: %w", err)
	}
	if upd.CompanyName != nil {
		p.CompanyName = strings.TrimSpace(*upd.CompanyName)
	}
	if upd.Description != nil {
		p.Description = strings.TrimSpace(*upd.Description)
	}
	if upd.LogoURL != nil {
		p.LogoURL = strings.TrimSpace(*upd.LogoURL)
	}
	p.UpdatedAt = time.Now()
	if err := uc.repo.UpdateEmployerProfile(p); err != nil {
		return nil, fmt.Errorf("update employer profile: %w", err)
	}
	return p, nil
}

func (uc *JobUsecase) ListCategories() ([]domain.Category, error) {
	cats, err := uc.repo.ListCategories()
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return cats, nil
}
