package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"koubachi-goserver/pkg/config"
	"koubachi-goserver/pkg/crypto"
	"koubachi-goserver/pkg/model"
	"koubachi-goserver/pkg/sensors"
	"koubachi-goserver/pkg/sqlite"
	"log"
	"net/http"
	"strings"
	"time"
)

const ContentType = "application/x-koubachi-aes-encrypted"

type API struct {
	Config *config.Config
	Sqlite *sqlite.Database
}

func New(config *config.Config) *API {
	return &API{
		Config: config,
		Sqlite: sqlite.New(config.Output.DbFile),
	}
}

func (api *API) AttachRoutes(r *gin.RouterGroup) {

	// index
	r.StaticFile("", "./assets/index.html")
	r.StaticFile("/favicon.ico", "./assets/favicon.ico")
	r.Static("/js", "./assets/js")

	// api
	a := r.Group("/v1")
	{
		device := a.Group("/smart_devices")
		{
			device.GET("", api.getDevices)
			device.PUT("/:macAddress", api.connect)
			device.POST("/:macAddress/config", api.config)
			device.POST("/:macAddress/readings", api.postReadings)

			device.GET("/:macAddress/soil_moisture", api.getReadings(model.SoilMoisture))
			device.GET("/:macAddress/battery_voltage", api.getReadings(model.BatteryVoltage))
			device.GET("/:macAddress/soil_temperature", api.getReadings(model.SoilTemperature))
			device.GET("/:macAddress/temperature", api.getReadings(model.Temperature))
			device.GET("/:macAddress/light", api.getReadings(model.Light))
			device.GET("/:macAddress/rssi", api.getReadings(model.Rssi))
		}
	}
}

func (api *API) connect(c *gin.Context) {
	macAddress := c.Param("macAddress")
	deviceKey := api.Config.Devices[macAddress].Key

	rawData, err := c.GetRawData()
	if err != nil {
		log.Panicf("error: %v", err)
	}

	key, _ := hex.DecodeString(deviceKey)
	body := crypto.Decrypt(key, rawData)

	// do nothing with body
	// persist.WriteSensor(macAddress, api.Config.Devices[macAddress])
	_ = body

	response := fmt.Sprintf("current_time=%d&last_config_change=%d", time.Now().Unix(), api.Config.LastConfigChange.Unix())
	responseEncoded := crypto.Encrypt(key, []byte(response))

	c.Data(http.StatusOK, ContentType, responseEncoded)
}

func (api *API) config(c *gin.Context) {
	macAddress := c.Param("macAddress")
	deviceKey := api.Config.Devices[macAddress].Key

	rawData, err := c.GetRawData()
	if err != nil {
		log.Panicf("error: %v", err)
	}

	key, _ := hex.DecodeString(deviceKey)
	body := crypto.Decrypt(key, rawData)

	// do nothing with body
	_ = body

	// create sensor configuration
	sensorData := sensors.GetSensors()
	var configStrings []string
	configStrings = append(configStrings,  fmt.Sprintf("current_time=%d", time.Now().Unix()))
	configStrings = append(configStrings,  "transmit_interval=14400", "transmit_app_led=1", "sensor_app_led=0", "day_threshold=10.0")
	for key, sensor := range sensorData {
		configStrings = append(configStrings, fmt.Sprintf("sensor_enabled[%d]=%d", key, bool2int(sensor.Enabled)))
		if sensor.PollingInterval > 0 {
			configStrings = append(configStrings, fmt.Sprintf("sensor_polling_interval[%d]=%d", key, sensor.PollingInterval))
		}
	}

	response := []byte(strings.Join(configStrings[:], "&"))
	responseEncoded := crypto.Encrypt(key, response)

	c.Data(http.StatusOK, ContentType, responseEncoded)
}

func (api *API) postReadings(c *gin.Context) {
	macAddress := c.Param("macAddress")
	deviceKey := api.Config.Devices[macAddress].Key

	rawData, err := c.GetRawData()
	if err != nil {
		log.Panicf("error: %v", err)
	}

	key, _ := hex.DecodeString(deviceKey)
	body := crypto.Decrypt(key, rawData)

	// do something with body
	data := sensors.Data{}
	_ = json.Unmarshal(body, &data)

	sensorData := sensors.GetSensors()
	for _, reading  := range data.Readings {
		// map special sensor persist
		mapper := sensorData[reading.Code]

		// special conversion of value
		reading.ConvertedValue = reading.RawValue
		if mapper.ConversionFunc != nil {
			reading.ConvertedValue = mapper.ConversionFunc(reading.RawValue, api.Config.Devices[macAddress].CalibrationParameters)
		}

		api.Sqlite.WriteReading(macAddress, mapper.Type, reading, api.Config.Devices[macAddress])
	}

	response := fmt.Sprintf("current_time=%d&last_config_change=%d", time.Now().Unix(), api.Config.LastConfigChange.Unix())
	responseEncoded := crypto.Encrypt(key, []byte(response))

	c.Data(http.StatusCreated, ContentType, responseEncoded)
}

func (api *API) getReadings(sensor string) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		macAddress := c.Param("macAddress")

		// get necessary ids to query database
		sensorId := api.Sqlite.GetSensorId(sensor)
		deviceId := api.Sqlite.GetDeviceId(macAddress, api.Config.Devices[macAddress])
		readings := api.Sqlite.GetReadings(deviceId, sensorId, 14)

		data := make([]model.ChartData, 0)
		for _, reading  := range readings {
			chartData := model.ChartData{
				T: time.Unix(reading.Timestamp, 0),
				Y: reading.ConvertedValue,
			}
			data = append(data, chartData)
		}

		c.JSON(http.StatusOK, data)
	}
	return fn
}

func (api *API) getDevices(c *gin.Context) {

	devices := api.Sqlite.GetDevices()

	data := make([]model.Device, 0)
	for _, device  := range devices {
		deviceData := model.Device{
			Id: device.Id,
			MacAddress: device.MacAddress,
			Name: device.Name,
		}
		data = append(data, deviceData)
	}

	c.JSON(http.StatusOK, data)
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}