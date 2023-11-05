package controller

import (
	"fmt"
	"time"

	"github.com/jhawk7/go-irrigation/pkg/adcsensor"
	"github.com/jhawk7/go-irrigation/pkg/common"
	"github.com/jhawk7/go-irrigation/pkg/pump"
	"periph.io/x/devices/v3/ads1x15"
)

type Controller struct {
	Pump                    *pump.WaterPumpRelay
	ADCSensor               *adcsensor.ADCSensor
	IdealMoisturePercentage float32
	Threshold               float32
	Name                    string
	NeedsWater              bool
	LatestReading           float32
	Channel                 ads1x15.Channel
}

func (c *Controller) CheckMoistureLv() {
	reading, readErr := c.ADCSensor.ReadMoistureValue(c.Channel)
	common.ErrorHandler(readErr, false)

	c.LatestReading = reading
	if c.LatestReading <= float32(c.Threshold) {
		c.NeedsWater = true
		common.LogInfo(fmt.Sprintf("%v needs water [reading: %.2f%%]", c.Name, c.LatestReading))
		c.pumpWater()
	}

	common.LogInfo(fmt.Sprintf("%v latest reading: %.2f%%", c.Name, c.LatestReading))
}

func (c *Controller) pumpWater() {
	for c.LatestReading < float32(c.IdealMoisturePercentage) {
		common.LogInfo(fmt.Sprintf("pumping water for %v", c.Name))
		c.Pump.Release()
		time.Sleep(5 * time.Second)

		reading, readErr := c.ADCSensor.ReadMoistureValue(c.Channel)
		common.ErrorHandler(readErr, true)
		c.LatestReading = reading
		common.LogInfo(fmt.Sprintf("%v [current: %v] [ideal: %v]", c.Name, c.LatestReading, c.IdealMoisturePercentage))
	}
	c.NeedsWater = false
}

// This function will be called repeatedly in a go routine
// func (c *Controller) PollMoistureLv() float32 {
// 	latestReading, readErr := c.ADCSensor.ReadMoistureValue(c.Channel)
// 	common.ErrorHandler(readErr, true)
// 	c.LatestReading = latestReading
// 	return latestReading
// }

func (c *Controller) PollMoistureLv() {
	ch := make(chan float32)
	c.ADCSensor.PollMoistureValue(c.Channel, ch)
	for reading := range ch {
		c.LatestReading = reading
	}
}
