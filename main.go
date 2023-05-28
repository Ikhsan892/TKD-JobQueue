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
	questioner, err := conn.OpenQueue("report_questioner")
	volumeAttachment, err := conn.OpenQueue("volume_attachment")
	if err != nil {
		panic(err)
	}

	// start consuming
	if errConsuming := queue.StartConsuming(prefetchLimit, pollDuration); errConsuming != nil {
		panic(errConsuming)
	}
	if errConsumingQuestioner := questioner.StartConsuming(prefetchLimit, pollDuration); errConsumingQuestioner != nil {
		panic(errConsumingQuestioner)
	}
	if errConsumingVolume := volumeAttachment.StartConsuming(prefetchLimit, pollDuration); errConsumingVolume != nil {
		panic(errConsumingVolume)
	}

	// assign handlers
	go handlers.HandlerTest(queue)
	go handlers.HandlerQuestioner(questioner, db, cfg)
	go handlers.HandlerVolumeAttachment(volumeAttachment, db, cfg)

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-sigQuit

	utils.Info(utils.DATABASE, "Closing Connection")

	<-conn.StopAllConsuming()
}
