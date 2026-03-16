package main

import (
	"net/http"
	"os"
)

func main() {
	logger := NewLogger(GetLogLevel())
	defer logger.Close()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		logger.Error("Failed to load config from %s: %v", configPath, err)
	}

	logger.Info("Config loaded: OAuth URL: %s, Scope: %s", config.OAuthURL, config.Scope)

	tokenManager := NewTokenManager(config, logger)
	provider := NewGigaChatProvider(tokenManager, logger)

	handlers := NewHTTPHandlers(provider)
	router := NewRouter(handlers, logger)

	server := &http.Server{
		Addr:    config.Addr + ":" + config.Port,
		Handler: router,
	}

	logger.Info("Starting OpenAI-compatible provider for GigaChat")
	logger.Info("Base URL: http://%s:%s/v1", config.Addr, config.Port)
	logger.Info("Config: %s, Log Level: %s", configPath, os.Getenv("LOG_LEVEL"))
	logger.Info("Health check: curl -X GET http://%s:%s/v1/models -H \"Authorization: Bearer test\"", config.Addr, config.Port)

	if err := server.ListenAndServe(); err != nil {
		logger.Error("Server failed to start: %v", err)
	}
}
