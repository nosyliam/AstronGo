package core

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

var Config *ServerConfig

type ServerConfig struct {
	Daemon struct {
		Name string
	}
	General struct {
		Eventlogger string
		DC_Files    []string
	}
	Uberdogs []struct {
		ID        int
		Class     string
		Anonymous string
	}
	MessageDirector struct {
		Bind    string
		Connect string
	}
}

func LoadConfig(path string, name string) (err error) {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(path)
	viper.SetConfigName(name)

	if err := viper.ReadInConfig(); err != nil {
		return errors.New(fmt.Sprintf("Unable to load configuration file: %v", err))
	}

	conf := &ServerConfig{}
	if err := viper.Unmarshal(conf); err != nil {
		return errors.New(fmt.Sprintf("Unable to decode configuration file: %v", err))
	}

	Config = conf
	return nil
}
