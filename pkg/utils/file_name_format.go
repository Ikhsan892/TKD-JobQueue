package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func FormatFilePath(outputLocation, fileName string) (string, string) {
	return formatFilePath(outputLocation, fileName, "xlsx")
}

func formatFileName(prefix, format string) string {
	rand.Seed(time.Now().UTC().UnixNano())
	randomInt := rand.Int()
	return fmt.Sprintf("%s%d.%s", prefix, randomInt, format)
}

func formatFilePath(outputLoc, processName, format string) (string, string) {
	fileName := formatFileName(processName, format)
	filePath := fmt.Sprintf("%s/%s", outputLoc, fileName)
	return filePath, fileName
}
