package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/dto"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrInvalidID = errors.New("invalid id")
var ErrForbidden = errors.New("forbidden")
var ErrCategoryExists = errors.New("category already exists")

var dayToInt = map[string]int{
	"monday":    1,
	"tuesday":   2,
	"wednesday": 3,
	"thursday":  4,
	"friday":    5,
	"saturday":  6,
	"sunday":    7,
}

func dayName(n int) string {
	switch n {
	case 1:
		return "monday"
	case 2:
		return "tuesday"
	case 3:
		return "wednesday"
	case 4:
		return "thursday"
	case 5:
		return "friday"
	case 6:
		return "saturday"
	case 7:
		return "sunday"
	default:
		return "unknown"
	}
}

type JobUsecase struct {
	repo     *repository.JobRepository
	userRepo *repository.UserRepository
}

func NewJobUsecase(repo *repository.JobRepository, userRepo *repository.UserRepository) *JobUsecase {
	return &JobUsecase{repo: repo, userRepo: userRepo}
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

func (uc *JobUsecase) Create(userIDStr string, req dto.CreateJobRequest) (*domain.Job, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if string(user.Role) != string(domain.RoleEmployer) {
		return nil, fmt.Errorf("user is not employer")
	}

	if _, err := uc.repo.GetCategoryByID(req.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("check category: %w", err)
	}

	emp, err := uc.repo.GetEmployerByUserID(uid)
	var employerID uuid.UUID
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			created, cerr := uc.repo.CreateEmployerProfile(uid)
			if cerr != nil {
				return nil, fmt.Errorf("create employer profile: %w", cerr)
			}
			employerID = created.ID
		} else {
			return nil, fmt.Errorf("get employer: %w", err)
		}
	} else {
		employerID = emp.ID
	}

	j := &domain.Job{
		ID:          uuid.New(),
		EmployerID:  employerID,
		CategoryID:  req.CategoryID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Type:        domain.JobType(req.Type),
		Status:      domain.JobStatus("open"),
		Salary:      req.Salary,
		Location:    strings.TrimSpace(req.Location),
		CreatedAt:   time.Now(),
	}

	for _, s := range req.Schedules {
		day := strings.ToLower(strings.TrimSpace(s.Day))
		n, ok := dayToInt[day]
		if !ok {
			return nil, fmt.Errorf("invalid day: %s", s.Day)
		}
		j.Schedules = append(j.Schedules, domain.JobSchedule{ID: 0, JobID: j.ID, Day: n, StartTime: s.StartTime, EndTime: s.EndTime})
	}

	if err := uc.repo.Create(j); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}
	return j, nil
}

func (uc *JobUsecase) Update(userIDStr string, jobIDStr string, req dto.UpdateJobRequest) (*domain.Job, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if string(user.Role) != string(domain.RoleEmployer) {
		return nil, fmt.Errorf("user is not employer")
	}

	jid, err := uuid.Parse(jobIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid job id: %w", err)
	}
	job, _, err := uc.repo.GetByID(jid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get job: %w", err)
	}
	if job == nil {
		return nil, gorm.ErrRecordNotFound
	}
	owned := false
	if job.EmployerID == uid {
		owned = true
	}
	if job.Employer != nil && job.Employer.UserID == uid {
		owned = true
	}
	if !owned {
		if empProfile, err := uc.repo.GetEmployerByUserID(uid); err == nil && empProfile != nil && empProfile.ID == job.EmployerID {
			owned = true
		}
	}
	if !owned {
		return nil, ErrForbidden
	}

	if req.Title != nil {
		job.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		job.Description = strings.TrimSpace(*req.Description)
	}
	if req.Type != nil {
		job.Type = domain.JobType(*req.Type)
	}
	if req.Salary != nil {
		job.Salary = *req.Salary
	}
	if req.Location != nil {
		job.Location = strings.TrimSpace(*req.Location)
	}
	if req.Status != nil {
		job.Status = domain.JobStatus(*req.Status)
	}
	job.UpdatedAt = time.Now()

	var schedules []domain.JobSchedule
	if req.Schedules != nil {
		for _, s := range *req.Schedules {
			day := strings.ToLower(strings.TrimSpace(s.Day))
			n, ok := dayToInt[day]
			if !ok {
				return nil, fmt.Errorf("invalid day: %s", s.Day)
			}
			schedules = append(schedules, domain.JobSchedule{ID: 0, JobID: job.ID, Day: n, StartTime: s.StartTime, EndTime: s.EndTime})
		}
		if err := uc.repo.ReplaceSchedules(job.ID, schedules); err != nil {
			return nil, fmt.Errorf("replace schedules: %w", err)
		}
	}

	if err := uc.repo.Update(job); err != nil {
		return nil, fmt.Errorf("update job: %w", err)
	}
	return job, nil
}

