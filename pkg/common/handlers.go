package common

import (
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
