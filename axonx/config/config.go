package config

import "io/ioutil"
import "encoding/json"

type Config struct {
	SrvAddress   string        `json:"srv_address,omitempty"`
	DebugAddress string        `json:"debug_address,omitempty"`
	Version      string        `json:"version,omitempty"`
	Ratelimit    int           `json:"ratelimit,omitempty"`
	Influx       *InfluxConfig `json:"influx,omitempty"`
}

type InfluxConfig struct {
	Address      string `json:"address,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Databasename string `json:"databasename,omitempty"`
}

var cfg = new(Config)

func LoadCfg(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil
	}
	return nil
}

func GetCfg() *Config {
	return cfg
}
