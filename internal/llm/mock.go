package llm

import (
	"context"
	"strings"
)

type mockProvider struct{}

func newMockProvider(Settings) (Provider, error) { return mockProvider{}, nil }

func (mockProvider) Generate(_ context.Context, request Request) (Result, error) {
	if request.Mode == ModeJSON {
		return Result{Model: "mock-gemini", Value: map[string]any{
			"topic":      "Gin LLM API",
			"summary":    "这是本地 mock 响应，用于验证 Go API 契约。",
			"key_points": []string{"Provider 工厂", "Gemini 适配器", "OpenAPI"},
		}}, nil
	}
	return Result{Model: "mock-gemini", Value: "AI Agent 可以理解目标、调用工具并返回结果。你问的是：" + request.Messages[len(request.Messages)-1].Content}, nil
}

func (m mockProvider) Stream(ctx context.Context, request Request, emit func(string) error) (Result, error) {
	result, err := m.Generate(ctx, Request{Messages: request.Messages, Mode: ModeText})
	if err != nil {
		return Result{}, err
	}
	for _, word := range strings.Fields(result.Value.(string)) {
		if err := emit(word + " "); err != nil {
			return Result{}, err
		}
	}
	return result, nil
}
