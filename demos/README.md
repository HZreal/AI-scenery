# 学习 Demo

每个 demo 聚焦一个技术点，并尽量以“后端智能服务”的形式落地。

## 统一结构

```text
demos/<number>-<topic>/
  README.md
  lesson/            # Go 版阶段特有逻辑；公共能力位于根级 internal/
  examples/
```

Python 等独立语言实现仍可在阶段目录的 `src/`、`tests/` 中保留，用于横向对照。

## README 推荐结构

```markdown
# <主题>

## 学习目标

这个 demo 学什么，解决什么问题。

## 技术选型

使用的语言、框架和关键依赖。

## 接口

接口路径、请求结构、响应结构。

## 运行方式

如何安装依赖和启动服务。

## 验证

如何运行测试，验证哪些场景。

## 复盘

适用场景、坑点和下一步。
```

## 统一响应结构

```json
{
  "result": "...",
  "metadata": {
    "model": "...",
    "latency_ms": 0,
    "cost_estimate": "...",
    "trace_id": "..."
  },
  "steps": [],
  "tool_calls": [],
  "sources": [],
  "errors": []
}
```
