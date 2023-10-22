package golightmyroom

import "image/color"

type Light interface {
	On()
	Off()
	Powered() bool
	Brightness() byte
	SetBrightness(byte)
	Model() string
}

type TemperatureControl interface {
	Temperature() uint16
	SetTemperature(uint16)
}

type ColorControl interface {
	Color() color.Color
	SetColor(color.Color)
}

type NameControl interface {
	Name() string
	SetName(string)
}
