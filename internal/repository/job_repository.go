package repository

import (
	"fmt"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"gorm.io/gorm"
)

type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository { return &JobRepository{db: db} }

type JobFilter struct {
	Search     string
	CategoryID *int
	Type       string
	Location   string
	Status     string
	Page       int
	Limit      int
}

func (r *JobRepository) List(filter JobFilter) ([]domain.Job, int64, error) {
	var jobs []domain.Job
	var total int64
	db := r.db.Model(&domain.Job{})

	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.CategoryID != nil {
		db = db.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.Type != "" {
		db = db.Where("type = ?", filter.Type)
	}
	if filter.Location != "" {
		db = db.Where("location ILIKE ?", "%"+filter.Location+"%")
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		db = db.Where("title ILIKE ? OR description ILIKE ?", s, s)
	}

	// count
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count jobs: %w", err)
	}

	// pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	off := (filter.Page - 1) * filter.Limit

	if err := db.Preload("Employer").Preload("Schedules").Order("created_at desc").Offset(off).Limit(filter.Limit).Find(&jobs).Error; err != nil {
		return nil, 0, fmt.Errorf("list jobs: %w", err)
	}

	return jobs, total, nil
}
