package models

import "time"

type Sensor struct {
	Id         int64  `json:"id"`
	MacAddress string `json:"macAddress"`
	Name       string `json:"name"`
}

type Type struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Reading struct {
	Id             int64     `json:"id"`
	Sensor         Sensor    `json:"sensor"`
	RawValue       float64   `json:"rawValue"`
	ConvertedValue float64   `json:"convertedValue"`
	Timestamp      time.Time `json:"timestamp"`
	Type           Type      `json:"type"`
}

type ChartData struct {
	T time.Time `json:"t"`
	Y float64   `json:"y"`
}