package repository

import (
	"fmt"
	"strings"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *JobRepository) GetWorkerProfileByUserID(userID uuid.UUID) (*domain.WorkerProfile, error) {
	var w domain.WorkerProfile
	q := r.db.Preload("Availabilities").Preload("CVs")
	if err := q.Where("user_id = ?", userID).First(&w).Error; err == nil {
		return &w, nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("get worker profile by user id: %w", err)
	}

	if err := q.First(&w, "id = ?", userID).Error; err == nil {
		return &w, nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("get worker profile by id fallback: %w", err)
	}

	return nil, gorm.ErrRecordNotFound
}

func (r *JobRepository) CreateWorkerProfile(userID uuid.UUID) (*domain.WorkerProfile, error) {
	wp := &domain.WorkerProfile{ID: userID, UserID: userID, FullName: ""}
	if err := r.db.Create(wp).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			if w, err2 := r.GetWorkerProfileByUserID(userID); err2 == nil {
				return w, nil
			}
		}
		return nil, fmt.Errorf("create worker profile: %w", err)
	}
	return wp, nil
}

func (r *JobRepository) UpdateWorkerProfile(wp *domain.WorkerProfile) error {
	if err := r.db.Save(wp).Error; err != nil {
		return fmt.Errorf("update worker profile: %w", err)
	}
	return nil
}

func (r *JobRepository) ListWorkerAvailabilities(workerID uuid.UUID) ([]domain.Availability, error) {
	var avs []domain.Availability
	if err := r.db.Where("worker_id = ?", workerID).Order("day asc").Find(&avs).Error; err != nil {
		return nil, fmt.Errorf("list availabilities: %w", err)
	}
	return avs, nil
}

func (r *JobRepository) ReplaceWorkerAvailabilities(workerID uuid.UUID, avs []domain.Availability) ([]domain.Availability, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.Where("worker_id = ?", workerID).Delete(&domain.Availability{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("delete availabilities: %w", err)
	}
	for i := range avs {
		avs[i].WorkerID = workerID
		avs[i].ID = 0
	}
	if len(avs) > 0 {
		if err := tx.Create(&avs).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("create availabilities: %w", err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return avs, nil
}

func (r *JobRepository) ListWorkerCVs(workerID uuid.UUID) ([]domain.WorkerCV, error) {
	var cvs []domain.WorkerCV
	if err := r.db.Where("worker_id = ?", workerID).Order("created_at desc").Find(&cvs).Error; err != nil {
		return nil, fmt.Errorf("list cvs: %w", err)
	}
	return cvs, nil
}

func (r *JobRepository) CountWorkerCVs(workerID uuid.UUID) (int64, error) {
	var cnt int64
	if err := r.db.Model(&domain.WorkerCV{}).Where("worker_id = ?", workerID).Count(&cnt).Error; err != nil {
		return 0, fmt.Errorf("count cvs: %w", err)
	}
	return cnt, nil
}

func (r *JobRepository) GetWorkerCVByID(cvID uuid.UUID) (*domain.WorkerCV, error) {
	var cv domain.WorkerCV
	if err := r.db.First(&cv, "id = ?", cvID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get cv by id: %w", err)
	}
	return &cv, nil
}

func (r *JobRepository) GetWorkerCVsByIDs(cvIDs []uuid.UUID, workerID uuid.UUID) ([]domain.WorkerCV, error) {
	var cvs []domain.WorkerCV
	if len(cvIDs) == 0 {
		return []domain.WorkerCV{}, nil
	}
	if err := r.db.Where("worker_id = ? AND id IN ?", workerID, cvIDs).Find(&cvs).Error; err != nil {
		return nil, fmt.Errorf("get cvs: %w", err)
	}
	return cvs, nil
}

func (r *JobRepository) UpdateWorkerCVs(cvs []domain.WorkerCV) error {
	if len(cvs) == 0 {
		return nil
	}
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for i := range cvs {
		if err := tx.Model(&domain.WorkerCV{}).
			Where("id = ? AND worker_id = ?", cvs[i].ID, cvs[i].WorkerID).
			Updates(map[string]any{"category_id": cvs[i].CategoryID, "file_url": cvs[i].FileURL}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("update cv: %w", err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (r *JobRepository) DeleteWorkerCVs(cvIDs []uuid.UUID, workerID uuid.UUID) ([]domain.WorkerCV, error) {
	cvs, err := r.GetWorkerCVsByIDs(cvIDs, workerID)
	if err != nil {
		return nil, err
	}
	if len(cvs) != len(cvIDs) {
		return nil, gorm.ErrRecordNotFound
	}
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.Where("worker_id = ? AND id IN ?", workerID, cvIDs).Delete(&domain.WorkerCV{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("delete cvs: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return cvs, nil
}

func (r *JobRepository) CreateWorkerCV(cv *domain.WorkerCV) error {
	if err := r.db.Create(cv).Error; err != nil {
		return fmt.Errorf("create cv: %w", err)
	}
	return nil
}

func (r *JobRepository) CreateWorkerCVs(cvs []domain.WorkerCV) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Create(&cvs).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("create cvs: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (r *JobRepository) DeleteWorkerCV(cvID, workerID uuid.UUID) (*domain.WorkerCV, error) {
	cv, err := r.GetWorkerCVByID(cvID)
	if err != nil {
		return nil, err
	}
	if cv.WorkerID != workerID {
		return nil, gorm.ErrRecordNotFound
	}
	if err := r.db.Delete(&domain.WorkerCV{}, "id = ?", cvID).Error; err != nil {
		return nil, fmt.Errorf("delete cv: %w", err)
	}
	return cv, nil
}

func (r *JobRepository) GetEmployerProfileByUserID(userID uuid.UUID) (*domain.EmployerProfile, error) {
	return r.GetEmployerByUserID(userID)
}

func (r *JobRepository) UpdateEmployerProfile(emp *domain.EmployerProfile) error {
	if err := r.db.Save(emp).Error; err != nil {
		return fmt.Errorf("update employer profile: %w", err)
	}
	return nil
}

func (r *JobRepository) ListCategories() ([]domain.Category, error) {
	var cats []domain.Category
	if err := r.db.Order("id asc").Find(&cats).Error; err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return cats, nil
}
