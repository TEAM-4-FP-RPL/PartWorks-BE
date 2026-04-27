package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleWorker   UserRole = "worker"
	RoleEmployer UserRole = "employer"
)

type User struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Email         string    `json:"email" gorm:"type:varchar(255);unique;not null"`
	PasswordHash  *string   `json:"password_hash,omitempty" gorm:"type:varchar(255)"`
	OAuthProvider *string   `json:"oauth_provider,omitempty" gorm:"type:varchar(50)"`
	OAuthID       *string   `json:"oauth_id,omitempty" gorm:"type:varchar(255)"`
	Role          UserRole  `json:"role" gorm:"type:varchar(20);not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
