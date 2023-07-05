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

func HandlerInventory(queue rmq.Queue, db *gorm.DB, cfg *configs.Config) {
	// maximum 5 worker or more
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("worker Inventory consumer %d", i)
		if _, err := queue.AddConsumer(name, functions.NewInventory(&functions.Inventory{
			WorkerIndex:   i,
			Config:        cfg,
			ProjectRepo:   postgresql.NewProjectAdapter(db),
			StructureRepo: postgresql.NewCompanyStructureAdapter(db),
			InventoryRepo: postgresql.NewQuestionerAdapter(db),
			LogProcess:    utils.NewJobQueueLog(db),
		})); err != nil {
			panic(err)
		}
	}
}
