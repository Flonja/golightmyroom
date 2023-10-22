package models

import (
	"encoding/binary"
	"errors"
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

func (c *CenturaColorSpot) mustReadCharacteristics(characteristic string) []byte {
	buf := make([]byte, 512)
	char := c.findCharacteristic(characteristic)
	n, err := char.Read(buf)
	if err != nil {
		panic(err)
	}
	return buf[:n]
}

var errTooLong = errors.New("buffer too long")

func (c *CenturaColorSpot) mustWriteCharacteristics(characteristic string, buf []byte) {
	if len(buf) > 512 {
		panic(errTooLong)
	}
	char := c.findCharacteristic(characteristic)
	if _, err := char.WriteWithoutResponse(buf); err != nil {
		panic(err)
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

func (c *CenturaColorSpot) Brightness() byte {
	return c.mustReadCharacteristics("932c32bd-0003-47a2-835a-a8d455b859dd")[0]
}

func (c *CenturaColorSpot) SetBrightness(b byte) {
	c.mustWriteCharacteristics("932c32bd-0003-47a2-835a-a8d455b859dd", []byte{byte(min(max(float64(b), 1), 254))})
}

func (c *CenturaColorSpot) Model() string {
	return string(c.mustReadCharacteristics("00002a24-0000-1000-8000-00805f9b34fb"))
}

func (c *CenturaColorSpot) White() color.Color {
	//TODO implement me
	panic("implement me")
}

func (c *CenturaColorSpot) SetWhite(color color.Color) {
	//TODO implement me
	panic("implement me")
}

func (c *CenturaColorSpot) getGamut() *Gamut {
	if c.gamut == nil {
		gamut := GetGamutForModel(c.Model())
		c.gamut = &gamut
	}
	return c.gamut
}

func (c *CenturaColorSpot) Color() color.Color {
	byteColor := c.mustReadCharacteristics("932c32bd-0005-47a2-835a-a8d455b859dd")
	x := float64(binary.LittleEndian.Uint16(byteColor[0:2])) / 0xffff
	y := float64(binary.LittleEndian.Uint16(byteColor[2:4])) / 0xffff
	return c.getGamut().XYYToColor(XYPoint{x, y}, float64(c.Brightness())/255)
}

func (c *CenturaColorSpot) SetColor(colour color.Color) {
	xy := c.getGamut().ColorToXY(colour)
	xBits := uint16(xy[0] * 0xffff)
	yBits := uint16(xy[1] * 0xffff)
	written := []byte{byte(xBits), byte(xBits >> 8), byte(yBits), byte(yBits >> 8)}
	c.mustWriteCharacteristics("932c32bd-0005-47a2-835a-a8d455b859dd", written)
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
