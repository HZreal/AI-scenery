# 01. LLM API

## Goal

学习 LLM 后端接口的最小封装：消息结构、上下文预算、普通/流式响应、结构化输出、usage、成本、错误 trace 与 OpenAPI。

## Suggested Tech Stack

本 demo 使用 Flask 实现 HTTP 层；模型调用仍使用 Python 标准库，便于分别理解 Web 路由、请求结构、响应结构和错误边界。

默认使用 `mock` provider，方便没有 API key 时也能跑通接口。你稍后在项目根目录的 `.env.local` 中填入真实 `OPENAI_API_KEY` 后，可切换为 `openai` provider。

## Core Concepts

- **消息输入**：主输入是带 `system`、`user`、`assistant` 角色的 `messages` 数组；`message` 仍可作为单条 user 消息简写。
- **上下文预算**：`MAX_INPUT_CHARS` 限制输入字符数；它是应用层保护，不等同模型 token 上下文窗口。
- **模式选择**：`mode=text` 返回文本，`mode=json` 返回 JSON 对象。
- **元数据**：成功响应包含 model、latency、输入字符数、usage、成本估算和 trace。
- **流式响应**：`stream=true` 返回 SSE 的 metadata、delta、completed 事件。
- **错误边界**：错误响应包含 code 和 trace id。
- **OpenAPI**：`/openapi.json` 提供规范，`/docs` 提供 Swagger UI。

## API Sketch

```http
POST /api/chat
Content-Type: application/json
```

```json
{
  "messages": [{"role": "user", "content": "解释什么是 AI Agent"}],
  "mode": "text"
}
```

## Local Env

项目根目录的 `.env.local` 会被 Git 忽略。先使用 mock：

```bash
AI_SCENERY_PROVIDER=mock
OPENAI_MODEL=gpt-5.5
OPENAI_API_KEY=
```

填入真实 key 后切换 OpenAI：

```bash
AI_SCENERY_PROVIDER=openai
OPENAI_MODEL=gpt-5.5
OPENAI_API_KEY=your-api-key
```

不要提交 `.env.local`。

## Run

从项目根目录启动：

```bash
PYTHONPATH=demos/01-llm-api/src uv run python -m llm_api.server
```

健康检查：

```bash
curl http://127.0.0.1:8001/health
curl http://127.0.0.1:8001/openapi.json
```

普通文本：

```bash
curl -s http://127.0.0.1:8001/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"解释什么是 AI Agent"}],"mode":"text"}'
```

结构化 JSON 字符串：

```bash
curl -s http://127.0.0.1:8001/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"解释什么是 AI Agent"}],"mode":"json"}'
```

流式输出：

```bash
curl -N http://127.0.0.1:8001/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"解释什么是 AI Agent"}],"mode":"text","stream":true}'
```

## Tests

```bash
PYTHONPATH=demos/01-llm-api/src uv run python -m unittest discover -s demos/01-llm-api/tests -v
```

## Acceptance

- 支持普通文本输出。
- 支持结构化 JSON 输出。
- 支持流式输出。
- 错误响应包含 trace id。

## Notes

这个 demo 使用 Flask 是为了学习 Python Web API 的路由、测试客户端和 SSE 响应；仍然没有引入 OpenAI SDK，目的是把“后端 API 如何包装 LLM 调用”这件事拆清楚。后续可以用 Gin 或 Express 实现同一接口作横向比较。
