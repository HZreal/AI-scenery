package promptcontext

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/HZreal/AI-scenery/internal/llm"
)

func (testProvider) Generate(_ context.Context, _ llm.Request) (llm.Result, error) {
	return llm.Result{Value: "测试结果", Model: "test-gemini"}, nil
}

func (testProvider) Stream(context.Context, llm.Request, func(string) error) (llm.Result, error) {
	return llm.Result{}, nil
}

type testProvider struct{}

func TestCompareReturnsBothPromptVersions(t *testing.T) {
	router := gin.New()
	Register(router.Group("/api/prompt-context"), testProvider{}, 100)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/prompt-context", strings.NewReader(`{"task":"解释 Prompt"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"baseline-v1"`) || !strings.Contains(recorder.Body.String(), `"grounded-v2"`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestCompareRejectsEmptyTask(t *testing.T) {
	router := gin.New()
	Register(router.Group("/api/prompt-context"), testProvider{}, 100)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/prompt-context", strings.NewReader(`{"task":" "}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"empty_task"`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
