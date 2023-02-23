package handlers

import (
	"assessment/configs"
	"assessment/pkg/adapter/postgresql"
	"assessment/pkg/functions"
	"fmt"
	"github.com/adjust/rmq/v5"
	"gorm.io/gorm"
)

func HandlerQuestioner(queue rmq.Queue, db *gorm.DB, cfg *configs.Config) {
	// maximum 5 worker or more
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("worker Questioner consumer %d", i)
		if _, err := queue.AddConsumer(name, functions.NewQuestioner(&functions.Questioner{
			WorkerIndex:    i,
			Config:         cfg,
			QuestionerRepo: postgresql.NewQuestionerAdapter(db),
			LogProcess:     nil,
		})); err != nil {
			panic(err)
		}
	}
}
