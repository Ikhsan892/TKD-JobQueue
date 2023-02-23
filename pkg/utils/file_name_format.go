package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func FileNameFormat(outputLocation, processName string) string {
	return formatFileName(outputLocation, processName, "xlsx")
}

func formatFileName(outputLoc, processName, format string) string {
	rand.Seed(time.Now().UTC().UnixNano())
	randomInt := rand.Int()
	return fmt.Sprintf("%s/%s%d.%s", outputLoc, processName, randomInt, format)
}
