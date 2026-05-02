package dto

type WorkerApplicationsJobEmployerDTO struct {
	CompanyName string `json:"company_name"`
	LogoURL     string `json:"logo_url"`
}

type WorkerApplicationsJobDTO struct {
	ID       string                           `json:"id"`
	Title    string                           `json:"title"`
	Type     string                           `json:"type"`
	Location string                           `json:"location"`
	Employer WorkerApplicationsJobEmployerDTO `json:"employer"`
}

type WorkerApplicationsCVCategoryDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type WorkerApplicationsCVDTo struct {
	ID       string                          `json:"id"`
	Category WorkerApplicationsCVCategoryDTO `json:"category"`
	FileURL  string                          `json:"file_url"`
}

type WorkerApplicationItemDTO struct {
	ID        string                   `json:"id"`
	Status    string                   `json:"status"`
	CoverNote string                   `json:"cover_note"`
	AppliedAt string                   `json:"applied_at"`
	Job       WorkerApplicationsJobDTO `json:"job"`
	CV        *WorkerApplicationsCVDTo `json:"cv"`
}

type EmployerApplicationsJobDTO struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type EmployerApplicationsWorkerDTO struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	PhotoURL string `json:"photo_url"`
}

type EmployerApplicationsCVCategoryDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type EmployerApplicationsCVDTO struct {
	ID       string                            `json:"id"`
	FileURL  string                            `json:"file_url"`
	Category EmployerApplicationsCVCategoryDTO `json:"category"`
}

type EmployerApplicationItemDTO struct {
	ID        string                        `json:"id"`
	Status    string                        `json:"status"`
	AppliedAt string                        `json:"applied_at"`
	Job       EmployerApplicationsJobDTO    `json:"job"`
	Worker    EmployerApplicationsWorkerDTO `json:"worker"`
	CV        *EmployerApplicationsCVDTO    `json:"cv"`
}

type EmployerJobApplicationsWorkerDTO struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	PhotoURL string `json:"photo_url"`
	Bio      string `json:"bio"`
	Skills   string `json:"skills"`
}

type EmployerJobApplicationItemDTO struct {
	ID        string                           `json:"id"`
	Status    string                           `json:"status"`
	CoverNote string                           `json:"cover_note"`
	AppliedAt string                           `json:"applied_at"`
	Worker    EmployerJobApplicationsWorkerDTO `json:"worker"`
	CV        *EmployerApplicationsCVDTO       `json:"cv"`
}
