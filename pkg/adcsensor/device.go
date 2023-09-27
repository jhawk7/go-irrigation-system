package adcsensor

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
	waterVoltage float32 = 5868  //voltage reading of sensor in water
	airVoltage   float32 = 13560 //voltage reading of sensor in air

	Channel0 ads1x15.Channel = ads1x15.Channel0
	Channel1 ads1x15.Channel = ads1x15.Channel1
	Channel2 ads1x15.Channel = ads1x15.Channel2
	Channel3 ads1x15.Channel = ads1x15.Channel3
)

type ADCSensor struct {
	bus   i2c.BusCloser
	ADC   *ads1x15.Dev
	freq  physic.Frequency
	delay time.Duration
	cache map[string]ads1x15.PinADC //channel-freq string to pin
}

func InitADCSensor() (sensor *ADCSensor, err error) {
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

	sensor = &ADCSensor{
		bus:   bus,
		cache: make(map[string]ads1x15.PinADC),
		ADC:   adc,
		freq:  physic.Frequency(common.GetenvInt("ADC_SAMPLE_FREQ")),
		delay: time.Duration(common.GetenvInt("CHECK_DELAY_SECONDS")),
	}
	return
}

func (sensor *ADCSensor) Close() (busErr error, adcErr error) {
	if closeErr := sensor.bus.Close(); closeErr != nil {
		busErr = fmt.Errorf("failed to properly close bus: %v", closeErr)
	}

	if haltErr := sensor.ADC.Halt(); haltErr != nil {
		adcErr = fmt.Errorf("failed to halt adc: %v", haltErr)
	}
	return
}

/*
. ADS1115 provides 4 channels to read values from
. The max freq for a pin is 860Hz or .86ksps per https://www.ti.com/product/ADS1115
. Misreads/duplicate reads can occur if function is called faster than the sampling set sampling rate of pin
*/
func (sensor *ADCSensor) ReadMoistureValue(channel ads1x15.Channel) (moisturePercentage float32, err error) {
	pin, pinErr := sensor.getPin(channel)
	if pinErr != nil {
		err = pinErr
		return
	}
	//defer pin.Halt() // doesn't close pin

	common.LogInfo(fmt.Sprintf("reading channel %v", channel))
	readSample, readErr := pin.Read()
	if readErr != nil {
		err = fmt.Errorf("failed to get reading from pin for channel %v; %v", channel, readErr)
		return
	}
	rawReading := readSample.Raw
	moisturePercentage = mapReading(rawReading)
	time.Sleep(time.Second * sensor.delay) //wait time between channel reads
	return
}

func (sensor *ADCSensor) getPin(channel ads1x15.Channel) (pin ads1x15.PinADC, err error) {
	key := fmt.Sprintf("%v-%v", channel, sensor.freq)
	if p, pinExists := sensor.cache[key]; pinExists {
		pin = p
	} else {
		p, pinErr := sensor.ADC.PinForChannel(channel, 5*physic.Volt, sensor.freq, ads1x15.BestQuality)
		if pinErr != nil {
			err = fmt.Errorf("failed create pin for channel; %v", pinErr)
			return
		}
		sensor.cache[key] = p
		common.LogInfo(fmt.Sprintf("caching mSensor pin for key: %v", key))
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

/*
Polling the channel continously will only work reliably for continously reading 1/4 pins.
The ADC needs time to switch channels and process voltage.
*/
func (sensor *ADCSensor) PollMoistureValue(channel ads1x15.Channel, readingCh chan float32) {
	pin, pinErr := sensor.getPin(channel)
	if pinErr != nil {
		common.ErrorHandler(pinErr, true)
	}
	defer pin.Halt() // doesn't close pin

	analogCh := pin.ReadContinuous()
	for analogRead := range analogCh {
		readingCh <- mapReading(analogRead.Raw)
	}
}
