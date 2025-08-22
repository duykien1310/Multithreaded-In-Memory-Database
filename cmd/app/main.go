package main

import (
	"backend/internal/config"
	"backend/internal/server"
	"log"
)

func init() {
	config.SetConfigFile("../../internal/config")
}

func main() {
	config := config.EnvConfig{
		Host:     config.GetString("host.address"),
		Port:     config.GetInt("host.port"),
		Protocol: config.GetString("protocol"),
	}

	s := server.NewServer(config.Host, config.Port, config.Protocol)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
