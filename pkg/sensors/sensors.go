package sensors

import (
	"encoding/json"
	"errors"
	"math"

	"koubachi-goserver/pkg/config"
)

type Sensors struct {
	Type            string
	Enabled         bool
	PollingInterval int
	ConversionFunc  func(x float64, config config.CalibrationParameters) float64
}

type Data struct {
	Timestamp int        `json:"timestamp"`
	Readings  []*Reading `json:"readings"`
}

type Reading struct {
	Timestamp int
	Code      int
	Value     float64
}

func (r *Reading) UnmarshalJSON(buf []byte) error {
	tmp := []interface{}{&r.Timestamp, &r.Code, &r.Value}
	wantLen := len(tmp)
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if g, e := len(tmp), wantLen; g != e {
		return errors.New("wrong number of fields in reading")
	}
	return nil
}

func GetSensors() map[int]Sensors {
	sensors := map[int]Sensors{}

	sensors[1] = Sensors{
		Type:            "board_temperature",
		Enabled:         false,
		PollingInterval: 3600,
		ConversionFunc:  nil,
	}
	sensors[2] = Sensors{
		Type:            "battery_voltage",
		Enabled:         true,
		PollingInterval: 86400,
		ConversionFunc:  nil,
	}
	sensors[6] = Sensors{
		Type:            "button",
		Enabled:         true,
		PollingInterval: 0,
		ConversionFunc:  func(x float64, config config.CalibrationParameters) float64 {
			return x / 1000
		},
	}
	sensors[7] = Sensors{
		Type:            "temperature",
		Enabled:         true,
		PollingInterval: 3600,
		ConversionFunc:  convertLm94022Temperature,
	}
	sensors[8] = Sensors{
		Type:            "light",
		Enabled:         true,
		PollingInterval: 3600,
		ConversionFunc:  convertSfh3710Light,
	}
	sensors[9] = Sensors{
		Type:            "rssi",
		Enabled:         true,
		PollingInterval: 0,
		ConversionFunc:  nil,
	}
	sensors[10] = Sensors{
		Type:            "soil_sensors_trigger",
		Enabled:         true,
		PollingInterval: 18000,
		ConversionFunc:  nil,
	}
	sensors[11] = Sensors{
		Type:            "soil_temperature",
		Enabled:         true,
		PollingInterval: 18000,
		ConversionFunc:  func(x float64, config config.CalibrationParameters) float64 {
			return x - 2.5
		},
	}
	sensors[12] = Sensors{
		Type:            "soil_moisture",
		Enabled:         true,
		PollingInterval: 0,
		ConversionFunc:  convertSoilMoisture,
	}
	sensors[15] = Sensors{
		Type:            "temperature",
		Enabled:         true,
		PollingInterval: 0,
		ConversionFunc:  func(x float64, config config.CalibrationParameters) float64 {
			return -46.85 + 175.72 * x / math.Pow(2,16)
		},
	}
	sensors[29] = Sensors{
		Type:            "light",
		Enabled:         true,
		PollingInterval: 0,
		ConversionFunc:  convertTsl2561Light,
	}

	return sensors
}

func convertLm94022Temperature(x float64, config config.CalibrationParameters) float64 {
	x = (x - config.SmuDCOffset) * config.SmuGain * 3.0
	x = 453.512485591335 - 163.565776259726 * x - 10.5408332222805 * math.Pow(x, 2) - config.TemperatureOffset - 273.15
	return x
}

func convertSfh3710Light(x float64, config config.CalibrationParameters) float64 {
	x = (x - config.DCOffsetCorrection) * config.SmuGain / 20.0 * 7.2
	x = 3333326.67 * ((math.Abs(x) + x) / 2)
	return x
}

func convertSoilMoisture(x float64, config config.CalibrationParameters) float64 {
	x = (x - config.MoistureMin) *
		((8778.25 - 3515.25) / (config.MoistureContinuity -
		config.MoistureMin)) + 3515.25
	x = 8.130159393183e-018 * math.Pow(x, 5) -
		0.000000000000259586800701037 * math.Pow(x, 4) +
		0.00000000328783014726288 * math.Pow(x, 3) -
		0.0000206371829755294 * math.Pow(x, 2) +
		0.0646453707101697 * x -
		79.7740602786336
	return math.Max(0.0, math.Min(6.0, x))
}

func convertTsl2561Light(x float64, config config.CalibrationParameters) float64 {
	intVal := int(x)
	data0 := float64((intVal >> 16) & 0xfffe)
	data1 := float64(intVal & 0xfffe)
	gain := (intVal >> 16) & 0x1
	intTime := intVal & 0x1
	if gain == 0x0 {
		data0 *= 16
		data1 *= 16
	}
	if intTime == 0x0 {
		data0 *= 1 / 0.252
		data1 *= 1 / 0.252
	}

	y := 0.0304 * data0 - 0.062 * data0 * math.Pow(data1 / data0, 1.4)
	if data0 == 0 || data1 / data0 > 1.30 {
		y = 0.0
	} else if data1 / data0 > 0.8 {
		y = 0.00146 * data0 - 0.00112 * data1
	} else if data1 / data0 > 0.61 {
		y = 0.0128 * data0 - 0.0153 * data1
	} else if data1 / data0 > 0.50 {
		y = 0.0224 * data0 - 0.031 * data1
	}
	return y * 5.0
}