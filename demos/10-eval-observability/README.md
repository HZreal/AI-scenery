# 10. Eval Observability

## Goal

学习如何评测、追踪和复盘 Agent 服务。

## API Sketch

```http
POST /api/eval
Content-Type: application/json
```

```json
{
  "suite": "agent-baseline"
}
```

## Acceptance

- 能运行固定评测用例。
- 能输出成功率和失败原因。
- 能根据 trace 复盘失败执行。

