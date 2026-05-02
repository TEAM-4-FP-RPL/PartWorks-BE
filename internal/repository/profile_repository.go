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

func (r *JobRepository) CreateWorkerCV(cv *domain.WorkerCV) error {
	if err := r.db.Create(cv).Error; err != nil {
		return fmt.Errorf("create cv: %w", err)
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
