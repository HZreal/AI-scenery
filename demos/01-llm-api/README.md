# 01. LLM API

## Goal

学习 LLM 后端接口的最小封装：普通响应、流式响应、结构化输出、错误处理和 trace。

## Suggested Tech Stack

默认优先 Python；如果后续要接 Web 前端，可用 TypeScript；如果要做后端工程化接口，可用 Go。

## API Sketch

```http
POST /api/chat
Content-Type: application/json
```

```json
{
  "message": "解释什么是 AI Agent",
  "mode": "text"
}
```

## Acceptance

- 支持普通文本输出。
- 支持结构化 JSON 输出。
- 支持流式输出。
- 错误响应包含 trace id。

