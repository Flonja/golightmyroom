package main

import "image/color"

// Light holds the main controls that every Hue light has.
type Light interface {
	// On turns on the light.
	On()
	// Off turns off the light.
	Off()
	// Powered returns true if the light is turned on and returns false if it's off.
	Powered() bool
	// Brightness returns a value from [1.0, 0.0], highest meaning brightest.
	Brightness() float64
	// SetBrightness set a brightness from [1.0, 0.0], highest meaning brightest.
	SetBrightness(float64)
	// Model shows the model name of the light, this is used for ColorControl.
	Model() string
}

// TemperatureControl hold the temperature controls for Hue lights that support it.
type TemperatureControl interface {
	// Temperature returns a temperature as Kelvin. (Search for "kelvin color temperature scale" for examples
	// or use default values, like TemperatureDefault)
	Temperature() uint16
	// SetTemperature allows you to set a Kelvin color temperature for the light.
	SetTemperature(uint16)
}

// ColorControl hold the color controls for Hue lights that support it.
type ColorControl interface {
	// Color returns the current color.
	Color() color.Color
	// SetColor sets the (supported) color of the light.
	// It may not be the exact color you set, since some lights have a limited color gamut.
	SetColor(color.Color)
}

// NameControl holds the name controls that every Hue light has.
// Though this has been seperated if some lights should not have their name changed.
type NameControl interface {
	// Name gets the light's name
	Name() string
	// SetName sets the light's name
	SetName(string)
}
