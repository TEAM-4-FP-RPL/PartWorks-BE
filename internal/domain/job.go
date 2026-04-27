package domain

import (
	"time"

	"github.com/google/uuid"
)

type JobType string

type JobStatus string

const (
	JobTypeOnsite JobType = "onsite"
	JobTypeRemote JobType = "remote"
	JobTypeHybrid JobType = "hybrid"

	JobStatusOpen   JobStatus = "open"
	JobStatusClosed JobStatus = "closed"
)

type Job struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey"`
	EmployerID  uuid.UUID        `json:"employer_id" gorm:"type:uuid;not null;index"`
	Employer    *EmployerProfile `json:"employer,omitempty" gorm:"foreignKey:EmployerID;references:ID;constraint:OnDelete:CASCADE"`
	CategoryID  int              `json:"category_id" gorm:"not null;index"`
	Title       string           `json:"title" gorm:"type:varchar(255);not null"`
	Description string           `json:"description" gorm:"type:text;not null"`
	Type        JobType          `json:"type" gorm:"type:varchar(20)"`
	Status      JobStatus        `json:"status" gorm:"type:varchar(20);default:'open'"`
	Salary      int64            `json:"salary" gorm:"type:bigint"`
	Location    string           `json:"location" gorm:"type:varchar(255)"`
	Schedules   []JobSchedule    `json:"schedules,omitempty" gorm:"foreignKey:JobID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

type JobSchedule struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	JobID     uuid.UUID `json:"job_id" gorm:"type:uuid;not null;index"`
	Job       *Job      `json:"job,omitempty" gorm:"foreignKey:JobID;references:ID;constraint:OnDelete:CASCADE"`
	Day       int       `json:"day" gorm:"type:smallint;not null"`
	StartTime string    `json:"start_time" gorm:"type:varchar(5);not null"`
	EndTime   string    `json:"end_time" gorm:"type:varchar(5);not null"`
}
