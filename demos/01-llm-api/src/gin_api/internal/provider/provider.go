package provider

import (
	"context"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/chat"
)

type Provider interface {
	Generate(context.Context, chat.Request) (chat.Result, error)
	Stream(context.Context, chat.Request, func(string) error) (chat.Result, error)
}

type Settings struct {
	Name          string
	GeminiAPIKey  string
	GeminiModel   string
	OpenAIAPIKey  string
	OpenAIModel   string
	OpenAIBaseURL string
}
