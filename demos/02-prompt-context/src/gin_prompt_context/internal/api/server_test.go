package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HZreal/AI-scenery/demos/02-prompt-context/src/gin_prompt_context/internal/prompt"
)

type testGenerator struct{}

func (testGenerator) Generate(_ context.Context, _, _ string) (prompt.Generation, error) {
	return prompt.Generation{Text: "测试结果", Model: "test-gemini"}, nil
}

func TestCompareReturnsBothPromptVersions(t *testing.T) {
	router := NewServer(testGenerator{}, 100)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/prompt-context", strings.NewReader(`{"task":"解释 Prompt"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"baseline-v1"`) || !strings.Contains(recorder.Body.String(), `"grounded-v2"`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestCompareRejectsEmptyTask(t *testing.T) {
	router := NewServer(testGenerator{}, 100)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/prompt-context", strings.NewReader(`{"task":" "}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"empty_task"`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
