package models

import "tinygo.org/x/bluetooth"

type BluetoothControlled interface {
	bluetoothDevice() *bluetooth.Device
	Disconnect() error
}
