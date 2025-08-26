package main

import (
	"backend/internal/command"
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

	h := command.NewHandler()

	s := server.NewServer(config.Host, config.Port, config.Protocol, h)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
