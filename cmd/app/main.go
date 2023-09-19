package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jhawk7/go-opentel/opentel"
	"github.com/jhawk7/go-pi-irrigation/pkg/common"
	"github.com/jhawk7/go-pi-irrigation/pkg/controller"
	"github.com/jhawk7/go-pi-irrigation/pkg/moisture_sensor"
	"github.com/jhawk7/go-pi-irrigation/pkg/water_pump"

	log "github.com/sirupsen/logrus"
	rpio "github.com/stianeikeland/go-rpio/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var METER_NAME string = "pi-irrigation"
var plantController1 *controller.Controller
var plantController2 *controller.Controller

func initTelemetry() {
	if otlpErr := opentel.InitOpentelProviders; otlpErr != nil {
		common.ErrorHandler(fmt.Errorf("failed to init otlp providers; [error: %v]", otlpErr), true)
	}
}

func main() {
	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		rpioErr := fmt.Errorf("failed to map memory access to gpio pins; [error: %v]", err)
		common.ErrorHandler(rpioErr, true)
	}
	// Unmap gpio memory when done
	defer rpio.Close()

	// Init moisture sensor ADC via I2C device
	mSensor, mErr := moisture_sensor.InitMoistureSensor()
	common.ErrorHandler(mErr, true)

	defer func() {
		err := mSensor.Close()
		common.ErrorHandler(err, true)
	}()

	// Init water pumps
	pump1 := water_pump.InitWaterPumpRelay(24)
	pump2 := water_pump.InitWaterPumpRelay(15)

	// Init controllers
	plantController1 = &controller.Controller{
		Pump:                    pump1,
		Msensor:                 mSensor,
		IdealMoisturePercentage: 90,
		Threshold:               22,
		Name:                    "plant1",
		NeedsWater:              false,
		Channel:                 moisture_sensor.Channel0,
	}

	plantController2 = &controller.Controller{
		Pump:                    pump2,
		Msensor:                 mSensor,
		IdealMoisturePercentage: 90,
		Threshold:               22,
		Name:                    "plant2",
		NeedsWater:              false,
		Channel:                 moisture_sensor.Channel1,
	}

	// Init opentelemetry metric and trace providers
	initTelemetry()
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
		plantController2.CheckMoistureLv()
		time.Sleep(time.Hour)
	}
}

func gaugeMoistureLevel() {
	// creates meter and gauge observer from opentel meter provider
	moistureReader := opentel.GetMeterProvider().Meter(METER_NAME)
	// gauge observer continuously polls data from callback
	moistureReader.NewFloat64GaugeObserver(fmt.Sprintf("%v.read", METER_NAME), moistureCallback)
}

var moistureCallback = func(ctx context.Context, result metric.Float64ObserverResult) {
	log.Info("observing moisture levels")
	//moistureReading1, moistureReading2 := readMoistureLevel()
	result.Observe(float64(plantController1.LatestReading), attribute.String("read.type", "percentage"), attribute.String("controller.name", plantController1.Name))
	result.Observe(float64(plantController2.LatestReading), attribute.String("read.type", "percentage"), attribute.String("controller.name", plantController2.Name))
	time.Sleep(time.Hour)
}

// func readMoistureLevel() (float64, float64) {
// 	moistureReading1 := float64(plantController1.PollMoistureLv())
// 	moistureReading2 := float64(plantController2.PollMoistureLv())
// 	log.Infof("Plant1 moisture reading: %v\nPlant2 moisture reading: %v", moistureReading1, moistureReading2)
// 	return moistureReading1, moistureReading2
// }
