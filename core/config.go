package core

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

var Config *ServerConfig
var Hash uint32

type Role struct {
	Type    string
	Bind    string
	Version string
	Tuning  struct {
		Interest_Timeout int
	}
	Client struct {
		Add_Interest      string
		Write_Buffer_Size int
		Keepalive         int
		Relocate          bool
	}
	Channels struct {
		Min int
		Max int
	}
}

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
	Eventlogger struct {
		Bind   string
		Output string `"`
	}
	Roles []Role
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
