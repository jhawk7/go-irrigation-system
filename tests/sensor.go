// package main

// import (
// 	"fmt"
// 	"time"

// 	"periph.io/x/conn/v3/i2c"
// 	"periph.io/x/conn/v3/i2c/i2creg"
// 	"periph.io/x/conn/v3/physic"
// 	"periph.io/x/devices/v3/ads1x15"
// 	"periph.io/x/host/v3"

// 	log "github.com/sirupsen/logrus"
// )

// /*
// * Create new connection to I2C bus on line 1 with address 0x49
// * run `i2cdetect -y 1` to view vtable for specific device addr
// * when loaded, a specific device entry folder /dev/i2c-* will be created; using bus 1 for /dev/i2c-1
//  */

// const (
// 	i2cBus       string  = "1"   // /i2c/dev/1 channel of ADC connected via i2c
// 	i2cAddr      uint16  = 0x49  // I2C address of the ADC device
// 	waterVoltage float32 = 5868  //voltage reading of sensor in water
// 	airVoltage   float32 = 13560 //voltage reading of sensor in air

// 	Channel0 ads1x15.Channel = ads1x15.Channel0
// 	Channel1 ads1x15.Channel = ads1x15.Channel1
// 	Channel2 ads1x15.Channel = ads1x15.Channel2
// 	Channel3 ads1x15.Channel = ads1x15.Channel3
// )

// type ADCMoistureSensor struct {
// 	bus   i2c.BusCloser
// 	cache map[string]ads1x15.PinADC
// 	ADC   *ads1x15.Dev
// }

// func InitMoistureSensor() (moistureSensor *ADCMoistureSensor, err error) {
// 	// Make sure periph is initialized.
// 	if _, initErr := host.Init(); initErr != nil {
// 		err = fmt.Errorf("failed to init periph pkg; %v", initErr)
// 		return
// 	}

// 	// Open default I²C bus.
// 	bus, busErr := i2creg.Open(i2cBus)
// 	if busErr != nil {
// 		err = fmt.Errorf("failed to open i2c bus; %v", busErr)
// 		return
// 	}

// 	// Create a new ADS1115 ADC.
// 	adc, adcErr := ads1x15.NewADS1115(bus, &ads1x15.Opts{I2cAddress: i2cAddr})
// 	if adcErr != nil {
// 		err = fmt.Errorf("failed to create new ADC; %v", adcErr)
// 		return
// 	}

// 	moistureSensor = &ADCMoistureSensor{
// 		bus:   bus,
// 		cache: make(map[string]ads1x15.PinADC),
// 		ADC:   adc,
// 	}

// 	return
// }

// func (moistureSensor *ADCMoistureSensor) Close() (err error) {
// 	if closeErr := moistureSensor.bus.Close(); closeErr != nil {
// 		err = fmt.Errorf("failed to properly close bus: %v", closeErr)
// 		return
// 	}
// 	return
// }

// // ADS1115 provides 4 channels to read values from
// func (moistureSensor *ADCMoistureSensor) ReadMoistureValue(channel ads1x15.Channel) (rawReading int32, moisturePercentage float32, err error) {
// 	pin, pinErr := moistureSensor.getPin(channel, 50*physic.Hertz)
// 	if pinErr != nil {
// 		err = pinErr
// 		return
// 	}
// 	defer pin.Halt() // doesn't close pin

// 	readSample, readErr := pin.Read()
// 	if readErr != nil {
// 		err = fmt.Errorf("failed to get reading from pin; %v", readErr)
// 		return
// 	}
// 	rawReading = readSample.Raw
// 	moisturePercentage = mapReading(rawReading)
// 	time.Sleep(time.Millisecond * 500)
// 	return
// }

// func (moistureSensor *ADCMoistureSensor) getPin(channel ads1x15.Channel, freq physic.Frequency) (pin ads1x15.PinADC, err error) {
// 	key := fmt.Sprintf("%v-%v", channel, freq)
// 	if p, pinExists := moistureSensor.cache[key]; pinExists {
// 		pin = p
// 	} else {
// 		p, pinErr := moistureSensor.ADC.PinForChannel(channel, 5*physic.Volt, freq, ads1x15.SaveEnergy)
// 		if pinErr != nil {
// 			err = fmt.Errorf("failed create pin for channel; %v", pinErr)
// 			return
// 		}
// 		moistureSensor.cache[key] = p
// 		pin = p
// 	}
// 	return
// }

// /*
// 		maps reading to a range specified by the min and max (inspired by map arduino func)
// 	  formula: (x - in_min) * (out_max - out_min) / (in_max - in_min) + out_min;
// */
// func mapReading(rawReading int32) float32 {
// 	moisturePercentage := (float32(rawReading)-airVoltage)*(100-0)/(waterVoltage-airVoltage) + 0
// 	return moisturePercentage
// }

// /*
// Reliable for continoulsy reading 1/4 pins
// */
// func (moistureSensor *ADCMoistureSensor) PollMoistureValue(channel ads1x15.Channel, readingCh chan float32) {
// 	pin, pinErr := moistureSensor.getPin(channel, 2*physic.Hertz)
// 	if pinErr != nil {
// 		log.Fatal(pinErr)
// 	}
// 	defer pin.Halt() // doesn't close pin

// 	analogCh := pin.ReadContinuous()
// 	for analogRead := range analogCh {
// 		readingCh <- mapReading(analogRead.Raw)
// 	}
// 	return
// }

// func main() {
// 	// Init moisture sensor ADC via I2C device
// 	mSensor, mErr := InitMoistureSensor()
// 	if mErr != nil {
// 		log.Fatal(fmt.Errorf("failed to init sensor %v", mErr))
// 	}

// 	defer func() {
// 		if err := mSensor.Close(); err != nil {
// 			log.Fatal(err)
// 		}
// 	}()

// 	for i := 0; i < 10; i++ {
// 		raw0, reading0, rErr0 := mSensor.ReadMoistureValue(Channel0)
// 		if rErr0 != nil {
// 			log.Fatal(rErr0)
// 		}
// 		fmt.Printf("Reading from channel 0 [raw: %.2d] [reading: %.2f]\n", raw0, reading0)

// 		raw1, reading1, rErr1 := mSensor.ReadMoistureValue(Channel1)
// 		if rErr1 != nil {
// 			log.Fatal(rErr1)
// 		}
// 		fmt.Printf("Reading from channel 1 [raw: %.2d] [reading: %.2f]\n", raw1, reading1)
// 	}

// 	fmt.Println("polling continously..")
// 	ch0 := make(chan float32)
// 	//ch1 := make(chan float32)
// 	go mSensor.PollMoistureValue(Channel0, ch0)
// 	//go mSensor.PollMoistureValue(Channel1, ch1)

// 	for read := range ch0 {
// 		fmt.Printf("Polled reading from ch0: %v\n", read)
// 		time.Sleep(time.Second)
// 	}
// }
