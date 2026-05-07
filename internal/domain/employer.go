package domain

import (
	"time"

	"github.com/google/uuid"
)

type EmployerProfile struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	CompanyName string    `json:"company_name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	LogoURL     string    `json:"logo_url" gorm:"type:text"`
	Jobs        []Job     `json:"jobs,omitempty" gorm:"foreignKey:EmployerID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
