package handlers

import (
	"assessment/pkg/functions"
	"github.com/adjust/rmq/v5"
)

func HandlerTest(queue rmq.Queue) {
	// maximum 5 worker or more
	for i := 0; i < 5; i++ {
		if _, err := queue.AddConsumer("test", functions.NewTest(i)); err != nil {
			panic(err)
		}
	}
}
