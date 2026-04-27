package handler

import (
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthHandler struct {
	usecase *usecase.AuthUsecase 
}

func NewAuthHandler(uc *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: uc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Register endpoint active"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Login endpoint active"})
}