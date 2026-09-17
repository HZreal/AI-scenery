package provider

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/chat"
)

func TestFactoryRejectsGeminiWithoutAPIKey(t *testing.T) {
	_, err := NewFactory().Create(Settings{Name: "gemini", GeminiModel: "gemini-3.8-flash"})
	if err == nil {
		t.Fatal("expected missing API key error")
	}
}

func TestOpenAIParamsPreserveMessageRolesAndJSONSchema(t *testing.T) {
	request, err := chat.NewRequest([]chat.Message{
		{Role: "system", Content: "你是学习助手。"},
		{Role: "user", Content: "列出两个学习点。"},
		{Role: "assistant", Content: "已记录上下文。"},
	}, "", "json")
	if err != nil {
		t.Fatal(err)
	}

	body, err := json.Marshal(openAIParams(request, "gpt-5.5"))
	if err != nil {
		t.Fatal(err)
	}
	serialized := string(body)
	for _, expected := range []string{`"role":"system"`, `"role":"user"`, `"role":"assistant"`, `"type":"json_schema"`} {
		if !strings.Contains(serialized, expected) {
			t.Fatalf("missing %s in request: %s", expected, serialized)
		}
	}
}

func TestFactoryRejectsOpenAIWithoutAPIKey(t *testing.T) {
	_, err := NewFactory().Create(Settings{Name: "openai", OpenAIModel: "gpt-5.5"})
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
