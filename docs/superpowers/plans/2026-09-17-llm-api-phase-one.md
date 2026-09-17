# LLM API Phase One Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成带 OpenAPI 文档的 LLM API 第一阶段学习 Demo。

**Architecture:** 使用 Flask + flask-smorest 将 Schema、参数校验和 OpenAPI 规范定义在一起。服务层以 Provider Result 和统一错误模型承载消息、上下文、usage、成本和流式事件。

**Tech Stack:** Python 3.12、uv、Flask、flask-smorest、Marshmallow、unittest。

**Spec:** `docs/superpowers/specs/2026-09-17-llm-api-phase-one-design.md`

## Global Constraints

- Python 依赖只通过 `uv add` 安装到项目 `.venv`。
- 真实 Key 只留在 `.env.local`；不得读取、打印或用其发起验证请求。
- OpenAPI 使用 3.0.3，并提供 `/openapi.json` 和 `/docs`。
- 测试只覆盖关键公开契约和高风险分支，不为简单辅助函数追求覆盖率。

---

### Task 1: Provider 与消息服务契约

**Files:**
- Create: `demos/01-llm-api/src/llm_api/schemas.py`
- Modify: `demos/01-llm-api/src/llm_api/service.py`
- Modify: `demos/01-llm-api/tests/test_service.py`

**Interfaces:**
- Consumes: `messages`, `message`, `mode`, `stream` 请求字段。
- Produces: `ChatRequest`、`ProviderResult`、`ChatError` 和统一 metadata。

- [ ] **Step 1: 写入消息数组、上下文超限、JSON 对象和成本计算的失败测试。**
- [ ] **Step 2: 运行服务层测试，确认新契约尚未实现。**
- [ ] **Step 3: 实现消息归一化、字符预算、Provider Result、usage/成本与 trace 错误。**
- [ ] **Step 4: 运行服务层测试，确认通过。**

### Task 2: OpenAPI 与 Flask 路由

**Files:**
- Modify: `pyproject.toml`
- Modify: `uv.lock`
- Modify: `demos/01-llm-api/src/llm_api/server.py`
- Modify: `demos/01-llm-api/tests/test_server.py`

**Interfaces:**
- Consumes: Task 1 的服务对象和 Marshmallow Schema。
- Produces: `/health`、`/api/chat`、`/openapi.json`、`/docs` 与 SSE 事件序列。

- [ ] **Step 1: 写入 OpenAPI、messages 请求、SSE completed 与错误 trace 的失败路由测试。**
- [ ] **Step 2: 运行路由测试，确认失败原因是旧路由没有该契约。**
- [ ] **Step 3: 用 `uv add flask-smorest` 安装依赖并实现 Blueprint、Schema、OpenAPI 和 SSE 路由。**
- [ ] **Step 4: 运行路由测试，确认通过。**

### Task 3: 学习材料与端到端验证

**Files:**
- Modify: `notes/01-llm-basics.md`
- Modify: `demos/01-llm-api/README.md`
- Modify: `demos/01-llm-api/examples/request.json`
- Modify: `demos/01-llm-api/examples/response.json`

**Interfaces:**
- Consumes: 最终 API 契约与运行命令。
- Produces: 可运行说明、关键概念说明与同步样例。

- [ ] **Step 1: 更新笔记、README 和样例，使其与 OpenAPI 契约一致。**
- [ ] **Step 2: 运行 `npm test`、`uv lock --check`、`git diff --check`。**
- [ ] **Step 3: 启动 Flask 服务，以 curl 验证 `/health`、`/openapi.json`、JSON 和 SSE。**
- [ ] **Step 4: 提交第一阶段实现。**
