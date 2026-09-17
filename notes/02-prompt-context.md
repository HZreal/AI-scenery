# 02. Prompt 与上下文工程

> 阶段状态：进行中。实践讲解见 [02-prompt-context-backend-practice.md](02-prompt-context-backend-practice.md)。

## 学习目标

理解如何通过 prompt、示例和上下文组织提升模型输出稳定性。

## 核心要点

- system prompt、developer prompt、user prompt 的职责。
- Few-shot 示例设计。
- 任务拆解与输出约束。
- 上下文裁剪和关键信息保留。
- prompt 版本管理与效果对比。

## Demo 方向

`demos/02-prompt-context/`，使用 Go/Gin + Gemini 实现多个版本的 Prompt 对比接口。

对同一个任务设计多版 prompt，记录输出质量、稳定性和适用场景。

## 验证场景

- 同一输入在不同 prompt 下输出差异可对比。
- 输出结构符合约定。
- prompt 变更有版本记录和复盘。
