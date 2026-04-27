package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/dto"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/usecase"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/httperror"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/response"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/validator"
)

type AuthHandler struct {
	uc *usecase.AuthUsecase
}

func NewAuthHandler(uc *usecase.AuthUsecase) *AuthHandler { return &AuthHandler{uc: uc} }

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validator.ValidateEmail(req.Email); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validator.ValidatePassword(req.Password); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	user, token, err := h.uc.Register(req.Email, req.Password, req.Role)
	if err != nil {
		if handled := httperror.HandleCreateConflict(w, err, "email already registered"); handled {
			return
		}
		if err == domain.ErrNotFound {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	response.JSON(w, http.StatusCreated, dto.AuthResponse{Token: token, ID: user.ID.String(), Email: user.Email, Role: string(user.Role)})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validator.ValidateEmail(req.Email); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validator.ValidatePassword(req.Password); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	user, token, err := h.uc.Login(req.Email, req.Password)
	if err != nil {
		if err == domain.ErrNotFound {
			response.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		log.Printf("auth: login error: %v", err)
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	response.JSON(w, http.StatusOK, dto.AuthResponse{Token: token, ID: user.ID.String(), Email: user.Email, Role: string(user.Role)})
}
