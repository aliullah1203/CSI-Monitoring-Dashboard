package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Version          string
	ServiceName      string
	HttpPort         int
	ConnectionString string
	CandidateID      string
	MQTTBroker       string
	MQTTPort         int
	FrontendURL      string
}

var configuration *Config

func loadConfig() {
	err := godotenv.Load()
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			fmt.Println("Warning: error loading .env file:", err)
		}
	}

	version := os.Getenv("VERSION")
	if version == "" {
		version = "1.0"
	}
	serviceName := os.Getenv("SERVICENAME")
	if serviceName == "" {
		serviceName = "production-monitor"
	}

	httpPortStr := os.Getenv("HTTPPORT")
	if httpPortStr == "" {
		httpPortStr = "8080"
	}
	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil {
		fmt.Println("Failed to parse HTTPPORT:", err)
		os.Exit(1)
	}

	connectionString := os.Getenv("DB_STRING")
	if connectionString == "" {
		fmt.Println("DB_STRING is required!")
		os.Exit(1)
	}

	candidateID := os.Getenv("CANDIDATE_ID")
	if candidateID == "" {
		candidateID = "CAND-05"
	}

	mqttBroker := os.Getenv("MQTT_BROKER")
	if mqttBroker == "" {
		mqttBroker = "152.42.238.142"
	}

	mqttPortStr := os.Getenv("MQTT_PORT")
	if mqttPortStr == "" {
		mqttPortStr = "1883"
	}
	mqttPort, _ := strconv.Atoi(mqttPortStr)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	configuration = &Config{
		Version:          version,
		ServiceName:      serviceName,
		HttpPort:         httpPort,
		ConnectionString: connectionString,
		CandidateID:      candidateID,
		MQTTBroker:       mqttBroker,
		MQTTPort:         mqttPort,
		FrontendURL:      frontendURL,
	}
}

func GetConfig() *Config {
	if configuration == nil {
		loadConfig()
	}
	return configuration
}
