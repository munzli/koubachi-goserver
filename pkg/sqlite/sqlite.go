package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"koubachi-goserver/pkg/config"
	"koubachi-goserver/pkg/sensors"
)

type Database struct {
	Client *sql.DB
}

func New(file string) *Database {
	db, _ := sql.Open("sqlite3", file)

	// create database structure if needed
	readings, _ := db.Prepare("create table if not exists readings ( id INTEGER constraint readings_pk primary key autoincrement references sensors, sensor INTEGER not null, rawvalue REAL, convertedvalue REAL, timestamp INTEGER not null, type INTEGER not null references types );")
	readings.Exec()

	sensors, _ := db.Prepare("create table if not exists sensors ( id INTEGER constraint sensors_pk primary key autoincrement, macaddress TEXT, name TEXT ); create unique index if not exists sensors_macaddress_uindex on sensors (macaddress);")
	sensors.Exec()

	types, _ := db.Prepare("create table if not exists types ( id INTEGER constraint types_pk primary key autoincrement, name TEXT not null ); create unique index if not exists types_name_uindex on types (name);")
	types.Exec()

	return &Database {
		Client: db,
	}
}

func (db *Database) getSensor(macAddress string, device config.Device) int64 {
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

func (db *Database) getType(readingType string) int64 {
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
	sensorId := db.getSensor(macAddress, device)
	typeId := db.getType(readingType)

	statement, _ := db.Client.Prepare("insert into readings (sensor, rawvalue, convertedvalue, timestamp, type) values (?, ?, ?, ?, ?)")
	statement.Exec(sensorId, reading.RawValue, reading.ConvertedValue, reading.Timestamp, typeId)
}
