package api

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"koubachi-goserver/pkg/config"
	"koubachi-goserver/pkg/crypto"
	"koubachi-goserver/pkg/sensors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const ContentType = "application/x-koubachi-aes-encrypted"

type API struct {
	Config *config.Config
}

func NewAPI(config *config.Config) *API {
	return &API{
		Config: config,
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
		// map special sensor data
		mapper := sensors[reading.Code]

		// special conversion of value
		value := reading.Value
		if mapper.ConversionFunc != nil {
			value = mapper.ConversionFunc(reading.Value, api.Config.Devices[macAddress].CalibrationParameters)
		}

		// prepare for csv
		path := fmt.Sprintf("%s/%s_%s.csv", api.Config.Output.Directory, macAddress, mapper.Type)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			writeCSVln([]string{"timestamp",mapper.Type,"raw_value"}, path)
		}
		record := []string{
			fmt.Sprintf("%d", reading.Timestamp),
			fmt.Sprintf("%f", value),
			fmt.Sprintf("%f", reading.Value),
		}
		writeCSVln(record, path)
	}

	response := fmt.Sprintf("current_time=%d&last_config_change=%d", time.Now().Unix(), api.Config.LastConfigChange.Unix())
	responseEncoded := crypto.Encrypt(key, []byte(response))

	c.Data(http.StatusCreated, ContentType, responseEncoded)
}

func writeCSVln(record []string, path string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer file.Close()

	if err != nil {
		log.Panicf("error: %v", err)
	}

	csvWriter := csv.NewWriter(file)
	_ = csvWriter.Write(record) // ignore errors for now
	csvWriter.Flush()
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}