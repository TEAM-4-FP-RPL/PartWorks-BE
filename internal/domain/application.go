package domain

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	JobID     uuid.UUID      `json:"job_id" gorm:"type:uuid;not null;index"`
	Job       *Job           `json:"job,omitempty" gorm:"foreignKey:JobID;references:ID;constraint:OnDelete:CASCADE"`
	WorkerID  uuid.UUID      `json:"worker_id" gorm:"type:uuid;not null;index"`
	Worker    *WorkerProfile `json:"worker,omitempty" gorm:"foreignKey:WorkerID;references:ID;constraint:OnDelete:CASCADE"`
	CVID      *uuid.UUID     `json:"cv_id" gorm:"type:uuid;index"`
	CV        *WorkerCV      `json:"cv,omitempty" gorm:"foreignKey:CVID;references:ID;constraint:OnDelete:SET NULL"`
	Status    string         `json:"status" gorm:"type:varchar(20);default:'pending'"`
	CoverNote string         `json:"cover_note" gorm:"type:text"`
	AppliedAt time.Time      `json:"applied_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}
