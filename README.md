# koubachi-goserver
this is a golang version of https://github.com/koalatux/koubachi-pyserver and is just a proof of concept atm

the conversions of values hasn't been fully tested

### run
create a config file (see `config/config.yml.example`)

change koubachi endpoint and wi-fi access (see https://github.com/koubachi-sensor/api-docs#change-the-sensors-server-address)
```
curl -X GET -G http://172.29.0.1/sos_config -d host=10.22.4.56 -d port=8005
```

build and run the container
```
docker build -t koubachi-goserver .
docker run -v $(pwd)/config:/app/config:ro -v $(pwd)/readings:/app/readings -p 8005:8005 koubachi-goserver
```