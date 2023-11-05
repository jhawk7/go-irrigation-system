package pump

import (
	"time"

	rpio "github.com/stianeikeland/go-rpio/v4"
)

type WaterPumpRelay struct {
	pin rpio.Pin //GPIO pin (not physical pin number)
}

func InitWaterPumpRelay(pin int) *WaterPumpRelay {
	w := WaterPumpRelay{
		pin: rpio.Pin(pin),
	}
	w.pin.Output()
	return &w
}

// Songle 2-Channel Relay uses active low input pins (activates relay on low voltage)
func (w *WaterPumpRelay) Release() {
	w.pin.Low()
	time.Sleep(time.Second * 4)
	w.pin.High()
}
