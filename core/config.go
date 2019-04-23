package core

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
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
		Bind string
	}
}

func loadConfig() *ServerConfig {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Printf("%v", err)
		return nil
	}

	conf := &ServerConfig{}
	err = viper.Unmarshal(conf)
	if err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}

	return conf
}

func init() {
	Config = loadConfig()
	if Config == nil {
		os.Exit(1)
	}
}
