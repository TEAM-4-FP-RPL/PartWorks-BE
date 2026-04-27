package usecase

import (
	"os"
	"strconv"
	"time"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/repository"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/hash"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/jwt"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/validator"
	"github.com/google/uuid"
)

type AuthUsecase struct {
	repo   *repository.UserRepository
	jwtTTL time.Duration
}

func NewAuthUsecase(repo *repository.UserRepository) *AuthUsecase {
	ttlH := 72
	if v := os.Getenv("JWT_TTL_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttlH = n
		}
	}
	return &AuthUsecase{repo: repo, jwtTTL: time.Duration(ttlH) * time.Hour}
}

func (uc *AuthUsecase) Register(email, password, role string) (*domain.User, string, error) {
	if err := validator.ValidateEmail(email); err != nil {
		return nil, "", err
	}
	if err := validator.ValidatePassword(password); err != nil {
		return nil, "", err
	}
	if role == "" {
		role = string(domain.RoleWorker)
	}
	if role != string(domain.RoleWorker) && role != string(domain.RoleEmployer) {
		role = string(domain.RoleWorker)
	}

	if _, err := uc.repo.GetByEmail(email); err == nil {
		return nil, "", domain.ErrConflict
	} else if err != nil && err != domain.ErrNotFound {
		return nil, "", err
	}

	h, err := hash.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: &h,
		Role:         domain.UserRole(role),
	}

	if err := uc.repo.Create(user); err != nil {
		if err == domain.ErrConflict {
			return nil, "", domain.ErrConflict
		}
		return nil, "", err
	}

	token, err := jwt.GenerateToken(user.ID.String(), user.Email, string(user.Role), uc.jwtTTL)
	return user, token, err
}

func (uc *AuthUsecase) Login(email, password string) (*domain.User, string, error) {
	if err := validator.ValidateEmail(email); err != nil {
		return nil, "", err
	}
	if err := validator.ValidatePassword(password); err != nil {
		return nil, "", err
	}

	user, err := uc.repo.GetByEmail(email)
	if err != nil {
		return nil, "", err
	}
	if user.PasswordHash == nil {
		return nil, "", domain.ErrNotFound
	}
	if err := hash.CheckPassword(*user.PasswordHash, password); err != nil {
		return nil, "", domain.ErrNotFound
	}
	token, err := jwt.GenerateToken(user.ID.String(), user.Email, string(user.Role), uc.jwtTTL)
	return user, token, err
}
