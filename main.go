package main

import (
	"encoding/binary"
	"strings"
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	mac, err := bluetooth.ParseMAC("A4:C1:38:7E:76:6C")
	must("parse a macadress", err)
	sensorMacAddr := bluetooth.Address{}
	sensorMacAddr.MAC = mac

	dataUuid, err := bluetooth.ParseUUID(strings.ToLower("ebe0ccb0-7a0a-4b0c-8a1a-6ff2997da3a6"))
	must("Failed parsing data uuid", err)
	unitUuid, err := bluetooth.ParseUUID(strings.ToLower("EBE0CCC1-7A0A-4B0C-8A1A-6FF2997DA3A6"))
	must("parse unit uuid", err)
	var foundDevice bluetooth.ScanResult
	var deviceConnection *bluetooth.Device
	must("enable BLE stack", adapter.Enable())
	println("Scanning...")
	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if device.Address == sensorMacAddr {
			foundDevice = device
			println("found device: ", device.Address.String(), device.RSSI, device.LocalName())
			err = adapter.StopScan()
			must("stop scan", err)
		}
	})
	must("start scan", err)

	println("Start looping connection")
	for {
		deviceConnection, err = adapter.Connect(foundDevice.Address, bluetooth.ConnectionParams{})
		if err != nil {
			println("Connecting failed: " + err.Error())
			time.Sleep(time.Second * time.Duration(3))
		} else {
			println("Connected to: " + foundDevice.Address.String())
			break
		}
	}
	defer disconnect(deviceConnection)

	var srvcs []bluetooth.DeviceService
	for {
		srvcs, err = deviceConnection.DiscoverServices([]bluetooth.UUID{})
		// print(errors.Is(err, Error("timeout")))
		if err != nil {
			println("failed to discover services:" + err.Error())
		} else {
			break
		}

	}
	var dataService bluetooth.DeviceService
	for i, srvc := range srvcs {
		println(i, " found service", srvc.UUID().String())
		if srvc.UUID() == dataUuid {
			println(i, " found service", srvc.UUID().String())
			dataService = srvc
		}
	}
	chars, err := dataService.DiscoverCharacteristics([]bluetooth.UUID{})
	must("discover characteristics", err)
	char := chars[0]
	for i, characteristic := range chars {
		println(i, " found characteristic", characteristic.UUID().String())
		if characteristic.UUID() == unitUuid {
			println(i, " found characteristic", characteristic.UUID().String())
			char = characteristic
		}
	}
	println("found characteristic", char.UUID().String())

	buf := make([]byte, 5)
	length, err := char.Read(buf)
	println(length)
	must("read data", err)
	buf2 := make([]byte, 2)
	copy(buf2, buf)
	test2 := binary.LittleEndian.Uint16(buf2)
	println(test2)

}

func disconnect(device *bluetooth.Device) {
	println("Disconnecting device")
	// err := device.Disconnect()
	// must("disconnect properly", err)
}

func must(action string, err error) {
	if err != nil {
		panic("Failed to " + action + ": " + err.Error())
	}
}
