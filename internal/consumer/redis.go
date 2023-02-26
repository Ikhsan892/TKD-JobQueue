package consumer

import (
	"assessment/configs"
	"assessment/pkg/utils"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
)

func ErrConnectRedis() {
	if msg := recover(); msg != nil {
		log.Fatal(msg)
		os.Exit(1)
	}
}

func ConnectRedis(cfg configs.RedisConfig) *redis.Client {
	defer ErrConnectRedis()

	var rdb *redis.Client

	opts, err := redis.ParseURL(fmt.Sprintf("redis://%s:%s@%s:%d/%d",
		cfg.RedisUsername,
		cfg.RedisPassword,
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisDatabase,
	))

	if err != nil {
		panic(err.Error())
	}

	rdb = redis.NewClient(opts)

	ctx := context.Background()
	if errConn := rdb.Ping(ctx).Err(); errConn != nil {
		utils.Error(utils.REDIS, errConn)
		os.Exit(1)
	} else {
		utils.Info(utils.REDIS, "Redis Connected")
	}

	return rdb
}
