package core

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"sync"
)

var (
	c        *ServerConfig
	confOnce sync.Once
)

type ServerConfig struct {
	Daemon struct {
		Name string `yaml:"name"`
	} `yaml:"daemon"`
	General struct {
		EventloggerIP string   `yaml:"eventlogger"`
		DCFiles       []string `yaml:"dc_files"`
	} `yaml:"general"`
	Uberdogs []struct {
		ID        int    `yaml:"id"`
		Class     string `yaml:"class"`
		Anonymous string `yaml:"anonymous"`
	} `yaml:"uberdogs"`
	MessageDirector struct {
		Bind string `yaml:"bind"`
	} `yaml:"messagedirector"`
}

func LoadConfig() *ServerConfig {
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

func GetConfig() *ServerConfig {
	confOnce.Do(func() {
		c = LoadConfig()
	})

	if c == nil {
		os.Exit(1)
	}

	return c
}
