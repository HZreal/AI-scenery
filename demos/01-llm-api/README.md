# 01. LLM 后端接口

> 阶段状态：已完成。下一阶段进入 [Prompt 与上下文工程](../../notes/02-prompt-context.md)。

## 学习目标

把一次 LLM 调用封装成稳定、可观察、可测试的后端能力：消息结构、上下文预算、普通/流式响应、结构化输出、usage、成本、错误 trace 与 OpenAPI。

## 本阶段产出

- Python/Flask 服务：`src/llm_api/`，支持 `mock` 与 OpenAI Responses Provider。
- Go lesson：`lesson/`，通过根级 Provider 工厂支持 `mock`、Gemini、OpenAI。
- 统一服务接口：`POST /api/llm/chat`，普通响应、结构化 JSON 与 SSE 都采用稳定的应用层响应协议。
- API 契约：统一服务提供 `/openapi.json` 和 `/docs`。
- 浏览器实践：根级 `web/` 统一展示所有阶段的普通响应、SSE 与结构化数据。
- 学习讲解：[第一阶段后端实践讲解](../../notes/01-llm-api-backend-practice.md)。

## 技术选型

Python 版本使用 Flask 和标准库 HTTP，方便拆开理解 Web 路由、请求结构、响应结构与错误边界。Go 版本使用 Gin、Google GenAI SDK 和 OpenAI Go SDK，侧重 Provider 接口、工厂模式与多模型适配。

默认使用 `mock` provider，方便没有 API key 时也能跑通接口。你稍后在项目根目录的 `.env.local` 中填入真实 `OPENAI_API_KEY` 后，可切换为 `openai` provider。

## 核心概念

- **消息输入**：主输入是带 `system`、`user`、`assistant` 角色的 `messages` 数组；`message` 仍可作为单条 user 消息简写。
- **上下文预算**：`MAX_INPUT_CHARS` 限制输入字符数；它是应用层保护，不等同模型 token 上下文窗口。
- **模式选择**：`mode=text` 返回文本，`mode=json` 返回 JSON 对象。
- **元数据**：成功响应包含 model、latency、输入字符数、usage、成本估算和 trace。
- **流式响应**：`stream=true` 返回 SSE 的 metadata、delta、completed 事件。
- **错误边界**：错误响应包含 code 和 trace id。
- **OpenAPI**：`/openapi.json` 提供规范，`/docs` 提供 Swagger UI。
- **Provider 适配**：模型厂商的消息格式、结构化输出和流式细节封装在 Provider 内，HTTP 调用方不依赖某个 SDK。

## 接口概览

```http
POST /api/llm/chat
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

Python/Flask 版本仍可独立启动，默认监听 `http://127.0.0.1:8001`：

```bash
PYTHONPATH=demos/01-llm-api/src uv run python -m llm_api.server
```

Go 版与阶段 2 共用统一服务：

```bash
AI_SCENERY_PROVIDER=mock go run ./cmd/ai-scenery
```

服务默认监听 `http://127.0.0.1:8002`。在浏览器打开 `http://127.0.0.1:8002/web/`，可视化查看普通响应、SSE 与元数据。

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

浏览器会打开 Flask 版 Swagger UI。展开 `POST /api/chat`，点击 `Try it out`，填写请求体后即可直接发起请求；它是调试接口最方便的入口。统一 Go 服务则在 `http://127.0.0.1:8002/docs` 提供 `POST /api/llm/chat`。

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

- [x] 支持普通文本、结构化 JSON 与 SSE 流式输出。
- [x] 支持 `system`、`user`、`assistant` 消息与输入预算。
- [x] 正常和错误响应均包含可关联的 `trace_id`。
- [x] 提供 OpenAPI、Swagger UI、请求/响应样例和最小契约测试。
- [x] Python 接入 OpenAI；Go 接入 Gemini 与 OpenAI，并以 mock 支持无密钥本地验证。
- [x] Gemini JSON 输出通过 `ResponseMIMEType + ResponseSchema` 约束为对象结构。

## 补充说明

当前浏览器聊天页的会话历史只保存在页面内存；刷新页面后会丢失，后端也没有持久化 `session_id`。这是有意保留给后续“状态、记忆与会话”阶段的边界。

本阶段还不是 Agent：没有工具调用、RAG、任务循环、长期记忆、审批或多 Agent 编排。下一阶段会先解决“如何设计、版本化和裁剪 Prompt 与上下文”，再逐步把它们组合进 Agent。

Go 的共享 Provider、统一入口与页面见仓库根目录 [README](../../README.md)。
