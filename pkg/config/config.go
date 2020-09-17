package config

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

type CalibrationParameters struct {
	TemperatureOffset  float64 `yaml:"LM94022_TEMPERATURE_OFFSET"`
	SmuDCOffset        float64 `yaml:"RN171_SMU_DC_OFFSET"`
	SmuGain            float64 `yaml:"RN171_SMU_GAIN"`
	DCOffsetCorrection float64 `yaml:"SFH3710_DC_OFFSET_CORRECTION"`
	MoistureContinuity float64 `yaml:"SOIL_MOISTURE_DISCONTINUITY"`
	MoistureMin        float64 `yaml:"SOIL_MOISTURE_MIN"`
}

type Output struct {
	DbFile string `yaml:"db_file"`
}

type devices map[string]Device

type Device struct {
	Name                  string                `yaml:"name"`
	Key                   string                `yaml:"key"`
	CalibrationParameters CalibrationParameters `yaml:"calibration_parameters"`
}

type Config struct {
	LastConfigChange time.Time
	Output  Output  `yaml:"output"`
	Devices devices `yaml:"devices"`
}

func New() *Config {
	config := Config{}

	yml, err := ioutil.ReadFile("config/config.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = yaml.Unmarshal(yml, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return &config
}