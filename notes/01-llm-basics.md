# 01. LLM 应用基础

## 学习目标

理解模型 API 调用的基本结构，并能封装一个稳定的后端接口。

## 核心要点

- 模型、消息、角色和上下文窗口。
- 普通响应与流式响应。
- 结构化输出与 JSON schema。
- 延迟、token 消耗和成本估算。
- API 错误、超时和重试。

## Demo 方向

`demos/01-llm-api/`

实现一个 `/api/chat` 或 `/api/generate` 接口，支持普通响应、流式响应和 JSON 输出。

## 验证场景

- 正常问题能返回文本结果。
- 指定 JSON 输出时，结果能被解析。
- API 失败或超时时，返回清晰错误和 trace id。


## 本阶段结论

一个可维护的 LLM 后端接口，不应只接收一段 prompt 再返回一段文本。最小完整闭环包括：消息角色、输入预算、普通与结构化输出、SSE、可观察元数据、错误 trace 和可机器读取的 API 契约。

## 消息与上下文

`messages` 以 `system`、`user`、`assistant` 保留对话角色。system 用于稳定任务边界，user 表达当前请求，assistant 可携带前序回答。Demo 的 `MAX_INPUT_CHARS` 是应用层输入预算：超限时拒绝请求，避免无意把长文本送入模型；它不等于模型真实 token 上下文窗口。

## 输出与流式

普通模式输出文本；结构化模式要求模型满足 JSON Schema，服务将其作为 JSON 对象返回。SSE 依次发出 metadata、delta、completed，客户端只依赖这套稳定事件名。Mock 用于学习传输格式；OpenAI Provider 会把上游文本 delta 转换为本服务事件。

## 延迟、usage、成本与错误

每次响应记录模型、耗时、输入字符数、usage 和 trace。OpenAI 返回 usage 后，只有显式配置输入/输出每百万 token 单价时才计算成本；不把会变化的价格硬编码进代码。错误统一提供 code、message 和 trace id，方便定位问题。

## OpenAPI

OpenAPI 是接口的标准说明书。`/openapi.json` 供程序读取、生成客户端或做契约检查；`/docs` 是同一份规范的浏览器界面，可直接试调接口。Schema 与 Flask 路由共用定义，避免文档和代码分叉。

完整的后端实践讲解见 [01-llm-api-backend-practice.md](01-llm-api-backend-practice.md)。
