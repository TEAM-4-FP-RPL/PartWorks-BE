package routes

import (
	"net/http"
	"strings"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/handler"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/middleware"
)

func NewRouter(authH *handler.AuthHandler, jobH *handler.JobHandler) http.Handler {
	mux := http.NewServeMux()
	registerPublicRoutes(mux, authH, jobH)
	registerJobRoutes(mux, jobH)
	registerEmployerRoutes(mux, jobH)
	registerWorkerRoutes(mux, jobH)

	return middleware.CORS(middleware.Log(mux))
}

func registerPublicRoutes(mux *http.ServeMux, authH *handler.AuthHandler, jobH *handler.JobHandler) {
	mux.HandleFunc("/auth/register", authH.Register)
	mux.HandleFunc("/auth/login", authH.Login)
	mux.HandleFunc("/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			jobH.GetCategories(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy","service":"partworks-be"}`))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

func registerJobRoutes(mux *http.ServeMux, jobH *handler.JobHandler) {
	jobsHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/apply") {
			if r.Method == http.MethodPost {
				jobH.Apply(w, r)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path == "/jobs" || r.URL.Path == "/jobs/" {
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
}

func registerEmployerRoutes(mux *http.ServeMux, jobH *handler.JobHandler) {
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
		// Check for /applications/{application_id}/status endpoint
		if strings.HasSuffix(strings.TrimSuffix(p, "/"), "/status") {
			if r.Method == http.MethodPatch {
				jobH.PatchEmployerJobApplicationStatus(w, r)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
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
	mux.HandleFunc("/employer/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jobH.GetEmployerProfile(w, r)
		case http.MethodPatch:
			jobH.PatchEmployerProfile(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func registerWorkerRoutes(mux *http.ServeMux, jobH *handler.JobHandler) {
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
	mux.HandleFunc("/worker/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jobH.GetWorkerProfile(w, r)
		case http.MethodPatch:
			jobH.PatchWorkerProfile(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/worker/availability", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jobH.GetWorkerAvailability(w, r)
		case http.MethodPatch:
			jobH.PatchWorkerAvailability(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	workerCVsCollection := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jobH.ListWorkerCVs(w, r)
		case http.MethodPost:
			jobH.UploadWorkerCV(w, r)
		case http.MethodPatch:
			jobH.PatchWorkerCVs(w, r)
		case http.MethodDelete:
			jobH.DeleteWorkerCVs(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
	mux.HandleFunc("/worker/cvs", workerCVsCollection)
	mux.HandleFunc("/worker/cvs/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimSuffix(r.URL.Path, "/")
		if p == "/worker/cvs" {
			workerCVsCollection(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			jobH.DeleteWorkerCV(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}
