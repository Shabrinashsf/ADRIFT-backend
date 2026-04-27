package utils

import "os"

func GetStage() bool {
	if os.Getenv("IS_PRODUCTION") == "false" {
		return false
	}
	return true
}
