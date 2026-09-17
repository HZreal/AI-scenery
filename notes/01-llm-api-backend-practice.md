# 第一阶段：LLM 后端 API 实践讲解

本阶段的目标不是研究模型训练，而是把一次 LLM 调用做成一个可被其他系统稳定调用、可观测、可测试的后端服务。

对应 Demo 位于 `demos/01-llm-api/`。启动后可访问：

- `POST /api/chat`：聊天与生成接口。
- `GET /health`：健康检查。
- `GET /openapi.json`：机器可读的 OpenAPI 规范。
- `GET /docs`：Swagger UI 接口文档和试调页面。

## 1. 这一阶段学什么

从后端开发角度，第一阶段要建立下面这些基础认识：

1. **LLM 是外部依赖，不是普通函数。** 调用会有网络延迟、超时、限流、返回格式变化和按 token 计费，因此要有边界层，而不是在业务代码里到处直接请求模型。
2. **请求需要明确契约。** API 不应只接受一段模糊字符串，而应定义消息、角色、输出模式、流式语义和错误形状。
3. **上下文是有限资源。** 每多带一段历史消息，都会增加输入量、延迟和费用；服务必须有明确预算。
4. **自然语言与结构化数据是两种输出。** 给人看的结果可以是文本；给程序继续处理的结果应该有 JSON Schema 约束。
5. **流式是传输协议设计。** 它不是简单把字符串切开，而是服务端持续向客户端发送状态、增量和完成事件。
6. **可观测性是接口的一部分。** trace id、模型、耗时、usage 和成本字段能让一次调用可复盘。
7. **OpenAPI 是后端契约的说明书。** 它让接口文档、试调页面和后续客户端生成有统一来源。

## 2. 总体设计

这个 Demo 分为三层：

```text
HTTP / OpenAPI 层
  Flask + flask-smorest
        |
服务层
  请求归一化、上下文预算、metadata、错误模型
        |
Provider 层
  MockLLMProvider / OpenAIResponsesProvider
```

### HTTP / OpenAPI 层

`server.py` 负责 HTTP 路由，不关心具体模型如何调用。`flask-smorest` 的 Schema 同时用于请求校验和生成 OpenAPI，因此接口文档不需要另写一份容易过期的 JSON 文件。

### 服务层

`service.py` 定义核心对象：

- `ChatMessage`：一条带角色的消息。
- `ChatRequest`：归一化后的请求，包含消息数组和输出模式。
- `ProviderResult`：Provider 返回的结果、模型名和 token usage。
- `ChatError`：带业务错误码与 HTTP 状态码的异常。

这一层负责把 HTTP 输入变成领域对象、限制输入规模、生成 trace，以及统一响应 metadata。

### Provider 层

`LLMProvider` 是抽象协议，只有两个能力：`generate()` 返回完整结果，`stream()` 返回文本增量。

`MockLLMProvider` 不需要 Key，用于本地开发和自动化测试。`OpenAIResponsesProvider` 使用标准库 HTTP 请求 Responses API。上层只依赖协议，因此未来接入其他模型服务时，不需要重写路由和业务契约。

## 3. 请求如何进入系统

推荐请求形状：

```json
{
  "messages": [
    {"role": "system", "content": "回答要简洁。"},
    {"role": "user", "content": "解释什么是 AI Agent。"}
  ],
  "mode": "text",
  "stream": false
}
```

### 消息与角色

`system` 放稳定规则，例如回答风格、任务边界。`user` 表示本次需求。`assistant` 可放进前序回答，构成短期多轮上下文。

旧的 `message` 字段仍被接受，但它只是单条 user 消息的简写：

```json
{"message": "解释 AI Agent"}
```

服务会把它转换成：

```json
{"messages": [{"role": "user", "content": "解释 AI Agent"}]}
```

不能同时传 `message` 与 `messages`，避免“哪一个才是真的输入”这种歧义。

## 4. 上下文预算与上下文窗口

模型有自己的 token 上下文窗口，但不同模型、版本和配置的限制不同。本 Demo 不假装精确计算模型 token，而是在应用层实现 `MAX_INPUT_CHARS` 字符预算，默认值为 `12000`。

流程是：

1. 把所有 `messages.content` 的字符数求和。
2. 与 `MAX_INPUT_CHARS` 比较。
3. 超限时返回 `context_limit_exceeded`，不静默截断。

这样做的价值是可预测：调用方知道请求没有被悄悄改写。真实生产服务通常还会结合模型 token 计数器、摘要、检索和历史裁剪策略；这些会在 RAG、记忆和 Agent 阶段继续学习。

## 5. 普通输出与结构化输出

