# LLM API 第一阶段设计

**目标：** 将 `demos/01-llm-api` 完成为具备 Flask、OpenAPI、消息输入、上下文边界、结构化输出、流式输出、用量/成本元数据和可测试错误契约的 LLM 后端 Demo。

**范围：** 只完成第一阶段 LLM 应用基础；不加入工具调用、RAG、Agent Loop、会话持久化或真实密钥。

## 架构

HTTP 层使用 Flask 与 `flask-smorest`。Marshmallow Schema 同时负责请求校验、响应序列化和 OpenAPI 生成，避免维护一份与代码分离的接口说明。

服务层继续通过 `LLMProvider` 屏蔽模型厂商差异。`MockLLMProvider` 用于全部自动化验证；`OpenAIResponsesProvider` 支持真实调用和上游 SSE 事件转发，但本阶段不使用真实 Key 做测试。

```text
Flask + flask-smorest
  -> 请求 Schema / OpenAPI
  -> ChatRequest 与上下文校验
  -> LLMProvider（mock 或 openai）
  -> JSON 或 SSE 响应
```

## API 契约

保留 `POST /api/chat` 与 `GET /health`，新增：

- `GET /openapi.json`：OpenAPI 3.0.3 规范。
- `GET /docs`：Swagger UI。

`POST /api/chat` 的主输入为：

```json
{
  "messages": [
    {"role": "system", "content": "你是简洁的学习助手。"},
    {"role": "user", "content": "解释什么是 AI Agent。"}
  ],
  "mode": "text",
  "stream": false
}
```

`messages` 支持 `system`、`user`、`assistant`。为兼容已有样例，单个 `message` 可作为 `messages` 的简写；请求不能同时提供两者，也不能两者皆缺。

普通响应返回文本 `result`；`mode=json` 返回 JSON 对象 `result`。成功响应均带 `metadata`：模型、耗时、trace、输入字符数、token usage 与成本。成本单价未配置时返回 `null`，不得写死会过期的模型价格。

错误响应统一为 `error`、`code`、`metadata.trace_id`。客户端错误使用 400，上游/Provider 错误使用 500。

## 上下文、流式与用量

服务对拼接后的输入执行 `MAX_INPUT_CHARS` 字符预算（默认 12000）。这是可控的应用层输入限制，不伪装成模型真实 token 上下文窗口；超限明确报错，不静默截断。

Provider 的非流式结果包含文本/JSON、usage 和模型名。OpenAI 的 usage 从 Responses 返回读取；Mock 不虚构 token 数。若设置输入/输出每百万 token 的美元单价，服务以 usage 计算成本，否则成本为 `null`。

流式响应使用 SSE：先发送 `metadata`，随后发送一个或多个 `delta`，最后发送 `completed`。Mock 按词生成 delta；OpenAI Provider 将上游 Responses SSE 的文本增量转换为同一事件格式。

## 文档与依赖

新增 `flask-smorest`，由 uv 锁定版本。Swagger UI 通过官方配置的 CDN 资源显示；即使浏览器无法加载 CDN，`/openapi.json` 仍可用。

扩写 `notes/01-llm-basics.md` 和 Demo README，说明消息角色、应用层上下文预算、结构化输出、SSE、usage/成本、OpenAPI 与 Key 的安全边界。更新 examples 的请求和响应形状。

## 测试与验收

这是学习 Demo，测试只覆盖关键公开契约和高风险分支，不为简单辅助函数追求覆盖率。

- Schema：消息简写、messages 角色、互斥校验、上下文超限。
- 服务：text、JSON 对象、usage/成本、Provider 错误携带 trace。
- 路由：健康检查、正常/错误响应、SSE 事件序列。
- OpenAPI：`/openapi.json` 包含 `/api/chat` 与请求/响应 schema；`/docs` 可返回页面。
- 回归：`npm test`、`uv lock --check`、`git diff --check`。

真实 OpenAI 请求不属于本阶段验收，因为它需要用户自行配置 `.env.local` 的 Key 且会产生外部费用。
