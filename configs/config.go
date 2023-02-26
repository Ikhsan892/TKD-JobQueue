package configs

import (
	"assessment/pkg/utils"
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DB     DBConfig
	Redis  RedisConfig
	Report ReportConfig
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		if os.Getenv("ENV") == "development" {
			utils.Error(fmt.Sprintf("Error loading .env file : %s", err.Error()))
			panic(nil)
		}
		utils.Warn("Env Not Loaded, skip to Env OS")
	}

	return &Config{
		DB:     LoadDBConfig(),
		Redis:  LoadRedisConfig(),
		Report: LoadReportConfig(),
	}
}
