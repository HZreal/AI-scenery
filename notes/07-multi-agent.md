# 07. 编排与多 Agent

## 学习目标

理解 planner-worker、reviewer、handoff 等多 Agent 协作模式。

## 核心要点

- Planner / Worker / Reviewer 分工。
- Handoff。
- 状态图与条件路由。
- 任务拆解与结果聚合。
- 多 Agent 的成本和复杂度边界。

## Demo 方向

`demos/07-multi-agent/`

实现规划 Agent、执行 Agent、审查 Agent 协作完成一个简单任务。

## 验证场景

- Planner 能输出明确步骤。
- Worker 能按步骤执行。
- Reviewer 能发现不符合要求的结果。

