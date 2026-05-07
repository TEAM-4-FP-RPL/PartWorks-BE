package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkerProfile struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	User           *User          `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	FullName       string         `json:"full_name" gorm:"type:varchar(255);not null"`
	PhoneNumber    string         `json:"phone_number" gorm:"type:varchar(20)"`
	Bio            string         `json:"bio" gorm:"type:text"`
	Skills         string         `json:"skills" gorm:"type:text"`
	PhotoURL       string         `json:"photo_url" gorm:"type:text"`
	CVs            []WorkerCV     `json:"cvs,omitempty" gorm:"foreignKey:WorkerID;constraint:OnDelete:CASCADE"`
	Availabilities []Availability `json:"availabilities,omitempty" gorm:"foreignKey:WorkerID;constraint:OnDelete:CASCADE"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

type WorkerCV struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	WorkerID   uuid.UUID      `json:"worker_id" gorm:"type:uuid;not null;index"`
	Worker     *WorkerProfile `json:"worker,omitempty" gorm:"foreignKey:WorkerID;references:ID;constraint:OnDelete:CASCADE"`
	CategoryID int            `json:"category_id" gorm:"not null"`
	FileURL    string         `json:"file_url" gorm:"type:text;not null"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
}
