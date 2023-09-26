package common

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func ErrorHandler(err error, fatal bool) {
	if err != nil {
		log.Errorf("error: %v", err)

		if fatal {
			panic(err)
		}
	}
}

func LogInfo(info string) {
	log.Info(info)
}

func GetenvInt(key string) int {
	strVal := os.Getenv(key)
	if strVal == "" {
		err := fmt.Errorf("no value found for key %v", key)
		ErrorHandler(err, true)
	}

	val, convErr := strconv.Atoi(strVal)
	if convErr != nil {
		err := fmt.Errorf("failed to convert string value %v to int [error: %v]", strVal, convErr)
		ErrorHandler(err, true)
	}

	return val
}
