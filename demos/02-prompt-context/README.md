# 02. Prompt 与上下文工程

> 阶段状态：进行中。本 Demo 聚焦版本化 Prompt 与上下文组织，通过统一 Go/Gin 服务接入 Gemini。

## 学习目标

把 Prompt 当作可维护的后端配置，而不是散落在业务代码中的字符串。通过同一任务运行多个 Prompt 版本，观察指令结构、few-shot 示例和参考上下文如何改变输出。

## 本 Demo 做了什么

- 提供 `POST /api/prompt-context`，对同一 `task` 比较多个 Prompt 版本。
- 内置 `baseline-v1`：最小、明确的任务指令。
- 内置 `grounded-v2`：使用角色、规则、XML 边界、few-shot 示例和“只将 context 视为资料”的约束。
- 对 `context` 做应用层字符预算；超限时保留开头和结尾，并标记中间裁剪。
- 通过根级共享 Provider 调用 Gemini 或 OpenAI，记录每个版本的输出、延迟和 token 用量。
- 对 Gemini 服务暂时繁忙（`503`）最多尝试 3 次，并采用有上限的退避等待；配额耗尽会直接返回明确错误。
- 提供 `/openapi.json` 与 `/docs`，便于从 Swagger UI 直接试调。

## 重要设计

Gemini 的 `GenerateContent` 使用 `SystemInstruction`、`user` 与 `model` 三种表达。通用 LLM 架构中的 developer 规则在本 Demo 中作为 system instruction 内的 `<rules>` 区段管理；它仍由后端控制，不能由用户请求覆盖。

`grounded-v2` 的用户输入保持固定顺序：few-shot 示例、`<context>`、`<task>`、最终执行指令。任务放在参考资料之后，避免长上下文淹没真正的问题。

这是学习用对比 API，因此会返回 system instruction 和最终 user prompt 预览。生产接口通常不应回显内部规则、敏感上下文或完整 Prompt。

## 接口

```http
POST /api/prompt-context
Content-Type: application/json
```

```json
{
  "task": "根据资料说明为什么需要上下文裁剪。",
  "context": "输入上下文会影响延迟和成本。没有必要的信息不应发送给模型。",
  "prompt_versions": ["baseline-v1", "grounded-v2"]
}
```

省略 `prompt_versions` 时默认执行两个版本。响应的 `result.comparisons` 会包含每个版本实际发送的指令、输出与用量，便于人工比较。

## 配置与启动

在项目根目录的 `.env.local` 中设置：

```bash
AI_SCENERY_PROVIDER=gemini
GEMINI_API_KEY=your-api-key
GEMINI_MODEL=gemini-2.5-flash-lite
AI_SCENERY_ADDRESS=127.0.0.1:8002
PROMPT_CONTEXT_MAX_CHARS=6000
```

启动：

```bash
go run ./cmd/ai-scenery
```

验证：

```bash
curl http://127.0.0.1:8002/health
open http://127.0.0.1:8002/web/
open http://127.0.0.1:8002/docs
curl -X POST http://127.0.0.1:8002/api/prompt-context \
  -H 'Content-Type: application/json' \
  --data @demos/02-prompt-context/examples/request.json
```

## 验证重点

- 不传版本时会返回 `baseline-v1` 与 `grounded-v2` 两份结果。
- 传入未知版本时返回 `400 invalid_prompt_version` 和 `trace_id`。
- 超过 `PROMPT_CONTEXT_MAX_CHARS` 的上下文会标记 `truncated: true`。
- 在 Gemini 可用时，同一任务可对比两个版本的输出、延迟和 token 用量；不要只凭单次结果就断言某版更好。
- `503` 会由 SDK 在单次请求内重试；`429` 配额耗尽会返回 `provider_quota_exceeded` 与 `trace_id`，页面会显示错误而不是误报流式完成。

运行测试：

```bash
go test ./...
```

## 复盘

Prompt 优化是实验过程，不是把指令写得越长越好。先定义任务和评估标准，再只改变一个变量，例如增加 few-shot 示例、调整输出约束或变更上下文组织；记录输出质量、稳定性、延迟与 token，才知道变化是否值得保留。
