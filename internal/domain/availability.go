package domain

import "github.com/google/uuid"

type Availability struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	WorkerID  uuid.UUID `json:"worker_id" gorm:"type:uuid;not null;index"`
	Day       int       `json:"day" gorm:"type:smallint;not null"`
	StartTime string    `json:"start_time" gorm:"type:varchar(5);not null"`
	EndTime   string    `json:"end_time" gorm:"type:varchar(5);not null"`
}
