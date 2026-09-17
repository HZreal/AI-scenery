package llmapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/HZreal/AI-scenery/internal/llm"
)

type testProvider struct{}

func (testProvider) Generate(_ context.Context, _ llm.Request) (llm.Result, error) {
	return llm.Result{Value: "测试响应", Model: "test-model"}, nil
}

func (testProvider) Stream(_ context.Context, _ llm.Request, emit func(string) error) (llm.Result, error) {
	if err := emit("测试 "); err != nil {
		return llm.Result{}, err
	}
	return llm.Result{Value: "测试", Model: "test-model"}, nil
}

func testRouter() *gin.Engine {
	router := gin.New()
	Register(router.Group("/api/llm"), testProvider{}, 100)
	return router
}

func TestChatReturnsStableResponseShape(t *testing.T) {
	router := testRouter()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/llm/chat", strings.NewReader(`{"message":"什么是 Agent？"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"result":"测试响应"`) || !strings.Contains(recorder.Body.String(), `"trace_id"`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestChatRejectsEmptyMessage(t *testing.T) {
	router := testRouter()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/llm/chat", strings.NewReader(`{"message":"  "}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"code":"empty_message"`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestOpenAPISpecContainsChatRoute(t *testing.T) {
	if OpenAPISpec["openapi"] != "3.0.3" || !strings.Contains(OpenAPISpec["paths"].(gin.H)["/api/llm/chat"].(gin.H)["post"].(gin.H)["summary"].(string), "生成") {
		t.Fatalf("unexpected OpenAPI spec: %#v", OpenAPISpec)
	}
}
