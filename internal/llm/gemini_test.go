package llm

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/genai"
)

func TestFactoryRejectsGeminiWithoutAPIKey(t *testing.T) {
	_, err := NewFactory().Create(Settings{Name: "gemini", GeminiModel: "gemini-3.8-flash"})
	if err == nil {
		t.Fatal("expected missing API key error")
	}
}

func TestOpenAIParamsPreserveMessageRolesAndJSONSchema(t *testing.T) {
	request, err := NewRequest([]Message{
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
	request, err := NewRequest([]Message{
		{Role: "system", Content: "你是学习助手。"},
		{Role: "user", Content: "你好"},
		{Role: "assistant", Content: "你好，我能帮助你。"},
	}, "", "json")
	if err != nil {
		t.Fatal(err)
	}

	contents, options := geminiInput(request)
	if options.SystemInstruction == nil || options.ResponseMIMEType != "application/json" || options.ResponseSchema == nil {
		t.Fatal("expected system instruction and structured JSON configuration")
	}
	if options.ResponseSchema.Type != genai.TypeObject || len(options.ResponseSchema.Required) != 3 {
		t.Fatalf("unexpected Gemini JSON schema: %#v", options.ResponseSchema)
	}
	if len(contents) != 2 || contents[0].Role != "user" || contents[1].Role != "model" {
		t.Fatalf("unexpected Gemini role mapping: %#v", contents)
	}
}
