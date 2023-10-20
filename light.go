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

type WhiteControl interface {
	White() color.Color
	SetWhite(color.Color)
}

type ColorControl interface {
	Color() color.Color
	SetColor(color.Color)
}

type NameControl interface {
	Name() string
	SetName(string)
}
