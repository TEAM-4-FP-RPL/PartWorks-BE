package routes

import (
	"net/http"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/handler"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/middleware"
)

func NewRouter(authH *handler.AuthHandler, jobH *handler.JobHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", authH.Register)
	mux.HandleFunc("/auth/login", authH.Login)

	jobsHandler := func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/jobs" || p == "/jobs/" {
			switch r.Method {
			case http.MethodGet:
				jobH.GetAll(w, r)
			case http.MethodPost:
				jobH.Create(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}
		// assume /jobs/{id}
		switch r.Method {
		case http.MethodGet:
			jobH.GetByID(w, r)
		case http.MethodPatch:
			jobH.Update(w, r)
		case http.MethodDelete:
			jobH.Delete(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
	mux.HandleFunc("/jobs", jobsHandler)
	mux.HandleFunc("/jobs/", jobsHandler)
	// add more routes here as project grows
	return middleware.CORS(middleware.Log(mux))
}
