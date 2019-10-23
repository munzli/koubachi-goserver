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

	// TODO dynamic
	response := []byte("current_time=1571731475&transmit_interval=55202&transmit_app_led=1&sensor_app_led=0&day_threshold=10.0&sensor_enabled[1]=0&sensor_enabled[2]=1&sensor_enabled[6]=1&sensor_enabled[7]=1&sensor_enabled[8]=1&sensor_enabled[9]=1&sensor_enabled[10]=1&sensor_enabled[11]=1&sensor_enabled[12]=1&sensor_enabled[15]=1&sensor_enabled[29]=1&sensor_enabled[4096]=0&sensor_enabled[4112]=0&sensor_enabled[4113]=0&sensor_enabled[4114]=0&sensor_enabled[4115]=0&sensor_enabled[4116]=0&sensor_enabled[4128]=0&sensor_enabled[8192]=0&sensor_enabled[8193]=0&sensor_enabled[8194]=0&sensor_enabled[8195]=0&sensor_polling_interval[2]=86400&sensor_polling_interval[7]=3600&sensor_polling_interval[8]=3600&sensor_polling_interval[10]=18000&sensor_polling_interval[15]=3600&sensor_polling_interval[29]=3600")
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