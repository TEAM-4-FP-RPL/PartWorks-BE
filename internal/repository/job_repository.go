package repository

import (
	"context"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobRepository interface {
	GetByEmployerID(ctx context.Context, employerID uuid.UUID) ([]domain.Job, error)
}

type jobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) GetByEmployerID(ctx context.Context, employerID uuid.UUID) ([]domain.Job, error) {
	var jobs []domain.Job
	err := r.db.WithContext(ctx).Where("employer_id = ?", employerID).Preload("Schedules").Find(&jobs).Error
	if err != nil {
		return nil, err
	}
	return jobs, nil
}