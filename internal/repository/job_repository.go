package repository

import (
	"fmt"
	"strings"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository { return &JobRepository{db: db} }

func (r *JobRepository) GetEmployerByUserID(userID uuid.UUID) (*domain.EmployerProfile, error) {
	var emp domain.EmployerProfile
	if err := r.db.Where("user_id = ?", userID).First(&emp).Error; err == nil {
		return &emp, nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("get employer by user id: %w", err)
	}

	if err := r.db.First(&emp, "id = ?", userID).Error; err == nil {
		return &emp, nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("get employer by id fallback: %w", err)
	}

	return nil, gorm.ErrRecordNotFound
}

func (r *JobRepository) CreateEmployerProfile(userID uuid.UUID) (*domain.EmployerProfile, error) {
	emp := &domain.EmployerProfile{
		ID:          userID,
		UserID:      userID,
		CompanyName: "",
	}
	if err := r.db.Create(emp).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			var e domain.EmployerProfile
			if err2 := r.db.Where("user_id = ?", userID).First(&e).Error; err2 == nil {
				return &e, nil
			}
		}
		return nil, fmt.Errorf("create employer profile: %w", err)
	}
	return emp, nil
}

func (r *JobRepository) Create(job *domain.Job) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Create(job).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (r *JobRepository) Update(job *domain.Job) error {
	if err := r.db.Save(job).Error; err != nil {
		return err
	}
	return nil
}

func (r *JobRepository) ReplaceSchedules(jobID uuid.UUID, schedules []domain.JobSchedule) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("job_id = ?", jobID).Delete(&domain.JobSchedule{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(schedules) > 0 {
		for i := range schedules {
			schedules[i].JobID = jobID
		}
		if err := tx.Create(&schedules).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (r *JobRepository) GetCategoryByID(id int) (*domain.Category, error) {
	var c domain.Category
	if err := r.db.First(&c, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &c, nil
}

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

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count jobs: %w", err)
	}

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

func (r *JobRepository) GetByID(id uuid.UUID) (*domain.Job, *domain.Category, error) {
	var job domain.Job
	if err := r.db.Preload("Employer").Preload("Schedules").First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, gorm.ErrRecordNotFound
		}
		return nil, nil, fmt.Errorf("get job by id: %w", err)
	}
	var cat domain.Category
	if err := r.db.First(&cat, job.CategoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &job, nil, nil
		}
		return nil, nil, fmt.Errorf("get category: %w", err)
	}
	return &job, &cat, nil
}
