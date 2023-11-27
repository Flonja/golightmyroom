package models

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"math"
	"tinygo.org/x/bluetooth"
)

type CenturaColorSpot struct {
	device   *bluetooth.Device
	services [][]bluetooth.DeviceCharacteristic
	gamut    *Gamut
}

func NewCenturaColorSpot(device *bluetooth.Device) (*CenturaColorSpot, error) {
	services, err := device.DiscoverServices(nil)
	if err != nil {
		return nil, err
	}
	var mappedServices [][]bluetooth.DeviceCharacteristic
	for _, service := range services {
		discoveredCharacteristics, err := service.DiscoverCharacteristics(nil)
		if err != nil {
			return nil, err
		}

		mappedServices = append(mappedServices, discoveredCharacteristics)
	}

	return &CenturaColorSpot{device: device, services: mappedServices}, nil
}

func (c *CenturaColorSpot) findCharacteristic(characteristic string) *bluetooth.DeviceCharacteristic {
	for _, service := range c.services {
		for _, deviceCharacteristic := range service {
			if deviceCharacteristic.String() == characteristic {
				return &deviceCharacteristic
			}
		}
	}
	return nil
}

func (c *CenturaColorSpot) mustReadCharacteristics(characteristic string) (buf []byte) {
	char := c.findCharacteristic(characteristic)

	maxTries := 5
	for maxTries > 0 {
		buf = make([]byte, 512)
		n, err := char.Read(buf)
		if err != nil {
			fmt.Printf("couldn't read from light: %v\n", err)
			maxTries--
			continue
		}
		return buf[:n]
	}
	return nil
}

var errTooLong = errors.New("buffer too long")

func (c *CenturaColorSpot) mustWriteCharacteristics(characteristic string, buf []byte) {
	if len(buf) > 512 {
		panic(errTooLong)
	}
	char := c.findCharacteristic(characteristic)

	maxTries := 5
	for maxTries > 0 {
		if _, err := char.WriteWithoutResponse(buf); err != nil {
			fmt.Printf("couldn't write to light: %v\n", err)
			maxTries--
			continue
		}
		break
	}
}

func (c *CenturaColorSpot) On() {
	c.mustWriteCharacteristics("932c32bd-0002-47a2-835a-a8d455b859dd", []byte{1})
}

func (c *CenturaColorSpot) Off() {
	c.mustWriteCharacteristics("932c32bd-0002-47a2-835a-a8d455b859dd", []byte{0})
}

func (c *CenturaColorSpot) Powered() bool {
	return c.mustReadCharacteristics("932c32bd-0002-47a2-835a-a8d455b859dd")[0] == 1
}

func (c *CenturaColorSpot) Brightness() float64 {
	return float64(c.mustReadCharacteristics("932c32bd-0003-47a2-835a-a8d455b859dd")[0]) / 254
}

func (c *CenturaColorSpot) SetBrightness(b float64) {
	c.mustWriteCharacteristics("932c32bd-0003-47a2-835a-a8d455b859dd", []byte{byte(min(max(b*254, 1), 254))})
}

func (c *CenturaColorSpot) Model() string {
	return string(c.mustReadCharacteristics("00002a24-0000-1000-8000-00805f9b34fb"))
}

func (c *CenturaColorSpot) Temperature() uint16 {
	return uint16(math.Round(1e6 / float64(binary.LittleEndian.Uint16(c.mustReadCharacteristics("932c32bd-0004-47a2-835a-a8d455b859dd")))))
}

func (c *CenturaColorSpot) SetTemperature(temperature uint16) {
	rawTemp := max(153, min(uint16(math.Round(1e6/float64(temperature))), 500))
	c.mustWriteCharacteristics("932c32bd-0004-47a2-835a-a8d455b859dd", []byte{byte(rawTemp), byte(rawTemp >> 8)})
}

func (c *CenturaColorSpot) getGamut() *Gamut {
	if c.gamut == nil {
		gamut := GetGamutForModel(c.Model())
		c.gamut = &gamut
	}
	return c.gamut
}

func (c *CenturaColorSpot) Color() color.Color {
	bits := c.mustReadCharacteristics("932c32bd-0005-47a2-835a-a8d455b859dd")
	x := float64(binary.LittleEndian.Uint16(bits[0:2])) / 0xffff
	y := float64(binary.LittleEndian.Uint16(bits[2:4])) / 0xffff
	return c.getGamut().XYYToColor(XYPoint{x, y}, float64(c.Brightness())/255)
}

func (c *CenturaColorSpot) SetColor(colour color.Color) {
	xy := c.getGamut().ColorToXY(colour)
	xBits := uint16(xy[0] * 0xffff)
	yBits := uint16(xy[1] * 0xffff)
	c.mustWriteCharacteristics("932c32bd-0005-47a2-835a-a8d455b859dd", []byte{byte(xBits), byte(xBits >> 8), byte(yBits), byte(yBits >> 8)})
}

func (c *CenturaColorSpot) Name() string {
	return string(c.mustReadCharacteristics("97fe6561-0003-4f62-86e9-b71ee2da3d22"))
}

func (c *CenturaColorSpot) SetName(name string) {
	c.mustWriteCharacteristics("97fe6561-0003-4f62-86e9-b71ee2da3d22", []byte(name))
}

func (c *CenturaColorSpot) bluetoothDevice() *bluetooth.Device {
	return c.device
}

func (c *CenturaColorSpot) Disconnect() error {
	return c.device.Disconnect()
}
