package utils

import "os"

func GetWorkingDirectoryContent(filePath string) string {
	dir := os.Getenv("APP_DIRECTORY")
	return dir + filePath
}

func GetTargetZipLocation() string {
	return os.Getenv("TARGET_DIRECTORY")
}
