# 06. Memory State

## Goal

学习会话状态、用户偏好和长期记忆的边界。

## API Sketch

```http
POST /api/memory-state
Content-Type: application/json
```

```json
{
  "session_id": "demo-session",
  "message": "以后回答我时请优先给代码例子"
}
```

## Acceptance

- 同一 session 能延续上下文。
- 用户偏好能影响后续回答。
- 清除 session 后不再使用旧上下文。

