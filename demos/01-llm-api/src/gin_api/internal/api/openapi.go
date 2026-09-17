package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OpenAPI 契约集中在这里，便于学习时对照接口实现与客户端调用。
var openAPISpec = gin.H{
	"openapi": "3.0.3",
	"info":    gin.H{"title": "AI Scenery Gin LLM API", "version": "1.0.0", "description": "Go/Gin 版 LLM API 学习 Demo。"},
	"paths": gin.H{
		"/health": gin.H{"get": gin.H{"summary": "检查服务状态", "responses": gin.H{"200": gin.H{"description": "服务正常"}}}},
		"/api/chat": gin.H{"post": gin.H{
			"summary": "生成 LLM 响应", "requestBody": gin.H{"required": true, "content": gin.H{"application/json": gin.H{"schema": gin.H{"$ref": "#/components/schemas/ChatRequest"}}}},
			"responses": gin.H{"200": gin.H{"description": "普通 JSON 响应或 SSE 流"}, "400": gin.H{"description": "请求错误"}, "500": gin.H{"description": "服务错误"}},
		}},
	},
	"components": gin.H{"schemas": gin.H{
		"Message":     gin.H{"type": "object", "required": []string{"role", "content"}, "properties": gin.H{"role": gin.H{"type": "string", "enum": []string{"system", "user", "assistant"}}, "content": gin.H{"type": "string"}}},
		"ChatRequest": gin.H{"type": "object", "properties": gin.H{"messages": gin.H{"type": "array", "items": gin.H{"$ref": "#/components/schemas/Message"}}, "message": gin.H{"type": "string"}, "mode": gin.H{"type": "string", "enum": []string{"text", "json"}, "default": "text"}, "stream": gin.H{"type": "boolean", "default": false}}},
	}},
}

func swaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><title>Gin LLM API 文档</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:"/openapi.json",dom_id:"#swagger-ui"})</script></body></html>`))
}
