# 02. Prompt Context

## Goal

对比不同 prompt 设计对输出质量和稳定性的影响。

## API Sketch

```http
POST /api/prompt-context
Content-Type: application/json
```

```json
{
  "task": "总结一段技术文档",
  "prompt_version": "v1"
}
```

## Acceptance

- 至少保留两个 prompt 版本。
- 同一输入可对比不同版本输出。
- README 记录每个版本的适用场景和问题。

