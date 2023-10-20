package golightmyroom

import (
	"context"
	"fmt"
	"github.com/flonja/golightmyroom/models"
	"time"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func ConnectToBluetooth(macAddress string) (Light, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return ConnectToBluetoothWithContext(macAddress, ctx)
}

func ConnectToBluetoothWithContext(macAddress string, ctx context.Context) (Light, error) {
	if err := adapter.Enable(); err != nil {
		return nil, err
	}

	ch := make(chan bluetooth.ScanResult, 1)
	if err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if result.Address.String() == macAddress {
			if err := adapter.StopScan(); err != nil {
				panic(err)
			}
			ch <- result
		}
	}); err != nil {
		return nil, err
	}

	var device *bluetooth.Device
	var err error
	select {
	case result := <-ch:
		device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("context deadline exceeded")
	}

	// TODO: recognise models
	model, err := models.NewCenturaColorSpot(device)
	return model, err
}
