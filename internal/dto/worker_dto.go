package dto

type WorkerAvailabilityDTO struct {
	Day       string `json:"day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type WorkerCVItemDTO struct {
	ID        string      `json:"id"`
	FileURL   string      `json:"file_url"`
	Category  CategoryDTO `json:"category"`
	CreatedAt string      `json:"created_at,omitempty"`
}

type WorkerProfileDTO struct {
	ID             string                  `json:"id"`
	FullName       string                  `json:"full_name"`
	PhoneNumber    string                  `json:"phone_number"`
	Bio            string                  `json:"bio"`
	Skills         string                  `json:"skills"`
	PhotoURL       string                  `json:"photo_url"`
	Availabilities []WorkerAvailabilityDTO `json:"availabilities"`
	CVs            []WorkerCVItemDTO       `json:"cvs"`
}

type UpdateWorkerProfileRequest struct {
	FullName    *string `json:"full_name,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	Skills      *string `json:"skills,omitempty"`
	PhotoURL    *string `json:"photo_url,omitempty"`
}

type UpdateWorkerAvailabilityRequest struct {
	Availabilities []WorkerAvailabilityDTO `json:"availabilities"`
}
