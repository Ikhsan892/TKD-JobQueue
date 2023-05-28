package entity

import "time"

type GetProject struct {
	Id                 uint      `json:"id"`
	Name               string    `json:"name" gorm:"not null;type=text"`
	Code               string    `json:"project_code" gorm:"not null;max=15"`
	DueDate            time.Time `json:"project_date" gorm:"not null"`
	Type               string    `json:"type" gorm:"not null"`
	TotalAssignedUsers uint      `json:"total_assigned_users" gorm:"not null"`
	Status             string    `json:"status" gorm:"not null"`
	Description        *string   `json:"description" gorm:"text"`
	CompanyId          uint      `json:"company_id"`
}
