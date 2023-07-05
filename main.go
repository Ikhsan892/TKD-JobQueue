package main

import (
	"assessment/configs"
	"assessment/internal/consumer"
	"assessment/pkg/handlers"
	"assessment/pkg/utils"
	"github.com/adjust/rmq/v5"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	prefetchLimit = 100
	pollDuration  = 100 * time.Millisecond
	numConsumers  = 5

	reportBatchSize = 10000
	consumeDuration = time.Millisecond
	shouldLog       = true
)

func ConsumingQueues(queues ...rmq.Queue) error {
	for _, queue := range queues {
		if err := queue.StartConsuming(prefetchLimit, pollDuration); err != nil {
			return err
		}
	}
	return nil
}

func logErrors(errChan <-chan error) {
	for err := range errChan {
		switch errT := err.(type) {
		case *rmq.HeartbeatError:
			if errT.Count == rmq.HeartbeatErrorLimit {
				log.Print("heartbeat error (limit): ", errT)
			} else {
				log.Print("heartbeat error: ", errT)
			}
		case *rmq.ConsumeError:
			log.Print("consume error: ", errT)
		case *rmq.DeliveryError:
			log.Print("delivery error: ", errT.Delivery, err)
		default:
			log.Print("other error: ", errT)
		}
	}
}

func main() {
	cfg := configs.NewConfig()
	db := consumer.ConnectionDB(cfg.DB)

	errChan := make(chan error)

	go logErrors(errChan)
	conn, errConnection := rmq.OpenConnectionWithRedisClient("assessment job queues", consumer.ConnectRedis(cfg.Redis), errChan)
	if errConnection != nil {
		utils.Error(utils.DATABASE, errConnection)
		os.Exit(1)
	}

	// register new queue
	queue, err := conn.OpenQueue("test")
	volume, err := conn.OpenQueue("volume")
	inventory, err := conn.OpenQueue("inventory")
	questioner, err := conn.OpenQueue("report_questioner")
	volumeAttachment, err := conn.OpenQueue("volume_attachment")
	if err != nil {
		panic(err)
	}

	// start consuming
	if errConsuming := ConsumingQueues(queue, questioner, volumeAttachment, volume, inventory); errConsuming != nil {
		panic(errConsuming)
	}

	// assign handlers
	go handlers.HandlerTest(queue)
	go handlers.HandlerQuestioner(questioner, db, cfg)
	go handlers.HandlerInventory(inventory, db, cfg)
	go handlers.HandlerVolumeAttachment(volumeAttachment, db, cfg)
	go handlers.HandlerVolume(volume, db, cfg)

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-sigQuit

	utils.Info(utils.DATABASE, "Closing Connection")

	<-conn.StopAllConsuming()
}
