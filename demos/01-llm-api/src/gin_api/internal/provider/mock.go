package provider

import (
	"context"
	"strings"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/chat"
)

type mockProvider struct{}

func newMockProvider(Settings) (Provider, error) { return mockProvider{}, nil }

func (mockProvider) Generate(_ context.Context, request chat.Request) (chat.Result, error) {
	if request.Mode == chat.ModeJSON {
		return chat.Result{Model: "mock-gemini", Value: map[string]any{
			"topic":      "Gin LLM API",
			"summary":    "这是本地 mock 响应，用于验证 Go API 契约。",
			"key_points": []string{"Provider 工厂", "Gemini 适配器", "OpenAPI"},
		}}, nil
	}
	return chat.Result{Model: "mock-gemini", Value: "AI Agent 可以理解目标、调用工具并返回结果。你问的是：" + request.Messages[len(request.Messages)-1].Content}, nil
}

func (m mockProvider) Stream(ctx context.Context, request chat.Request, emit func(string) error) (chat.Result, error) {
	result, err := m.Generate(ctx, chat.Request{Messages: request.Messages, Mode: chat.ModeText})
	if err != nil {
		return chat.Result{}, err
	}
	for _, word := range strings.Fields(result.Value.(string)) {
		if err := emit(word + " "); err != nil {
			return chat.Result{}, err
		}
	}
	return result, nil
}