func (uc *JobUsecase) Delete(userIDStr string, jobIDStr string) error {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if string(user.Role) != string(domain.RoleEmployer) {
		return fmt.Errorf("user is not employer")
	}
	jid, err := uuid.Parse(jobIDStr)
	if err != nil {
		return fmt.Errorf("invalid job id: %w", err)
	}
	job, _, err := uc.repo.GetByID(jid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return fmt.Errorf("get job: %w", err)
	}
	if job == nil {
		return gorm.ErrRecordNotFound
	}
	owned := false
	if job.EmployerID == uid {
		owned = true
	}
	if job.Employer != nil && job.Employer.UserID == uid {
		owned = true
	}
	if !owned {
		if empProfile, err := uc.repo.GetEmployerByUserID(uid); err == nil && empProfile != nil && empProfile.ID == job.EmployerID {
			owned = true
		}
	}
	if !owned {
		return ErrForbidden
	}
	if err := uc.repo.Delete(job.ID); err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	return nil
}

func (uc *JobUsecase) GetEmployerJobs(userIDStr string, filter JobFilter) ([]dto.JobResponse, int64, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, 0, fmt.Errorf("get user: %w", err)
	}
	if string(user.Role) != string(domain.RoleEmployer) {
		return nil, 0, fmt.Errorf("user is not employer")
	}
	emp, err := uc.repo.GetEmployerByUserID(uid)
	var employerID uuid.UUID
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			employerID = uid
		} else {
			return nil, 0, fmt.Errorf("get employer: %w", err)
		}
	} else {
		employerID = emp.ID
	}
	jobs, total, err := uc.repo.ListByEmployer(employerID, repository.JobFilter{
		Status: filter.Status,
		Page:   filter.Page,
		Limit:  filter.Limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list employer jobs: %w", err)
	}
	out := make([]dto.JobResponse, 0, len(jobs))
	for _, j := range jobs {
		catDTO := dto.CategoryDTO{ID: j.CategoryID}
		if c, err := uc.repo.GetCategoryByID(j.CategoryID); err == nil && c != nil {
			catDTO.Name = c.Name
		}
		empDTO := dto.EmployerDTO{}
		if j.Employer != nil {
			empDTO.ID = j.Employer.ID.String()
			empDTO.CompanyName = j.Employer.CompanyName
			empDTO.LogoURL = j.Employer.LogoURL
		}
		schedules := make([]dto.JobScheduleDTO, 0, len(j.Schedules))
		for _, s := range j.Schedules {
			schedules = append(schedules, dto.JobScheduleDTO{Day: dayName(s.Day), StartTime: s.StartTime, EndTime: s.EndTime})
		}
		count, _ := uc.repo.CountApplications(j.ID)
		out = append(out, dto.JobResponse{
			ID:              j.ID.String(),
			Title:           j.Title,
			Type:            string(j.Type),
			Status:          string(j.Status),
			Salary:          j.Salary,
			Location:        j.Location,
			Category:        catDTO,
			Employer:        empDTO,
			TotalApplicants: count,
			Schedules:       schedules,
			CreatedAt:       j.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, total, nil
}

func (uc *JobUsecase) CreateCategory(userIDStr string, req dto.CreateCategoryRequest) (*domain.Category, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if string(user.Role) != string(domain.RoleEmployer) {
		return nil, fmt.Errorf("user is not employer")
	}

	// Check if category with same name already exists
	existing, err := uc.repo.GetCategoryByName(req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check existing category: %w", err)
	}
	if existing != nil {
		return nil, ErrCategoryExists
	}

	// Normalize name: trim and lowercase for storage
	name := strings.TrimSpace(req.Name)
	// Store with proper case but check case-insensitive

	cat := &domain.Category{
		ID:   0,
		Name: name,
	}

	if err := uc.repo.CreateCategory(cat); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return cat, nil
}
