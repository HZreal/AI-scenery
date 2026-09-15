# 09. Safety Control

## Goal

学习 Agent 安全边界、人工审批和高风险操作控制。

## API Sketch

```http
POST /api/safety-control
Content-Type: application/json
```

```json
{
  "task": "准备删除临时文件"
}
```

## Acceptance

- 普通操作自动执行。
- 高风险操作进入待审批状态。
- 越权参数被拒绝并记录原因。

