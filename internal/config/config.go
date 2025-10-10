package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

const (
	MaxConnection = 20000
	OpRead        = 0
	OpWrite       = 1
	CRLF          = "\r\n"
	WithScore     = "WITHSCORES"
	BufferSize    = 1024
)

func SetConfigFile(path string) {
	viper.SetConfigName("config")
	viper.AddConfigPath(path)
	viper.SetConfigType("json")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error reading config file", err)
	}
}

func GetString(key string) string {
	return viper.GetString(fmt.Sprintf("%v", key))
}
func GetInt(key string) int {
	return viper.GetInt(fmt.Sprint(key))
}
