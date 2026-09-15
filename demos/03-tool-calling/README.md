# 03. Tool Calling

## Goal

学习如何让模型安全地调用后端函数完成确定性任务。

## API Sketch

```http
POST /api/tool-calling
Content-Type: application/json
```

```json
{
  "message": "帮我计算 3 件 129 元商品打 8 折后的总价"
}
```

## Acceptance

- 工具参数有 schema 和校验。
- 工具失败时返回清晰错误。
- 响应记录 tool_calls。

