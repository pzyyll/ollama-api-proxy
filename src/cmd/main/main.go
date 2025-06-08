package main

import (
	"log/slog"
	"net/http"
	"os"

	"ollama-api-proxy/src/internal/config"
	"ollama-api-proxy/src/internal/core"
	"ollama-api-proxy/src/internal/state"

	"github.com/joho/godotenv"
)

var logLevel = new(slog.LevelVar)

func initConfig() *config.Config {
	// Initialize the log level with a default value
	logLevel.Set(slog.LevelInfo)
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	slog.SetDefault(slog.New(h))

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		// Skip error if .env file is not found
		if !os.IsNotExist(err) {
			slog.Error("Error loading .env file", "error", err)
			panic(err)
		}
	}

	// Load config
	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		panic(err)
	}

	// Set log level from environment variable if it exists
	var level slog.Level
	if err := level.UnmarshalText([]byte(config.LogLevel)); err != nil {
		slog.Error("Invalid log level", "level", config.LogLevel, "error", err)
		panic(err)
	}
	logLevel.Set(level)
	return config
}

func main() {
	// Run the application
	cfg := initConfig()
	models, _ := config.LoadModels("models.yml")

	appState := &state.State{
		Config: cfg,
		Models: models,
		HttpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
	
	engine := core.InitRouterEngine(appState)
	core.Run(engine, cfg)
}
