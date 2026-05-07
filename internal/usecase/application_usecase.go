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

var ErrInvalidStatus = errors.New("invalid status")

type ApplicationFilter struct {
	Status string
	Page   int
	Limit  int
}

func (uc *JobUsecase) Apply(userIDStr string, jobIDStr string, cvIDStr string, coverNote string) (*domain.Application, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if string(user.Role) != string(domain.RoleWorker) {
		return nil, fmt.Errorf("user is not worker")
	}

	jid, err := uuid.Parse(jobIDStr)
	if err != nil {
		return nil, ErrInvalidID
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

	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("worker profile not found")
		}
		return nil, fmt.Errorf("get worker: %w", err)
	}

	cvID, err := uuid.Parse(cvIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cv id: %w", err)
	}
	cv, err := uc.repo.GetCVByID(cvID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("cv not found")
		}
		return nil, fmt.Errorf("get cv: %w", err)
	}
	if cv.WorkerID != wp.ID {
		return nil, fmt.Errorf("cv does not belong to worker")
	}

	has, err := uc.repo.HasApplication(job.ID, wp.ID)
	if err != nil {
		return nil, fmt.Errorf("check application: %w", err)
	}
	if has {
		return nil, fmt.Errorf("already applied to this job")
	}

	app := &domain.Application{
		ID:        uuid.New(),
		JobID:     job.ID,
		WorkerID:  wp.ID,
		CVID:      &cvID,
		Status:    "pending",
		CoverNote: coverNote,
		AppliedAt: time.Now(),
	}

	if err := uc.repo.CreateApplication(app); err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	return app, nil
}

func (uc *JobUsecase) ListWorkerApplications(userIDStr string, filter ApplicationFilter) ([]domain.Application, map[int]domain.Category, int64, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get user: %w", err)
	}
	if string(user.Role) != string(domain.RoleWorker) {
		return nil, nil, 0, fmt.Errorf("user is not worker")
	}

	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, 0, fmt.Errorf("worker profile not found")
		}
		return nil, nil, 0, fmt.Errorf("get worker: %w", err)
	}

	apps, total, err := uc.repo.ListWorkerApplications(wp.ID, filter.Status, filter.Page, filter.Limit)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("list applications: %w", err)
	}

	uniq := make(map[int]struct{})
	ids := make([]int, 0)
	for i := range apps {
		if apps[i].CV != nil {
			cid := apps[i].CV.CategoryID
			if _, ok := uniq[cid]; !ok {
				uniq[cid] = struct{}{}
				ids = append(ids, cid)
			}
		}
	}
	cats, err := uc.repo.GetCategoriesByIDs(ids)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get categories: %w", err)
	}

	return apps, cats, total, nil
}

func (uc *JobUsecase) ListEmployerApplications(userIDStr string, filter ApplicationFilter) ([]domain.Application, map[int]domain.Category, int64, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get user: %w", err)
	}
	if string(user.Role) != string(domain.RoleEmployer) {
		return nil, nil, 0, fmt.Errorf("user is not employer")
	}

	emp, err := uc.repo.GetEmployerByUserID(uid)
	var employerID uuid.UUID
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			employerID = uid
		} else {
			return nil, nil, 0, fmt.Errorf("get employer: %w", err)
		}
	} else {
		employerID = emp.ID
	}

	apps, total, err := uc.repo.ListEmployerApplications(employerID, filter.Status, filter.Page, filter.Limit)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("list employer applications: %w", err)
	}

	uniq := make(map[int]struct{})
	ids := make([]int, 0)
	for i := range apps {
		if apps[i].CV != nil {
			cid := apps[i].CV.CategoryID
			if _, ok := uniq[cid]; !ok {
				uniq[cid] = struct{}{}
				ids = append(ids, cid)
			}
		}
	}
	cats, err := uc.repo.GetCategoriesByIDs(ids)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get categories: %w", err)
	}

	return apps, cats, total, nil
}

func (uc *JobUsecase) ListEmployerJobApplications(userIDStr, jobIDStr string, filter ApplicationFilter) ([]domain.Application, map[int]domain.Category, int64, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get user: %w", err)
	}
	if string(user.Role) != string(domain.RoleEmployer) {
		return nil, nil, 0, fmt.Errorf("user is not employer")
	}

	jid, err := uuid.Parse(jobIDStr)
	if err != nil {
		return nil, nil, 0, ErrInvalidID
	}

	emp, err := uc.repo.GetEmployerByUserID(uid)
	var employerID uuid.UUID
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			employerID = uid
		} else {
			return nil, nil, 0, fmt.Errorf("get employer: %w", err)
		}
	} else {
		employerID = emp.ID
	}

	job, _, err := uc.repo.GetByID(jid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, 0, gorm.ErrRecordNotFound
		}
		return nil, nil, 0, fmt.Errorf("get job: %w", err)
	}
	if job == nil {
		return nil, nil, 0, gorm.ErrRecordNotFound
	}
	if job.EmployerID != employerID {
		return nil, nil, 0, ErrForbidden
	}

	apps, total, err := uc.repo.ListJobApplications(jid, filter.Status, filter.Page, filter.Limit)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("list job applications: %w", err)
	}

	uniq := make(map[int]struct{})
	ids := make([]int, 0)
	for i := range apps {
		if apps[i].CV != nil {
			cid := apps[i].CV.CategoryID
			if _, ok := uniq[cid]; !ok {
				uniq[cid] = struct{}{}
				ids = append(ids, cid)
			}
		}
	}
	cats, err := uc.repo.GetCategoriesByIDs(ids)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get categories: %w", err)
	}

	return apps, cats, total, nil
}

func (uc *JobUsecase) UpdateEmployerApplicationStatus(userIDStr, applicationIDStr, status string) (*domain.Application, error) {
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

	st := strings.ToLower(strings.TrimSpace(status))
	switch st {
	case "pending", "accepted", "rejected":
	default:
		return nil, ErrInvalidStatus
	}

	emp, err := uc.repo.GetEmployerByUserID(uid)
	var employerID uuid.UUID
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			employerID = uid
		} else {
			return nil, fmt.Errorf("get employer: %w", err)
		}
	} else {
		employerID = emp.ID
	}

	appID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		return nil, ErrInvalidID
	}

	app, err := uc.repo.GetApplicationByID(appID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get application: %w", err)
	}

	job, _, err := uc.repo.GetByID(app.JobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get job: %w", err)
	}
	if job == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if job.EmployerID != employerID {
		return nil, ErrForbidden
	}

	updated, err := uc.repo.UpdateApplicationStatus(appID, st)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("update status: %w", err)
	}
	return updated, nil
}

func (uc *JobUsecase) WithdrawWorkerApplication(userIDStr, applicationIDStr string) error {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	user, err := uc.userRepo.GetByID(uid)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if string(user.Role) != string(domain.RoleWorker) {
		return fmt.Errorf("user is not worker")
	}

	wp, err := uc.repo.GetWorkerByUserID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("worker profile not found")
		}
		return fmt.Errorf("get worker: %w", err)
	}

	appID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		return ErrInvalidID
	}

	app, err := uc.repo.GetApplicationByID(appID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return fmt.Errorf("get application: %w", err)
	}
	if app.WorkerID != wp.ID {
		return ErrForbidden
	}

	if err := uc.repo.DeleteWorkerApplication(appID, wp.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return fmt.Errorf("delete application: %w", err)
	}
	return nil
}
