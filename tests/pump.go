// package main

// import (
// 	"fmt"
// 	log "github.com/sirupsen/logrus"
// 	rpio "github.com/stianeikeland/go-rpio/v4"
// 	"time"
// )

// type WaterPumpRelay struct {
// 	pin rpio.Pin //GPIO pin (not physical pin number)
// }

// func InitWaterPumpRelay(pin int) *WaterPumpRelay {
// 	w := WaterPumpRelay{
// 		pin: rpio.Pin(pin),
// 	}
// 	w.pin.Output()
// 	return &w
// }

// // Songle 2-Channel Relay uses active low input pins (activates relay on low voltage)
// func (w *WaterPumpRelay) Release() {
// 	w.pin.Low()
// 	time.Sleep(time.Second * 3)
// 	w.pin.High()
// }

// func main() {
// 	// Open and map memory to access gpio, check for errors
// 	if err := rpio.Open(); err != nil {
// 		rpioErr := fmt.Errorf("failed to map memory access to gpio pins; [error: %v]", err)
// 		log.Fatal(rpioErr)
// 	}
// 	// Unmap gpio memory when done
// 	defer rpio.Close()

// 	// Init water pumps
// 	pump1 := InitWaterPumpRelay(24)
// 	pump2 := InitWaterPumpRelay(15)

// 	pump1.Release()
// 	time.Sleep(time.Second)
// 	pump2.Release()
// }
