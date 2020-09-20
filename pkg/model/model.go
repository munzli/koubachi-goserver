package model

import "time"

const SoilSensorsTrigger = "soil_sensors_trigger"
const BoardTemperature   = "board_temperature"
const SoilTemperature    = "soil_temperature"
const BatteryVoltage     = "battery_voltage"
const SoilMoisture       = "soil_moisture"
const Temperature        = "temperature"
const Button             = "button"
const Light              = "light"
const Rssi               = "rssi"

type Device struct {
	Id         int64  `json:"id"`
	MacAddress string `json:"macAddress"`
	Name       string `json:"name"`
}

type Sensor struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Reading struct {
	Id             int64     `json:"id"`
	Device         Device    `json:"device"`
	RawValue       float64   `json:"rawValue"`
	ConvertedValue float64   `json:"convertedValue"`
	Timestamp      time.Time `json:"timestamp"`
	Sensor         Sensor    `json:"sensor"`
}

type ChartData struct {
	T time.Time `json:"t"`
	Y float64   `json:"y"`
}