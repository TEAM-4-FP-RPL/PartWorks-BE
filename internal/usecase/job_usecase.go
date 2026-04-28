package usecase
import (
	"fmt"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/repository"
	"github.com/google/uuid"
)

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

func (uc *JobUsecase) Update(jobID, employerID uuid.UUID, req *domain.Job) (*domain.Job, error) {
	existing, err := uc.repo.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if existing.EmployerID != employerID {
		return nil, domain.ErrForbidden
	}
	existing.Title = req.Title
	existing.Salary = req.Salary
	existing.Status = req.Status
	existing.Schedules = req.Schedules
	for i := range existing.Schedules {
		existing.Schedules[i].JobID = jobID
	}
	if err := uc.repo.Update(existing); err != nil {
		return nil, fmt.Errorf("usecase update job: %w", err)
	}
	return existing, nil
}

func (uc *JobUsecase) Create(job *domain.Job) error {
	job.ID = uuid.New()
	job.Status = domain.JobStatusOpen
	
	for i := range job.Schedules {
		job.Schedules[i].JobID = job.ID
	}

	if err := uc.repo.Create(job); err != nil {
		return err
	}

	return nil
}
