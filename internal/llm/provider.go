package llm

import (
	"context"
)

type Provider interface {
	Generate(context.Context, Request) (Result, error)
	Stream(context.Context, Request, func(string) error) (Result, error)
}

type Settings struct {
	Name          string
	GeminiAPIKey  string
	GeminiModel   string
	OpenAIAPIKey  string
	OpenAIModel   string
	OpenAIBaseURL string
}
