package repository

import (
	"fmt"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *JobRepository) GetWorkerByUserID(userID uuid.UUID) (*domain.WorkerProfile, error) {
	var w domain.WorkerProfile
	if err := r.db.Where("user_id = ?", userID).First(&w).Error; err == nil {
		return &w, nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("get worker by user id: %w", err)
	}

	if err := r.db.First(&w, "id = ?", userID).Error; err == nil {
		return &w, nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("get worker by id fallback: %w", err)
	}

	var u domain.User
	if err := r.db.First(&u, "id = ?", userID).Error; err == nil {
		return &domain.WorkerProfile{ID: userID, UserID: userID, FullName: ""}, nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("get user fallback: %w", err)
	}

	return nil, gorm.ErrRecordNotFound
}

func (r *JobRepository) GetCVByID(cvID uuid.UUID) (*domain.WorkerCV, error) {
	var cv domain.WorkerCV
	if err := r.db.First(&cv, "id = ?", cvID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get cv: %w", err)
	}
	return &cv, nil
}

func (r *JobRepository) HasApplication(jobID, workerID uuid.UUID) (bool, error) {
	var a domain.Application
	if err := r.db.Where("job_id = ? AND worker_id = ?", jobID, workerID).First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("check application: %w", err)
	}
	return true, nil
}

func (r *JobRepository) CreateApplication(app *domain.Application) error {
	if err := r.db.Create(app).Error; err != nil {
		return fmt.Errorf("create application: %w", err)
	}
	return nil
}

func (r *JobRepository) ListWorkerApplications(workerID uuid.UUID, status string, page, limit int) ([]domain.Application, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	base := r.db.Model(&domain.Application{}).Where("worker_id = ?", workerID)
	if status != "" {
		base = base.Where("status = ?", status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count worker applications: %w", err)
	}

	var apps []domain.Application
	q := r.db.Where("worker_id = ?", workerID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.
		Preload("Job").
		Preload("Job.Employer").
		Preload("CV").
		Order("applied_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&apps).Error; err != nil {
		return nil, 0, fmt.Errorf("list worker applications: %w", err)
	}

	return apps, total, nil
}

func (r *JobRepository) ListEmployerApplications(employerID uuid.UUID, status string, page, limit int) ([]domain.Application, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	base := r.db.Model(&domain.Application{}).
		Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("jobs.employer_id = ?", employerID)
	if status != "" {
		base = base.Where("applications.status = ?", status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count employer applications: %w", err)
	}

	var apps []domain.Application
	q := r.db.Model(&domain.Application{}).
		Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("jobs.employer_id = ?", employerID)
	if status != "" {
		q = q.Where("applications.status = ?", status)
	}
	if err := q.
		Preload("Job").
		Preload("Worker").
		Preload("CV").
		Order("applications.applied_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&apps).Error; err != nil {
		return nil, 0, fmt.Errorf("list employer applications: %w", err)
	}

	return apps, total, nil
}

func (r *JobRepository) ListJobApplications(jobID uuid.UUID, status string, page, limit int) ([]domain.Application, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	base := r.db.Model(&domain.Application{}).Where("job_id = ?", jobID)
	if status != "" {
		base = base.Where("status = ?", status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count job applications: %w", err)
	}

	var apps []domain.Application
	q := r.db.Model(&domain.Application{}).Where("job_id = ?", jobID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.
		Preload("Worker").
		Preload("CV").
		Order("applied_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&apps).Error; err != nil {
		return nil, 0, fmt.Errorf("list job applications: %w", err)
	}

	return apps, total, nil
}

func (r *JobRepository) GetCategoriesByIDs(ids []int) (map[int]domain.Category, error) {
	out := make(map[int]domain.Category)
	if len(ids) == 0 {
		return out, nil
	}
	var cats []domain.Category
	if err := r.db.Where("id IN ?", ids).Find(&cats).Error; err != nil {
		return nil, fmt.Errorf("get categories by ids: %w", err)
	}
	for _, c := range cats {
		out[c.ID] = c
	}
	return out, nil
}

func (r *JobRepository) GetApplicationByID(appID uuid.UUID) (*domain.Application, error) {
	var app domain.Application
	if err := r.db.First(&app, "id = ?", appID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get application by id: %w", err)
	}
	return &app, nil
}

func (r *JobRepository) UpdateApplicationStatus(appID uuid.UUID, status string) (*domain.Application, error) {
	var app domain.Application
	if err := r.db.First(&app, "id = ?", appID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get application for update: %w", err)
	}
	app.Status = status
	if err := r.db.Save(&app).Error; err != nil {
		return nil, fmt.Errorf("update application status: %w", err)
	}
	return &app, nil
}

func (r *JobRepository) DeleteWorkerApplication(appID, workerID uuid.UUID) error {
	res := r.db.Where("id = ? AND worker_id = ?", appID, workerID).Delete(&domain.Application{})
	if res.Error != nil {
		return fmt.Errorf("delete worker application: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
