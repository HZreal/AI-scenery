# 01. LLM 后端接口

## 学习目标

学习 LLM 后端接口的最小封装：消息结构、上下文预算、普通/流式响应、结构化输出、usage、成本、错误 trace 与 OpenAPI。

## 技术选型

本 demo 使用 Flask 实现 HTTP 层；模型调用仍使用 Python 标准库，便于分别理解 Web 路由、请求结构、响应结构和错误边界。

默认使用 `mock` provider，方便没有 API key 时也能跑通接口。你稍后在项目根目录的 `.env.local` 中填入真实 `OPENAI_API_KEY` 后，可切换为 `openai` provider。

## 核心概念

- **消息输入**：主输入是带 `system`、`user`、`assistant` 角色的 `messages` 数组；`message` 仍可作为单条 user 消息简写。
- **上下文预算**：`MAX_INPUT_CHARS` 限制输入字符数；它是应用层保护，不等同模型 token 上下文窗口。
- **模式选择**：`mode=text` 返回文本，`mode=json` 返回 JSON 对象。
- **元数据**：成功响应包含 model、latency、输入字符数、usage、成本估算和 trace。
- **流式响应**：`stream=true` 返回 SSE 的 metadata、delta、completed 事件。
- **错误边界**：错误响应包含 code 和 trace id。
- **OpenAPI**：`/openapi.json` 提供规范，`/docs` 提供 Swagger UI。

## 接口概览

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

## 本地环境变量

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

## 启动服务

从项目根目录启动：

```bash
PYTHONPATH=demos/01-llm-api/src uv run python -m llm_api.server
```

服务默认监听 `http://127.0.0.1:8001`。保持这个终端运行，再打开另一个终端验证接口。

## 验证接口

### 1. 检查服务是否存活

```bash
curl http://127.0.0.1:8001/health
```

预期返回 `status: ok`，表示 Flask 服务已启动。

### 2. 在浏览器中调用 OpenAPI 文档

```bash
open http://127.0.0.1:8001/docs
```

浏览器会打开 Swagger UI。展开 `POST /api/chat`，点击 `Try it out`，填写请求体后即可直接发起请求；它是调试接口最方便的入口。

也可以直接查看 OpenAPI 描述文件：

```bash
curl -s http://127.0.0.1:8001/openapi.json | uv run python -m json.tool
```

### 3. 验证普通文本响应

```bash
curl -X POST http://127.0.0.1:8001/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"用一句话解释什么是上下文窗口。"}'
```

预期响应包含 `result` 和 `metadata`；`metadata` 中会有 `model`、`latency_ms`、`cost_estimate`、`trace_id` 等字段。

### 4. 验证结构化 JSON 输出

```bash
curl -X POST http://127.0.0.1:8001/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"system","content":"你是学习助手。"},{"role":"user","content":"列出两个 Agent 开发学习点。"}],"mode":"json"}'
```

预期 `result` 是 JSON 对象，而不是需要再次解析的 JSON 字符串。

### 5. 验证流式响应

```bash
curl -N -X POST http://127.0.0.1:8001/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"流式响应有什么用？","stream":true}'
```

预期按顺序看到 `metadata`、多个 `delta`，最后是 `completed` 事件。这是服务自己的 SSE 协议，调用方不需要依赖模型厂商的原始事件格式。

### 6. 验证错误输入

```bash
curl -i -X POST http://127.0.0.1:8001/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"   "}'
```

预期返回 `400`，响应中包含明确的 `code`、`error` 和可用于排查的 `metadata.trace_id`。

## 运行测试

在仓库根目录运行完整的基础验证：

```bash
npm test
```

只运行本 Demo 的 Python 测试：

```bash
PYTHONPATH=demos/01-llm-api/src uv run python -m unittest discover -s demos/01-llm-api/tests -v
```

## 完成标准

- 支持普通文本输出。
- 支持结构化 JSON 输出。
- 支持流式输出。
- 错误响应包含 trace id。

## 补充说明

这个 demo 使用 Flask 是为了学习 Python Web API 的路由、测试客户端和 SSE 响应；仍然没有引入 OpenAI SDK，目的是把“后端 API 如何包装 LLM 调用”这件事拆清楚。后续可以用 Gin 或 Express 实现同一接口作横向比较。
