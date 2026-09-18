# 09. 安全与可控性

## 学习目标

学习 Agent 安全边界、人工审批和高风险操作控制。

## 接口草图

```http
POST /api/safety-control
Content-Type: application/json
```

```json
{
  "task": "准备删除临时文件"
}
```

## 完成标准

- 普通操作自动执行。
- 高风险操作进入待审批状态。
- 越权参数被拒绝并记录原因。
