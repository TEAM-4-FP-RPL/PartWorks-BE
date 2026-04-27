package usecase

import (
    "context"
    "github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
    "github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/repository"
    "github.com/google/uuid"
)

type JobUsecase interface {
    GetEmployerJobs(ctx context.Context, employerID uuid.UUID) ([]domain.Job, error)
    DeleteJob(ctx context.Context, id uuid.UUID, employerID uuid.UUID) error 
}

type jobUsecase struct {
    jobRepo repository.JobRepository
}

func NewJobUsecase(repo repository.JobRepository) JobUsecase {
    return &jobUsecase{jobRepo: repo}
}

func (u *jobUsecase) GetEmployerJobs(ctx context.Context, employerID uuid.UUID) ([]domain.Job, error) {
    return u.jobRepo.GetByEmployerID(ctx, employerID)
}

func (u *jobUsecase) DeleteJob(ctx context.Context, id uuid.UUID, employerID uuid.UUID) error {
    return u.jobRepo.Delete(ctx, id, employerID)
}