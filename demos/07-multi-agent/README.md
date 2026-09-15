# 07. Multi Agent

## Goal

学习多个 Agent 如何分工、交接和审查结果。

## API Sketch

```http
POST /api/multi-agent
Content-Type: application/json
```

```json
{
  "task": "把一段需求拆成实现步骤并检查遗漏"
}
```

## Acceptance

- Planner 输出步骤。
- Worker 执行步骤。
- Reviewer 给出审查结果。

