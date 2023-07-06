package moisture_sensor

import (
	"fmt"
	"time"

	"github.com/jhawk7/go-pi-irrigation/pkg/common"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/ads1x15"
	"periph.io/x/host/v3"
)

/*
* Create new connection to I2C bus on line 1 with address 0x49
* run `i2cdetect -y 1` to view vtable for specific device addr
* when loaded, a specific device entry folder /dev/i2c-* will be created; using bus 1 for /dev/i2c-1
 */

const (
	i2cBus       string  = "1"   // /i2c/dev/1 channel of ADC connected via i2c
	i2cAddr      uint16  = 0x49  // I2C address of the ADC device
	airVoltage   float32 = 5868  //voltage reading of sensor in air
	waterVoltage float32 = 13560 //voltage reading of sensor in water

	Channel0 ads1x15.Channel = ads1x15.Channel0
	Channel1 ads1x15.Channel = ads1x15.Channel1
	Channel2 ads1x15.Channel = ads1x15.Channel2
	Channel3 ads1x15.Channel = ads1x15.Channel3
)

type ADCMoistureSensor struct {
	bus   i2c.BusCloser
	ADC   *ads1x15.Dev
	cache map[string]ads1x15.PinADC //channel-freq string to pin
}

func InitMoistureSensor() (moistureSensor *ADCMoistureSensor, err error) {
	// Make sure periph is initialized.
	if _, initErr := host.Init(); err != nil {
		err = fmt.Errorf("failed to init periph pkg; %v", initErr)
		return
	}

	// Open default I²C bus.
	bus, busErr := i2creg.Open(i2cBus)
	if busErr != nil {
		err = fmt.Errorf("failed to open i2c bus; %v", busErr)
		return
	}

	// Create a new ADS1115 ADC.
	adc, adcErr := ads1x15.NewADS1115(bus, &ads1x15.Opts{I2cAddress: i2cAddr})
	if adcErr != nil {
		err = fmt.Errorf("failed to create new ADC; %v", adcErr)
		return
	}

	moistureSensor = &ADCMoistureSensor{
		bus:   bus,
		cache: make(map[string]ads1x15.PinADC),
		ADC:   adc,
	}
	return
}

func (moistureSensor *ADCMoistureSensor) Close() (err error) {
	if closeErr := moistureSensor.bus.Close(); closeErr != nil {
		err = fmt.Errorf("failed to properly close bus: %v", closeErr)
		return
	}
	return
}

// ADS1115 provides 4 channels to read values from
func (moistureSensor *ADCMoistureSensor) ReadMoistureValue(channel ads1x15.Channel) (moisturePercentage float32, err error) {
	pin, pinErr := moistureSensor.getPin(channel, 1*physic.Hertz)
	if pinErr != nil {
		err = pinErr
		return
	}
	defer pin.Halt() // doesn't close pin

	readSample, readErr := pin.Read()
	if readErr != nil {
		err = fmt.Errorf("failed to get reading from pin; %v", readErr)
		return
	}
	rawReading := readSample.Raw
	moisturePercentage = mapReading(rawReading)
	time.Sleep(time.Millisecond * 500) //wait time between channel reads
	return
}

func (moistureSensor *ADCMoistureSensor) PollMoistureValue(channel ads1x15.Channel, readingCh chan float32) {
	pin, pinErr := moistureSensor.getPin(channel, 10*physic.MilliHertz)
	if pinErr != nil {
		common.ErrorHandler(pinErr, true)
	}
	defer pin.Halt() // doesn't close pin

	analogCh := pin.ReadContinuous()
	for analogRead := range analogCh {
		readingCh <- mapReading(analogRead.Raw)
	}
	return
}

func (moistureSensor *ADCMoistureSensor) getPin(channel ads1x15.Channel, freq physic.Frequency) (pin ads1x15.PinADC, err error) {
	key := fmt.Sprintf("%v-%v", channel, freq)
	if p, pinExists := moistureSensor.cache[key]; pinExists {
		pin = p
	} else {
		p, pinErr := moistureSensor.ADC.PinForChannel(channel, 5*physic.Volt, freq, ads1x15.SaveEnergy)
		if pinErr != nil {
			err = fmt.Errorf("failed create pin for channel; %v", pinErr)
			return
		}
		moistureSensor.cache[key] = p
		pin = p
	}
	return
}

/*
		maps reading to a range specified by the min and max (inspired by map arduino func)
	  formula: (x - in_min) * (out_max - out_min) / (in_max - in_min) + out_min;
*/
func mapReading(rawReading int32) float32 {
	moisturePercentage := (float32(rawReading)-airVoltage)*(100-0)/(waterVoltage-airVoltage) + 0
	return moisturePercentage
}
