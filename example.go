package golightmyroom

import (
	"context"
	"sync"
)

func main() {
	lights, err := multipleLights("mac address 1", "mac address 2")
	if err != nil {
		panic(err)
	}

	runMultiple(lights, func(light Light) {
		light.SetBrightness(0.5)
		if temperatureControlled, ok := light.(TemperatureControl); ok {
			temperatureControlled.SetTemperature(TemperatureWarmWhite)
		}
	})
}

func multipleLights(macAddresses ...string) (lights []Light, err error) {
	for _, macAddress := range macAddresses {
		var light Light
		light, err = ConnectToBluetoothWithContext(macAddress, context.Background())
		if err != nil {
			return nil, err
		}
		lights = append(lights, light)
	}
	return lights, nil
}

func runMultiple(lights []Light, f func(light Light)) {
	var wg sync.WaitGroup
	for _, light := range lights {
		wg.Add(1)

		light := light
		go func() {
			f(light)
			wg.Done()
		}()
	}
	wg.Wait()
}
