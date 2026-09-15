# 04. RAG

## Goal

基于本项目学习笔记实现知识问答，回答必须带来源。

## API Sketch

```http
POST /api/rag
Content-Type: application/json
```

```json
{
  "question": "Agent Loop 的停止条件有哪些？"
}
```

## Acceptance

- 能检索 `notes/` 中的内容。
- 回答包含 sources。
- 无依据问题不编造。

