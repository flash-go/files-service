package main

import (
	"os"

	// SDK
	"github.com/flash-go/sdk/config"
	"github.com/flash-go/sdk/state"

	// Other
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Create state service
	stateService := state.New(os.Getenv("CONSUL_AGENT"))

	// Create config
	cfg := config.New(
		stateService,
		os.Getenv("SERVICE_NAME"),
	)

	// Set KV from env map
	cfg.SetEnvMap(envMap)
}
