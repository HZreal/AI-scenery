# Demos

每个 demo 聚焦一个技术点，并尽量以“后端智能服务”的形式落地。

## 统一结构

```text
demos/<number>-<topic>/
  README.md
  src/
  tests/
  examples/
```

## README 推荐结构

```markdown
# <Topic>

## Goal

这个 demo 学什么，解决什么问题。

## Tech Stack

使用的语言、框架和关键依赖。

## API

接口路径、请求结构、响应结构。

## Run

如何安装依赖和启动服务。

## Tests

如何运行测试，验证哪些场景。

## Notes

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

