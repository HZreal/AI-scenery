# 07. 多 Agent 协作

## 学习目标

学习多个 Agent 如何分工、交接和审查结果。

## 接口草图

```http
POST /api/multi-agent
Content-Type: application/json
```

```json
{
  "task": "把一段需求拆成实现步骤并检查遗漏"
}
```

## 完成标准

- Planner 输出步骤。
- Worker 执行步骤。
- Reviewer 给出审查结果。