`mode=text` 返回文本：

```json
{
  "result": "AI Agent 可以围绕目标理解任务、调用工具并返回结果。",
  "metadata": {"trace_id": "trace-..."}
}
```

`mode=json` 则要求 Provider 按 JSON Schema 生成对象，当前 schema 包含 `topic`、`summary`、`key_points`。服务端把结果解析为真正的 JSON 对象，而不是把 JSON 再塞入字符串。

这一区别很重要：文本适合人阅读；JSON 对象适合后续程序做字段访问、存库、校验或继续交给工具调用。

OpenAI Provider 会在 Responses 请求中提交 `text.format` 和严格 JSON Schema。即使模型被要求输出 JSON，服务端仍会 `json.loads` 校验；“提示词说要 JSON”不等于后端可以放弃验证。

## 6. Provider 调用与 usage

非流式调用路径如下：

```text
POST /api/chat
  -> ChatInputSchema 校验字段
  -> request_from_payload 归一化消息
  -> create_chat_response 检查上下文预算
  -> provider.generate
  -> 组装 result + metadata
```

OpenAI Responses 的原始 REST 响应中，输出文本位于 `output[].content[]`。代码只提取 `type=output_text` 的内容；这比依赖 SDK 专用的便捷字段更适合当前的标准库 HTTP 实现。

metadata 包含：

- `model`：实际 Provider 选择的模型名。
- `latency_ms`：本次 Provider 调用耗时。
- `input_characters`：应用层输入规模。
- `usage.input_tokens`、`usage.output_tokens`：Provider 返回时记录；Mock 为 `null`，不伪造数据。
- `cost_estimate_usd`：只有配置输入和输出每百万 token 单价时才计算；没有可靠单价时返回 `null`。
- `trace_id`：一次请求的关联标识。

成本单价通过环境变量配置，而不是硬编码在代码中，因为模型价格会变化。

## 7. SSE 流式响应

当 `stream=true` 时，接口使用 `text/event-stream`。本服务定义自己的稳定事件协议：

```text
event: metadata
data: {"trace_id":"...","model":"mock-llm"}

event: delta
data: {"delta":"AI "}

event: completed
data: {"metadata":{...}}
```

后端不直接把 OpenAI 的事件格式透传给客户端。`OpenAIResponsesProvider` 只读取上游的 `response.output_text.delta`，再转换成自己的 `delta` 事件。这样未来换模型服务时，前端或调用方不必改变。

Mock Provider 按词产生增量，只用于验证 SSE 协议和客户端处理逻辑；它不代表模型真实生成速度。

## 8. 错误、trace 与 OpenAPI

服务层用 `ChatError` 表达可预期的错误，例如：

- `empty_message`：消息为空。
- `ambiguous_input`：同时传了 `message` 与 `messages`。
- `context_limit_exceeded`：输入超过应用预算。
- `missing_api_key`：选择 OpenAI Provider 但未配置 Key。
- `provider_network_error`：调用模型服务时网络失败。

错误响应统一带 `error`、`code` 与 `metadata.trace_id`。这使调用方可以按 code 处理，而不是解析自然语言错误文本；也可以用 trace id 查日志或关联调用链。

OpenAPI 的作用不是“多一个页面”，而是把这些字段、状态码与请求 schema 固化成 API 契约。`/openapi.json` 给程序读取，`/docs` 给开发者试调；两者来自同一套 Flask Schema。

## 9. 如何调用与验证

启动：

```bash
PYTHONPATH=demos/01-llm-api/src uv run python -m llm_api.server
```

调用结构化输出：

```bash
curl -s http://127.0.0.1:8001/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"列出 LLM API 的学习点"}],"mode":"json"}'
```

查看契约：

```bash
open http://127.0.0.1:8001/docs
curl http://127.0.0.1:8001/openapi.json
```

运行测试：

```bash
npm test
```

当前测试只覆盖高价值契约：消息简写、上下文超限、JSON 对象、OpenAPI 路径、SSE 完成事件、错误 trace，以及原始 OpenAI 响应文本解析。

## 10. 本阶段完成了什么，下一阶段学什么

本阶段已经完成“单次 LLM 后端调用”的工程骨架。它能接收结构化输入，选择 Provider，控制输入大小，处理普通/结构化/流式输出，并提供可观测 metadata 和 OpenAPI 契约。

它还不是 Agent：没有工具调用、任务规划、长期记忆、RAG、人工审批或多 Agent 协作。这些分别属于后续阶段。下一阶段的 Prompt 与上下文工程，会在当前 `messages`、system 规则和上下文预算的基础上，学习如何让输入内容更稳定、更可控。
