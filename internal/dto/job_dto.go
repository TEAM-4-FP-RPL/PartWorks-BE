package dto

import (
	"time"

	"github.com/google/uuid"
)


type CategoryDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
}

type EmployerDTO struct {
	ID          string `json:"id"`
	CompanyName string `json:"company_name"`
	LogoURL     string `json:"logo_url"`
}

type JobScheduleDTO struct {
	Day       string `json:"day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type JobResponse struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	Type      string           `json:"type"`
	Status    string           `json:"status"`
	Salary    int64            `json:"salary"`
	Location  string           `json:"location"`
	Category  CategoryDTO      `json:"category"`
	Employer  EmployerDTO      `json:"employer"`
	Schedules []JobScheduleDTO `json:"schedules"`
	CreatedAt string           `json:"created_at"`
}


type JobScheduleRequest struct {
	Day       string `json:"day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type CreateJobRequest struct {
	CategoryID  int                  `json:"category_id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Type        string               `json:"type"`
	Salary      int64                `json:"salary"`
	Location    string               `json:"location"`
	Schedules   []JobScheduleRequest `json:"schedules"`
}

type JobDataResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateJobResponse struct {
	Data    JobDataResponse `json:"data"`
	Message string          `json:"message"`
}
