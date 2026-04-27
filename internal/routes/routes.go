package routes

import (
	"net/http"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/handler"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/middleware"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/response"
)

func NewRouter(authH *handler.AuthHandler, jobH *handler.JobHandler) http.Handler {
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/auth/register", authH.Register)
	mux.HandleFunc("/auth/login", authH.Login)

	// Job routes
	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jobH.GetAll(w, r)
		case http.MethodPost:
			// Protected route: Auth and Employer role required
			middleware.Auth(middleware.Roles("employer")(http.HandlerFunc(jobH.Create))).ServeHTTP(w, r)
		default:
			response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	return middleware.CORS(middleware.Log(mux))
}
