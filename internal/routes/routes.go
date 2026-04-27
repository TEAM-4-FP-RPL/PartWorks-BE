package routes

import (
	"net/http"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/handler"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/middleware"
)

func NewRouter(authH *handler.AuthHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", authH.Register)
	mux.HandleFunc("/auth/login", authH.Login)
	// add more routes here as project grows
	return middleware.CORS(middleware.Log(mux))
}
