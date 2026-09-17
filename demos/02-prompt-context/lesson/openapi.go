package promptcontext

import "github.com/gin-gonic/gin"

var OpenAPISpec = gin.H{
	"openapi": "3.0.3",
	"info":    gin.H{"title": "AI Scenery Prompt Context API", "version": "1.0.0", "description": "Prompt 与上下文工程学习 Demo。"},
	"paths": gin.H{
		"/health": gin.H{"get": gin.H{"summary": "检查服务状态", "responses": gin.H{"200": gin.H{"description": "服务正常"}}}},
		"/api/prompt-context": gin.H{"post": gin.H{
			"summary": "比较多个 Prompt 版本", "requestBody": gin.H{"required": true, "content": gin.H{"application/json": gin.H{"schema": gin.H{"$ref": "#/components/schemas/PromptCompareRequest"}}}},
			"responses": gin.H{"200": gin.H{"description": "Prompt 对比结果"}, "400": gin.H{"description": "请求错误"}, "502": gin.H{"description": "Gemini 调用错误"}},
		}},
	},
	"components": gin.H{"schemas": gin.H{
		"PromptCompareRequest": gin.H{"type": "object", "required": []string{"task"}, "properties": gin.H{
			"task":            gin.H{"type": "string", "description": "需要完成的任务"},
			"context":         gin.H{"type": "string", "description": "任务参考资料"},
			"prompt_versions": gin.H{"type": "array", "items": gin.H{"type": "string", "enum": []string{"baseline-v1", "grounded-v2"}}},
		}},
	}},
}
