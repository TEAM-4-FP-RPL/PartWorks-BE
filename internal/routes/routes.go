package routes

import (
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/handler"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewRouter(authH *handler.AuthHandler, jobH *handler.JobHandler) http.Handler {
	r := gin.Default()

	r.POST("/auth/register", func(c *gin.Context) { c.String(200, "Register Placeholder") })
	r.POST("/auth/login", func(c *gin.Context) { c.String(200, "Login Placeholder") })
	r.GET("/employer/jobs", jobH.GetEmployerJobs)
	r.DELETE("/employer/jobs/:id", jobH.DeleteJob)

	return r
}