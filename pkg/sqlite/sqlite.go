package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"koubachi-goserver/pkg/config"
	"koubachi-goserver/pkg/sensors"
	"time"
)

type Database struct {
	Client *sql.DB
}

type Device struct {
	Id         int64
	MacAddress string
	Name       string
}

type Sensor struct {
	Id   int64
	Name string
}

type Reading struct {
	Id             int64
	DeviceId       int64
	RawValue       float64
	ConvertedValue float64
	Timestamp      int64
	SensorId       int64
}

func New(file string) *Database {
	db, _ := sql.Open("sqlite3", file)

	// create database structure if needed
	readings, _ := db.Prepare("create table if not exists readings ( id INTEGER constraint readings_pk primary key autoincrement, device INTEGER not null references devices, rawvalue REAL, convertedvalue REAL, timestamp INTEGER not null, sensor INTEGER not null references sensors);")
	readings.Exec()

	devices, _ := db.Prepare("create table if not exists devices ( id INTEGER constraint devices_pk primary key autoincrement, macaddress TEXT, name TEXT ); create unique index if not exists devices_macaddress_uindex on devices (macaddress);")
	devices.Exec()

	sensors, _ := db.Prepare("create table if not exists sensors ( id INTEGER constraint sensors_pk primary key autoincrement, name TEXT not null ); create unique index if not exists sensors_name_uindex on sensors (name);")
	sensors.Exec()

	return &Database {
		Client: db,
	}
}

func (db *Database) GetDeviceId(macAddress string, device config.Device) int64 {
	row := db.Client.QueryRow("select id from devices where macaddress = $1", macAddress)
	id := new(int64)
	err := row.Scan(id)
	if err == sql.ErrNoRows {
		statement, _ := db.Client.Prepare("insert into devices (macaddress, name) values (?, ?)")
		result, _ := statement.Exec(macAddress, device.Name)
		lastInsertedId, _ := result.LastInsertId()
		return lastInsertedId
	}
	return *id
}

func (db *Database) GetSensorId(sensor string) int64 {
	row := db.Client.QueryRow("select id from sensors where name = $1", sensor)
	id := new(int64)
	err := row.Scan(id)
	if err == sql.ErrNoRows {
		statement, _ := db.Client.Prepare("insert into sensors (name) values (?)")
		result, _ := statement.Exec(sensor)
		lastInsertedId, _ := result.LastInsertId()
		return lastInsertedId
	}
	return *id
}

func (db *Database) WriteReading(macAddress, sensor string, reading *sensors.Reading, device config.Device) {
	deviceId := db.GetDeviceId(macAddress, device)
	sensorId := db.GetSensorId(sensor)

	statement, _ := db.Client.Prepare("insert into readings (device, rawvalue, convertedvalue, timestamp, sensor) values (?, ?, ?, ?, ?)")
	defer statement.Close()

	statement.Exec(deviceId, reading.RawValue, reading.ConvertedValue, reading.Timestamp, sensorId)
}

func (db *Database) GetReadings(deviceId, sensorId int64, days int) []*Reading {
	timestamp := time.Now().AddDate(0, 0, -days)
	rows, _ := db.Client.Query("select * from readings where timestamp > $1 and device = $2 and sensor = $3", timestamp.Unix(), deviceId, sensorId)
	defer rows.Close()

	readings := make([]*Reading, 0)
	for rows.Next() {
		reading := new(Reading)
		err := rows.Scan(&reading.Id, &reading.DeviceId, &reading.RawValue, &reading.ConvertedValue, &reading.Timestamp, &reading.SensorId)
		if err == sql.ErrNoRows {
			return readings
		}
		readings = append(readings, reading)
	}
	return readings
}

func (db *Database) GetDevices() []*Device {
	rows, _ := db.Client.Query("select * from devices")
	defer rows.Close()

	devices := make([]*Device, 0)
	for rows.Next() {
		device := new(Device)
		err := rows.Scan(&device.Id, &device.MacAddress, &device.Name)
		if err == sql.ErrNoRows {
			return devices
		}
		devices = append(devices, device)
	}
	return devices
}