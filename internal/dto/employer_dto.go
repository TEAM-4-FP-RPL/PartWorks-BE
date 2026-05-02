package dto

type EmployerProfileDTO struct {
	ID          string `json:"id"`
	CompanyName string `json:"company_name"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
}

type UpdateEmployerProfileRequest struct {
	CompanyName *string `json:"company_name,omitempty"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
}
