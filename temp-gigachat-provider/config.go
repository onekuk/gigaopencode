package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	AuthorizationKey string `json:"authorization_key"`
	OAuthURL         string `json:"oauth_url"`
	Scope            string `json:"scope"`
	Addr             string `json:"addr"`
	Port             string `json:"port"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Установка значений по умолчанию
	if config.OAuthURL == "" {
		config.OAuthURL = "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	}
	if config.Scope == "" {
		config.Scope = "GIGACHAT_API_PERS"
	}
	if config.Addr == "" {
		config.Addr = "localhost"
	}
	if config.Port == "" {
		config.Port = "8080"
	}

	// Переопределение из переменных окружения
	if envAddr := os.Getenv("ADDR"); envAddr != "" {
		config.Addr = envAddr
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		config.Port = envPort
	}

	if config.AuthorizationKey == "" {
		return nil, fmt.Errorf("authorization_key is required in config")
	}

	return &config, nil
}