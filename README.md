# AI Scenery

AI Scenery 是一个用于个人学习 AI 与 AI Agent 后端应用开发的笔记和实践仓库。

学习方式采用快速迭代：

1. 理解一个关键技术点。
2. 写一篇结构化学习笔记。
3. 做一个可运行的小 demo。
4. 用测试、样例和复盘确认掌握边界。

## 学习主线

本项目以“后端智能服务”为主线。每个 demo 尽量落成一个可被 API 调用的服务能力，而不是只停留在脚本或 notebook。

语言不强制单栈：

- Python：AI、RAG、Agent 框架、评测实验。
- TypeScript / JavaScript：Web API、前端交互、工具生态和 MCP 集成。
- Go：后端服务化、工程接入、中间件和高并发场景。

## 目录结构

```text
notes/              # 学习笔记和路线图
demos/              # 每个技术点一个可运行 demo
resources/          # 论文、官方文档、数据集和资料
scripts/            # 评测、导入、辅助脚本
```

每个 demo 建议使用统一结构：

```text
demos/<number>-<topic>/
  README.md
  src/
  tests/
  examples/
```

## 学习路径

路线总览见 [notes/00-roadmap.md](notes/00-roadmap.md)。

当前核心学习 tracks：

1. LLM 应用基础
2. Prompt 与上下文工程
3. 工具调用 / Function Calling
4. RAG 与知识库
5. Agent Loop
6. 状态、记忆与会话
7. 编排与多 Agent
8. MCP 与外部能力集成
9. 安全、审批与可控性
10. 评测、观测与部署

## 每个学习单元的完成标准

- 有一篇学习笔记。
- 有一个可运行 demo。
- 有请求和响应样例。
- 有至少 2-3 个验证场景。
- 有复盘：适用场景、坑点、下一步。

