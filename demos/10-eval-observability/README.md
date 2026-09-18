# 10. 评测与观测

## 学习目标

学习如何评测、追踪和复盘 Agent 服务。

## 接口草图

```http
POST /api/eval
Content-Type: application/json
```

```json
{
  "suite": "agent-baseline"
}
```

## 完成标准

- 能运行固定评测用例。
- 能输出成功率和失败原因。
- 能根据 trace 复盘失败执行。
