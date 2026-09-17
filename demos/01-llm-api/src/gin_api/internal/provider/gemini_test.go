package provider

import (
	"testing"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/chat"
)

func TestFactoryRejectsGeminiWithoutAPIKey(t *testing.T) {
	_, err := NewFactory().Create(Settings{Name: "gemini", Model: "gemini-3.8-flash"})
	if err == nil {
		t.Fatal("expected missing API key error")
	}
}

func TestGeminiInputMapsSystemAndAssistantRoles(t *testing.T) {
	request, err := chat.NewRequest([]chat.Message{
		{Role: "system", Content: "你是学习助手。"},
		{Role: "user", Content: "你好"},
		{Role: "assistant", Content: "你好，我能帮助你。"},
	}, "", "json")
	if err != nil {
		t.Fatal(err)
	}

	contents, options := geminiInput(request)
	if options.SystemInstruction == nil || options.ResponseMIMEType != "application/json" {
		t.Fatal("expected system instruction and JSON mode configuration")
	}
	if len(contents) != 2 || contents[0].Role != "user" || contents[1].Role != "model" {
		t.Fatalf("unexpected Gemini role mapping: %#v", contents)
	}
}
