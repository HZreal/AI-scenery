# 08. MCP 与外部能力集成

## 学习目标

理解 MCP 如何把本地资源、工具和提示暴露给 AI 应用。

## 核心要点

- MCP host、client、server 的关系。
- Tools、resources、prompts 的区别。
- JSON-RPC 通信模型。
- 工具描述与参数 schema。
- 本地能力暴露时的安全边界。

## Demo 方向

`demos/08-mcp/`

实现一个最小 MCP server，把本地学习资料或小工具暴露给 Agent。

## 验证场景

- MCP client 能列出工具或资源。
- 能读取一个本地学习资源。
- 非授权路径或参数会被拒绝。

