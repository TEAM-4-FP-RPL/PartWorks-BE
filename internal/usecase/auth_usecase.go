package usecase

import (
	"os"
	"strconv"
	"time"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/repository"
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

	/* if err := validator.ValidateEmail(email); err != nil {
		return nil, "", err
	}
	*/

	if role == "" {
		role = string(domain.RoleWorker)
	}

	if _, err := uc.repo.GetByEmail(email); err == nil {
		return nil, "", domain.ErrConflict
	}

	h := "hashed_password_placeholder"

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: &h,
		Role:         domain.UserRole(role),
	}

	if err := uc.repo.Create(user); err != nil {
		return nil, "", err
	}

	return user, "dummy-token-register", nil
}

func (uc *AuthUsecase) Login(email, password string) (*domain.User, string, error) {
	user, err := uc.repo.GetByEmail(email)
	if err != nil {
		return nil, "", err
	}

	return user, "dummy-token-login", nil
}