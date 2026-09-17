package main

import (
	"log"

	"github.com/HZreal/AI-scenery/demos/02-prompt-context/src/gin_prompt_context/internal/api"
	"github.com/HZreal/AI-scenery/demos/02-prompt-context/src/gin_prompt_context/internal/config"
	"github.com/HZreal/AI-scenery/demos/02-prompt-context/src/gin_prompt_context/internal/prompt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	generator, err := prompt.NewGeminiGenerator(cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		log.Fatal(err)
	}

	server := api.NewServer(generator, cfg.MaxContextChars)
	log.Printf("Serving prompt context demo on http://%s", cfg.Address)
	if err := server.Run(cfg.Address); err != nil {
		log.Fatal(err)
	}
}
