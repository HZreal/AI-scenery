# 05. Agent Loop

## Goal

学习 Agent 如何通过多步循环调用工具完成任务。

## API Sketch

```http
POST /api/agent-loop
Content-Type: application/json
```

```json
{
  "task": "查询资料并生成一段摘要"
}
```

## Acceptance

- 响应记录 steps。
- 响应记录 tool_calls。
- 有最大步骤数和停止条件。

