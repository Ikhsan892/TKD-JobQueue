package utils

import (
	"gorm.io/gorm"
	"os"
)

type LogInfo struct {
	gorm.Model
	ProcessName    string `json:"process_name"`
	ProcessPayload string `json:"process_payload"`
	ProcessStatus  string `json:"process_status"`
	ProcessType    string `json:"process_type"`
	ProcessResult  string `json:"process_result"`
	IssuedBy       uint   `json:"issued_by"`
}

type IJobQueueLog interface {
	CreateJobQueueLog(data LogInfo)
	UpdateJobQueueLog(data LogInfo)
}

type JobQueueLog struct {
	db *gorm.DB
	id uint
}

func NewJobQueueLog(db *gorm.DB) *JobQueueLog {
	return &JobQueueLog{
		db: db,
		id: 0,
	}
}

// CreateJobQueueLog create log process
func (jobQueue *JobQueueLog) CreateJobQueueLog(data LogInfo) {
	tx := jobQueue.db.Begin()
	if errCreate := tx.Table(os.Getenv("TABLE_LOG")).Create(&data).Error; errCreate != nil {
		Error("CREATE_TABLE_LOG_PROCESS", errCreate)
		tx.Rollback()
	} else {
		tx.Commit()
	}

	jobQueue.id = data.ID
}

// UpdateJobQueueLog update log process
func (jobQueue *JobQueueLog) UpdateJobQueueLog(data LogInfo) {
	tx := jobQueue.db.Begin()
	if errUpdate := tx.Table(os.Getenv("TABLE_LOG")).Where("id = ?", jobQueue.id).Updates(&data).Error; errUpdate != nil {
		Error("CREATE_TABLE_LOG_PROCESS", errUpdate)
		tx.Rollback()
	}
	tx.Commit()
}
