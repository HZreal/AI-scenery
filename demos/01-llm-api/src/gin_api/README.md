# Gin LLM 后端接口

这是与 Python `llm_api` 并列的 Go/Gin 实现，用于学习同一个 LLM 后端接口如何在 Go 中分层、替换模型提供商并公开 OpenAPI。

## 学习要点

- `internal/provider.Provider` 是模型提供商抽象；HTTP 层不依赖 Gemini SDK 的类型。
- `provider.Factory` 根据 `AI_SCENERY_GO_PROVIDER` 创建 `mock`、`gemini` 或 `openai` 实现；后续新增模型只需注册新的 Builder。
- `geminiProvider` 使用官方 `google.golang.org/genai` SDK，将 `system/user/assistant` 映射为 Gemini 的 `systemInstruction/user/model`。
- `openAIProvider` 使用官方 `openai-go` 的 Responses API，并保留 `system/user/assistant` 消息角色。
- Gin 提供 `POST /api/chat`、`GET /health`、`GET /openapi.json` 和 `GET /docs`。
- `stream=true` 使用稳定的 SSE 事件：`metadata`、`delta`、`completed`、`error`。
- `../public/index.html` 是零依赖的流式聊天页面，由 Gin 在 `/public/` 提供。

## 配置环境变量

项目根目录的 `.env.local` 被 Git 忽略。先用 mock 验证服务：

```bash
AI_SCENERY_GO_PROVIDER=mock
GIN_LLM_API_ADDRESS=127.0.0.1:8002
MAX_INPUT_CHARS=12000
```

调用 Gemini 时改为：

```bash
AI_SCENERY_GO_PROVIDER=gemini
GEMINI_API_KEY=your-api-key
GEMINI_MODEL=gemini-3.8-flash
```

调用 OpenAI 时改为：

```bash
AI_SCENERY_GO_PROVIDER=openai
OPENAI_API_KEY=your-api-key
OPENAI_MODEL=gpt-5.5
# 可选：兼容网关或代理地址；留空时使用 OpenAI 默认地址。
OPENAI_BASE_URL=
```

真实密钥只放在 `.env.local`，不要打印或提交。示例模型名来自当前 Google 官方 Go SDK 文档；如该模型在你的账号不可用，请在 `.env.local` 中换成 Gemini API 控制台可用的模型名。

## 启动服务

在仓库根目录运行：

```bash
go run ./demos/01-llm-api/src/gin_api
```

本项目当前最低要求 Go `1.25.0`，用于匹配新版 OpenAI SDK；当前本机已验证 Go `1.26.8` 可运行。默认 `GOTOOLCHAIN=auto` 会在项目需要更高版本时按 `go.mod` 自动选择工具链。

## 在 GoLand 中运行

新建 `Go Build` 配置，不要复用截图中的 Python 配置：

- `Run kind` 选择 `Package`，包路径填写 `github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api`。
- `Working directory` 设为仓库根目录 `/Users/huang/Documents/ChatGPT/AI-scenery`，这样才能自动读取根目录的 `.env.local`。
- 开始阶段可在环境变量中设置 `AI_SCENERY_GO_PROVIDER=mock`；需要调用 Gemini 时只在 `.env.local` 设置 `GEMINI_API_KEY`。

GoLand 应使用当前安装的 Go 1.26.8 SDK（`/usr/local/go`）。该运行配置不需要 Python SDK，也不需要把 `go run` 填到 Python 的脚本路径。

## 验证接口

服务启动后，在另一个终端执行：

```bash
curl http://127.0.0.1:8002/health
open http://127.0.0.1:8002/docs
curl -s http://127.0.0.1:8002/openapi.json | python3 -m json.tool
open http://127.0.0.1:8002/public/
```

在 Swagger UI 中展开 `POST /api/chat`，点击 `Try it out` 可直接调试。也可以使用命令行：

```bash
curl -X POST http://127.0.0.1:8002/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"用一句话解释什么是上下文窗口。"}'
```

结构化 JSON 响应：

```bash
curl -X POST http://127.0.0.1:8002/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"列出两个 Agent 开发学习点。","mode":"json"}'
```

流式响应：

```bash
curl -N -X POST http://127.0.0.1:8002/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"流式响应有什么用？","stream":true}'
```

## 流式聊天页面

启动服务后访问 `http://127.0.0.1:8002/`，会跳转到 `/public/` 中的聊天页面。页面通过 `fetch` 读取 SSE 数据流，将每个 `delta` 事件立即追加到助手消息中，并在下一轮请求中提交当前页面的对话历史。

修改 `public/index.html` 后需要重启 Gin 服务，已启动的旧进程不会自动加载新的路由或静态文件。

## 运行测试

```bash
go test ./...
```

测试覆盖 API 成功与错误响应、OpenAPI 暴露、Gemini/OpenAI 消息角色映射，以及两种 Provider 缺失密钥时的工厂保护。它们使用本地 stub 或 mock，不会发出真实模型请求。

## 复盘

工厂模式解决的是“服务启动时选择哪种提供商”，而接口抽象解决的是“HTTP 业务如何不被具体 SDK 绑住”。Gemini 与 OpenAI 的 SDK 差异只存在于各自适配器中；真实调用的网络错误、配额错误和无效 JSON 会统一映射为带 `trace_id` 的错误响应。模型价格会随时间变化，因此本 Demo 暂不硬编码成本估算。
