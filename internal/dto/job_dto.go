package dto

type CategoryDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
}

type EmployerDTO struct {
	ID          string `json:"id"`
	CompanyName string `json:"company_name"`
	LogoURL     string `json:"logo_url"`
	Description string `json:"description,omitempty"`
}

type JobScheduleDTO struct {
	Day       string `json:"day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type CreateJobScheduleRequest struct {
	Day       string `json:"day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type CreateJobRequest struct {
	CategoryID  int                        `json:"category_id"`
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Type        string                     `json:"type"`
	Salary      int64                      `json:"salary"`
	Location    string                     `json:"location"`
	Schedules   []CreateJobScheduleRequest `json:"schedules"`
}

// UpdateJobRequest uses pointer fields for optional values
type UpdateJobRequest struct {
	Title       *string                     `json:"title,omitempty"`
	Description *string                     `json:"description,omitempty"`
	Type        *string                     `json:"type,omitempty"`
	Salary      *int64                      `json:"salary,omitempty"`
	Location    *string                     `json:"location,omitempty"`
	Status      *string                     `json:"status,omitempty"`
	Schedules   *[]CreateJobScheduleRequest `json:"schedules,omitempty"`
}

type JobResponse struct {
	ID               string           `json:"id"`
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	Type             string           `json:"type"`
	Status           string           `json:"status"`
	Salary           int64            `json:"salary"`
	Location         string           `json:"location"`
	Category         CategoryDTO      `json:"category"`
	Employer         EmployerDTO      `json:"employer"`
	TotalApplicants  int64            `json:"total_applicants"`
	Schedules        []JobScheduleDTO `json:"schedules"`
	WorkHoursPerWeek int              `json:"work_hours_per_week"`
	CreatedAt        string           `json:"created_at"`
}
