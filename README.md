# koubachi-goserver
this is a golang version of https://github.com/koalatux/koubachi-pyserver and is just a proof of concept atm

the conversions of values hasn't been fully tested

### run
create a config file (see `config/config.yml.example`)

change koubachi endpoint and wi-fi access (see https://github.com/koubachi-sensor/api-docs#change-the-sensors-server-address)
```
curl -X GET -G http://172.29.0.1/sos_config -d host=192.168.1.119 -d port=8005
```

build and run the container
```
docker build -t koubachi-goserver .
docker run -v $(pwd)/config:/app/config:ro -v $(pwd)/readings:/app/readings -p 8005:8005 koubachi-goserver
```

### sqlite tables
```
create table readings
(
    id             INTEGER
        constraint readings_pk
            primary key autoincrement
        references sensors,
    sensor         INTEGER not null,
    rawvalue       REAL,
    convertedvalue REAL,
    timestamp      INTEGER not null,
    type           INTEGER not null
        references types
);

create table sensors
(
    id         INTEGER
        constraint sensors_pk
            primary key autoincrement,
    macaddress TEXT,
    name       TEXT
);

create unique index sensors_macaddress_uindex
    on sensors (macaddress);

create table types
(
    id   INTEGER
        constraint types_pk
            primary key autoincrement,
    name TEXT not null
);

create unique index types_name_uindex
    on types (name);
```