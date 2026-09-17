package main

import (
	"log"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/api"
	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/config"
	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/provider"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	llm, err := provider.NewFactory().Create(cfg.Provider)
	if err != nil {
		log.Fatal(err)
	}

	server := api.NewServer(llm, cfg.MaxInputChars)
	log.Printf("Serving Gin LLM API demo on http://%s", cfg.Address)
	if err := server.Run(cfg.Address); err != nil {
		log.Fatal(err)
	}
}
