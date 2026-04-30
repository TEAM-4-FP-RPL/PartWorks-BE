package usecase

import (
	"errors"
	"fmt"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrInvalidID = errors.New("invalid id")

type JobUsecase struct {
	repo *repository.JobRepository
}

func NewJobUsecase(repo *repository.JobRepository) *JobUsecase { return &JobUsecase{repo: repo} }

type JobFilter struct {
	Search     string
	CategoryID *int
	Type       string
	Location   string
	Status     string
	Page       int
	Limit      int
}

func (uc *JobUsecase) GetAll(filter JobFilter) ([]domain.Job, int64, error) {
	jobs, total, err := uc.repo.List(repository.JobFilter{
		Search:     filter.Search,
		CategoryID: filter.CategoryID,
		Type:       filter.Type,
		Location:   filter.Location,
		Status:     filter.Status,
		Page:       filter.Page,
		Limit:      filter.Limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("usecase get all jobs: %w", err)
	}
	return jobs, total, nil
}

func (uc *JobUsecase) GetByID(idStr string) (*domain.Job, *domain.Category, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, nil, ErrInvalidID
	}
	job, cat, err := uc.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, gorm.ErrRecordNotFound
		}
		return nil, nil, fmt.Errorf("usecase get job by id: %w", err)
	}
	if job == nil {
		return nil, nil, gorm.ErrRecordNotFound
	}
	return job, cat, nil
}
