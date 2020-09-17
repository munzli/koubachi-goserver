package model

type Sensor struct {
	Id         int64
	MacAddress string
	Name       string
}

type Type struct {
	Id   int64
	Name string
}
