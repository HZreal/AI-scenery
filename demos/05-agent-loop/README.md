# 05. Agent 循环

## 学习目标

学习 Agent 如何通过多步循环调用工具完成任务。

## 接口草图

```http
POST /api/agent-loop
Content-Type: application/json
```

```json
{
  "task": "查询资料并生成一段摘要"
}
```

## 完成标准

- 响应记录 steps。
- 响应记录 tool_calls。
- 有最大步骤数和停止条件。
