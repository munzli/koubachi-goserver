package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"koubachi-goserver/pkg/config"
	"koubachi-goserver/pkg/crypto"
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
	a := r.Group("/v1")
	{
		device := a.Group("/smart_devices")
		{
			device.PUT("/:macAddress", api.connect)
			device.POST("/:macAddress/config", api.config)
			device.POST("/:macAddress/readings", api.readings)
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
	sensors := sensors.GetSensors()
	var configStrings []string
	configStrings = append(configStrings,  fmt.Sprintf("current_time=%d", time.Now().Unix()))
	configStrings = append(configStrings,  "transmit_interval=55202", "transmit_app_led=1", "sensor_app_led=0", "day_threshold=10.0")
	for key, sensor := range sensors {
		configStrings = append(configStrings, fmt.Sprintf("sensor_enabled[%d]=%d", key, bool2int(sensor.Enabled)))
		if sensor.PollingInterval > 0 {
			configStrings = append(configStrings, fmt.Sprintf("sensor_polling_interval[%d]=%d", key, sensor.PollingInterval))
		}
	}

	response := []byte(strings.Join(configStrings[:], "&"))
	responseEncoded := crypto.Encrypt(key, []byte(response))

	c.Data(http.StatusOK, ContentType, responseEncoded)
}

func (api *API) readings(c *gin.Context) {
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
	json.Unmarshal(body, &data)

	sensors := sensors.GetSensors()
	for _, reading  := range data.Readings {
		// map special sensor persist
		mapper := sensors[reading.Code]

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

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}