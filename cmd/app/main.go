package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jhawk7/go-irrigation/pkg/adcsensor"
	"github.com/jhawk7/go-irrigation/pkg/common"
	"github.com/jhawk7/go-irrigation/pkg/controller"
	"github.com/jhawk7/go-irrigation/pkg/pump"
	"github.com/jhawk7/go-opentel/opentel"

	rpio "github.com/stianeikeland/go-rpio/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	meterName         = "rpi-irrigation"
	waterPump1        = 24
	waterPump2        = 15
	idealMoisture     = float32(90)
	moistureThreshold = float32(22)
)

var plantController1 *controller.Controller
var plantController2 *controller.Controller
var checkDelay = common.GetenvInt("CHECK_DELAY_SECONDS")
var callbackDelay = common.GetenvInt("CALLBACK_DELAY_SECONDS")

func main() {
	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		rpioErr := fmt.Errorf("failed to map memory access to gpio pins; [error: %v]", err)
		common.ErrorHandler(rpioErr, true)
	}
	// Unmap gpio memory when done
	defer rpio.Close()

	// Init moisture sensor ADC via I2C device
	adcSensor, mErr := adcsensor.InitADCSensor()
	common.ErrorHandler(mErr, true)

	defer func() {
		busErr, haltErr := adcSensor.Close()
		common.ErrorHandler(busErr, true)
		common.ErrorHandler(haltErr, true)
	}()

	// Init water pumps
	pump1 := pump.InitWaterPumpRelay(waterPump1)
	pump2 := pump.InitWaterPumpRelay(waterPump2)

	// Init controllers
	plantController1 = &controller.Controller{
		Pump:                    pump1,
		ADCSensor:               adcSensor,
		IdealMoisturePercentage: idealMoisture,
		Threshold:               moistureThreshold,
		Name:                    "plant1",
		NeedsWater:              false,
		Channel:                 adcsensor.Channel0,
	}

	plantController2 = &controller.Controller{
		Pump:                    pump2,
		ADCSensor:               adcSensor,
		IdealMoisturePercentage: idealMoisture,
		Threshold:               moistureThreshold,
		Name:                    "plant2",
		NeedsWater:              false,
		Channel:                 adcsensor.Channel1,
	}

	// Init opentelemetry metric and trace providers
	if opentelErr := opentel.InitOpentelProviders; opentelErr() != nil {
		common.ErrorHandler(fmt.Errorf("failed to init opentel providers; [error: %v]", opentelErr()), true)
	}

	defer func() {
		shutdownErr := opentel.ShutdownOpentelProviders()
		common.ErrorHandler(shutdownErr, true)
	}()

	// Continuously update latest moisture reading
	// go plantController1.PollMoistureLv()
	// go plantController2.PollMoistureLv()
	go gaugeMoistureLevel()

	for {
		plantController1.CheckMoistureLv()
		time.Sleep(time.Second * time.Duration(checkDelay))
		plantController2.CheckMoistureLv()
		time.Sleep(time.Second * time.Duration(checkDelay))
	}
}

func gaugeMoistureLevel() {
	// creates meter and gauge observer from opentel meter provider
	moistureReader := opentel.GetMeterProvider().Meter(meterName)
	// gauge observer continuously polls data from callback
	moistureReader.NewFloat64GaugeObserver(fmt.Sprintf("%v.read", meterName), moistureCallback)
}

var moistureCallback = func(ctx context.Context, result metric.Float64ObserverResult) {
	common.LogInfo("observing moisture levels")
	//moistureReading1, moistureReading2 := readMoistureLevel()
	result.Observe(float64(plantController1.LatestReading), attribute.String("read.type", "percentage"), attribute.String("controller.name", plantController1.Name))
	result.Observe(float64(plantController2.LatestReading), attribute.String("read.type", "percentage"), attribute.String("controller.name", plantController2.Name))
	common.LogInfo(fmt.Sprintf("Plant1 Reading: %.2f%% Plant2 Reading: %.2f%%", float64(plantController1.LatestReading), float64(plantController2.LatestReading)))
	time.Sleep(time.Second * time.Duration(callbackDelay))
}

// func readMoistureLevel() (float64, float64) {
// 	moistureReading1 := float64(plantController1.PollMoistureLv())
// 	moistureReading2 := float64(plantController2.PollMoistureLv())
// 	log.Infof("Plant1 moisture reading: %v\nPlant2 moisture reading: %v", moistureReading1, moistureReading2)
// 	return moistureReading1, moistureReading2
// }
