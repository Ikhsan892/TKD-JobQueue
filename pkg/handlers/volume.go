package handlers

import (
	"assessment/configs"
	"assessment/pkg/adapter/postgresql"
	"assessment/pkg/functions"
	"assessment/pkg/utils"
	"fmt"
	"github.com/adjust/rmq/v5"
	"gorm.io/gorm"
)

func HandlerVolume(queue rmq.Queue, db *gorm.DB, cfg *configs.Config) {
	// maximum 5 worker or more
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("worker Volume consumer %d", i)
		if _, err := queue.AddConsumer(name, functions.NewVolumeReport(&functions.VolumeReport{
			WorkerIndex:   i,
			Config:        cfg,
			ProjectRepo:   postgresql.NewProjectAdapter(db),
			StructureRepo: postgresql.NewCompanyStructureAdapter(db),
			VolumeRepo:    postgresql.NewVolumeAdapter(db),
			LogProcess:    utils.NewJobQueueLog(db),
		})); err != nil {
			panic(err)
		}
	}
}
