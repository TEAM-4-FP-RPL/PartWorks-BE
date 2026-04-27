package handler

import (
	"net/http"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Register Placeholder"))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login Placeholder"))
}