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

type Sensor struct {
	Id         int64
	MacAddress string
	Name       string
}

type Type struct {
	Id   int64
	Name string
}

type Reading struct {
	Id             int64
	SensorId       int64
	RawValue       float64
	ConvertedValue float64
	Timestamp      int64
	TypeId         int64
}

func New(file string) *Database {
	db, _ := sql.Open("sqlite3", file)

	// create database structure if needed
	readings, _ := db.Prepare("create table if not exists readings ( id INTEGER constraint readings_pk primary key autoincrement, sensor INTEGER not null references sensors, rawvalue REAL, convertedvalue REAL, timestamp INTEGER not null, type INTEGER not null references types);")
	readings.Exec()

	sensors, _ := db.Prepare("create table if not exists sensors ( id INTEGER constraint sensors_pk primary key autoincrement, macaddress TEXT, name TEXT ); create unique index if not exists sensors_macaddress_uindex on sensors (macaddress);")
	sensors.Exec()

	types, _ := db.Prepare("create table if not exists types ( id INTEGER constraint types_pk primary key autoincrement, name TEXT not null ); create unique index if not exists types_name_uindex on types (name);")
	types.Exec()

	return &Database {
		Client: db,
	}
}

func (db *Database) getSensorId(macAddress string, device config.Device) int64 {
	row := db.Client.QueryRow("select id from sensors where macaddress = $1", macAddress)
	id := new(int64)
	err := row.Scan(id)
	if err == sql.ErrNoRows {
		statement, _ := db.Client.Prepare("insert into sensors (macaddress, name) values (?, ?)")
		result, _ := statement.Exec(macAddress, device.Name)
		lastInsertedId, _ := result.LastInsertId()
		return lastInsertedId
	}
	return *id
}

func (db *Database) getTypeId(readingType string) int64 {
	row := db.Client.QueryRow("select id from types where name = $1", readingType)
	id := new(int64)
	err := row.Scan(id)
	if err == sql.ErrNoRows {
		statement, _ := db.Client.Prepare("insert into types (name) values (?)")
		result, _ := statement.Exec(readingType)
		lastInsertedId, _ := result.LastInsertId()
		return lastInsertedId
	}
	return *id
}

func (db *Database) WriteReading(macAddress, readingType string, reading *sensors.Reading, device config.Device) {
	sensorId := db.getSensorId(macAddress, device)
	typeId := db.getTypeId(readingType)

	statement, _ := db.Client.Prepare("insert into readings (sensor, rawvalue, convertedvalue, timestamp, type) values (?, ?, ?, ?, ?)")
	defer statement.Close()

	statement.Exec(sensorId, reading.RawValue, reading.ConvertedValue, reading.Timestamp, typeId)
}

func (db *Database) GetReadings(days int) []*Reading {
	timestamp := time.Now().AddDate(0, 0, -days)
	rows, _ := db.Client.Query("select * from readings where timestamp > $1", timestamp.Unix())
	defer rows.Close()

	readings := make([]*Reading, 0)
	for rows.Next() {
		reading := new(Reading)
		err := rows.Scan(&reading.Id, &reading.SensorId, &reading.RawValue, &reading.ConvertedValue, &reading.Timestamp, &reading.TypeId)
		if err == sql.ErrNoRows {
			return readings
		}
		readings = append(readings, reading)
	}
	return readings
}