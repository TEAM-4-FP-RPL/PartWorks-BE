package routes

import (
	"net/http"
	"strings"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/handler"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/middleware"
)

func NewRouter(authH *handler.AuthHandler, jobH *handler.JobHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", authH.Register)
	mux.HandleFunc("/auth/login", authH.Login)

	jobsHandler := func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasSuffix(p, "/apply") {
			if r.Method == http.MethodPost {
				jobH.Apply(w, r)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		p = r.URL.Path
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
	mux.HandleFunc("/employer/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			jobH.GetEmployerJobs(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/employer/jobs/", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/employer/jobs/" {
			if r.Method == http.MethodGet {
				jobH.GetEmployerJobs(w, r)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if strings.HasSuffix(strings.TrimSuffix(p, "/"), "/applications") {
			if r.Method == http.MethodGet {
				jobH.GetEmployerJobApplications(w, r)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/worker/applications", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			jobH.GetWorkerApplications(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/worker/applications/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			jobH.DeleteWorkerApplication(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/employer/applications", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			jobH.GetEmployerApplications(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/employer/applications/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(strings.TrimSuffix(r.URL.Path, "/"), "/status") {
			if r.Method == http.MethodPatch {
				jobH.PatchEmployerApplicationStatus(w, r)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/worker/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jobH.GetWorkerProfile(w, r)
		case http.MethodPut:
			jobH.PutWorkerProfile(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/worker/availability", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jobH.GetWorkerAvailability(w, r)
		case http.MethodPut:
			jobH.PutWorkerAvailability(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/worker/cvs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jobH.ListWorkerCVs(w, r)
		case http.MethodPost:
			jobH.UploadWorkerCV(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/worker/cvs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			jobH.DeleteWorkerCV(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/employer/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jobH.GetEmployerProfile(w, r)
		case http.MethodPut:
			jobH.PutEmployerProfile(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			jobH.GetCategories(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	return middleware.CORS(middleware.Log(mux))
}
