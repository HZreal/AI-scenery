from __future__ import annotations

import json
import os
import time
import urllib.error
import urllib.request
import uuid
from dataclasses import dataclass
from decimal import Decimal, InvalidOperation
from typing import Any, Iterator, Literal, Protocol

Mode = Literal["text", "json"]
Role = Literal["system", "user", "assistant"]


class ChatError(Exception):
    def __init__(self, code: str, message: str, status_code: int = 400) -> None:
        super().__init__(message)
        self.code = code
        self.status_code = status_code


@dataclass(frozen=True)
class ChatMessage:
    role: Role
    content: str

    def __post_init__(self) -> None:
        if self.role not in {"system", "user", "assistant"}:
            raise ChatError("invalid_role", "role must be system, user, or assistant")
        if not self.content.strip():
            raise ChatError("empty_message", "message content must not be empty")


@dataclass(frozen=True)
class ChatRequest:
    messages: tuple[ChatMessage, ...]
    mode: Mode = "text"

    def __post_init__(self) -> None:
        if not self.messages:
            raise ChatError("empty_messages", "messages must not be empty")
        if self.mode not in {"text", "json"}:
            raise ChatError("invalid_mode", "mode must be text or json")

    @property
    def input_characters(self) -> int:
        return sum(len(message.content) for message in self.messages)


@dataclass(frozen=True)
class ProviderResult:
    result: str | dict[str, Any]
    model: str
    input_tokens: int | None = None
    output_tokens: int | None = None


class LLMProvider(Protocol):
    def generate(self, request: ChatRequest) -> ProviderResult: ...
    def stream(self, request: ChatRequest) -> Iterator[str]: ...


class MockLLMProvider:
    model = "mock-llm"

    def generate(self, request: ChatRequest) -> ProviderResult:
        if request.mode == "json":
            return ProviderResult({"topic": "LLM API", "summary": "这是一个本地 mock 响应，用于验证结构化输出契约。", "key_points": ["消息结构", "上下文预算", "流式响应"]}, self.model)
        return ProviderResult(f"AI Agent 可以围绕目标理解任务、调用工具并返回结果。你问的是：{request.messages[-1].content}", self.model)

    def stream(self, request: ChatRequest) -> Iterator[str]:
        result = self.generate(ChatRequest(request.messages, "text")).result
        for word in str(result).split():
            yield f"{word} "


class OpenAIResponsesProvider:
    def __init__(self, api_key: str, model: str = "gpt-5.5", base_url: str = "https://api.openai.com/v1") -> None:
        self.api_key, self.model, self.base_url = api_key, model, base_url.rstrip("/")

    def generate(self, request: ChatRequest) -> ProviderResult:
        data = self._request(self._payload(request, stream=False))
        output = data.get("output_text")
        if not isinstance(output, str):
            raise ChatError("provider_response_invalid", "OpenAI response did not contain output text", 500)
        if request.mode == "json":
            try:
                output = json.loads(output)
            except json.JSONDecodeError as exc:
                raise ChatError("provider_json_invalid", "OpenAI response was not valid JSON", 500) from exc
        usage = data.get("usage") or {}
        return ProviderResult(output, self.model, usage.get("input_tokens"), usage.get("output_tokens"))

    def stream(self, request: ChatRequest) -> Iterator[str]:
        req = urllib.request.Request(f"{self.base_url}/responses", data=json.dumps(self._payload(request, stream=True)).encode(), headers={"Authorization": f"Bearer {self.api_key}", "Content-Type": "application/json"}, method="POST")
        try:
            with urllib.request.urlopen(req, timeout=60) as response:
                for raw in response:
                    line = raw.decode().strip()
                    if line.startswith("data: "):
                        event = json.loads(line[6:])
                        if event.get("type") == "response.output_text.delta":
                            yield event.get("delta", "")
                        if event.get("type") == "error":
                            raise ChatError("provider_stream_error", event.get("message", "OpenAI streaming failed"), 500)
        except urllib.error.URLError as exc:
            raise ChatError("provider_network_error", f"OpenAI API network error: {exc.reason}", 500) from exc

    def _payload(self, request: ChatRequest, stream: bool) -> dict[str, Any]:
        payload: dict[str, Any] = {"model": self.model, "input": [{"role": item.role, "content": item.content} for item in request.messages], "stream": stream}
        if request.mode == "json":
            payload["text"] = {"format": {"type": "json_schema", "name": "llm_api_demo", "strict": True, "schema": {"type": "object", "additionalProperties": False, "properties": {"topic": {"type": "string"}, "summary": {"type": "string"}, "key_points": {"type": "array", "items": {"type": "string"}}}, "required": ["topic", "summary", "key_points"]}}}
        return payload

    def _request(self, payload: dict[str, Any]) -> dict[str, Any]:
        req = urllib.request.Request(f"{self.base_url}/responses", data=json.dumps(payload).encode(), headers={"Authorization": f"Bearer {self.api_key}", "Content-Type": "application/json"}, method="POST")
        try:
            with urllib.request.urlopen(req, timeout=60) as response:
                return json.loads(response.read().decode())
        except urllib.error.HTTPError as exc:
            raise ChatError("provider_api_error", f"OpenAI API error {exc.code}", 500) from exc
        except urllib.error.URLError as exc:
            raise ChatError("provider_network_error", f"OpenAI API network error: {exc.reason}", 500) from exc


def request_from_payload(payload: dict[str, Any]) -> ChatRequest:
    messages, message = payload.get("messages"), payload.get("message")
    if messages is not None and message is not None:
        raise ChatError("ambiguous_input", "provide messages or message, not both")
    if messages is None:
        if not isinstance(message, str):
            raise ChatError("missing_messages", "messages or message is required")
        messages = [{"role": "user", "content": message}]
    if not isinstance(messages, list):
        raise ChatError("invalid_messages", "messages must be an array")
    if not all(isinstance(item, dict) for item in messages):
        raise ChatError("invalid_messages", "each message must be an object")
    return ChatRequest(tuple(ChatMessage(role=item.get("role"), content=item.get("content", "")) for item in messages), payload.get("mode", "text"))


def max_input_characters() -> int:
    try:
        return max(1, int(os.environ.get("MAX_INPUT_CHARS", "12000")))
    except ValueError as exc:
        raise ChatError("invalid_context_limit", "MAX_INPUT_CHARS must be a positive integer", 500) from exc


def create_chat_response(request: ChatRequest, provider: LLMProvider, trace_id: str | None = None) -> dict[str, Any]:
    trace_id = trace_id or new_trace_id()
    limit = max_input_characters()
    if request.input_characters > limit:
        raise ChatError("context_limit_exceeded", f"input exceeds MAX_INPUT_CHARS={limit}")
    started = time.perf_counter()
    result = provider.generate(request)
    return {"result": result.result, "metadata": metadata(result, request.input_characters, trace_id, int((time.perf_counter() - started) * 1000))}


def metadata(result: ProviderResult, input_characters: int, trace_id: str, latency_ms: int = 0) -> dict[str, Any]:
    return {"model": result.model, "latency_ms": latency_ms, "input_characters": input_characters, "usage": {"input_tokens": result.input_tokens, "output_tokens": result.output_tokens}, "cost_estimate_usd": calculate_cost(result.input_tokens, result.output_tokens), "trace_id": trace_id}


def calculate_cost(input_tokens: int | None, output_tokens: int | None) -> str | None:
    if input_tokens is None or output_tokens is None:
        return None
    try:
        return str((Decimal(input_tokens) * Decimal(os.environ["OPENAI_INPUT_USD_PER_1M_TOKENS"]) + Decimal(output_tokens) * Decimal(os.environ["OPENAI_OUTPUT_USD_PER_1M_TOKENS"])) / Decimal(1_000_000))
    except (KeyError, InvalidOperation):
        return None


def new_trace_id() -> str:
    return f"trace-{uuid.uuid4().hex[:12]}"


def error_response(error: ChatError, trace_id: str) -> dict[str, Any]:
    return {"error": str(error), "code": error.code, "metadata": {"trace_id": trace_id}}


def load_env_file(path: str) -> None:
    if os.path.exists(path):
        for line in open(path, encoding="utf-8"):
            if "=" in line and not line.lstrip().startswith("#"):
                key, value = line.strip().split("=", 1)
                os.environ.setdefault(key.strip(), value.strip().strip('"').strip("'"))


def provider_from_env() -> LLMProvider:
    provider = os.environ.get("AI_SCENERY_PROVIDER", "mock").strip().lower()
    if provider == "mock":
        return MockLLMProvider()
    if provider == "openai":
        key = os.environ.get("OPENAI_API_KEY", "").strip()
        if not key:
            raise ChatError("missing_api_key", "OPENAI_API_KEY is required when AI_SCENERY_PROVIDER=openai", 500)
        return OpenAIResponsesProvider(key, os.environ.get("OPENAI_MODEL", "gpt-5.5"), os.environ.get("OPENAI_BASE_URL", "https://api.openai.com/v1"))
    raise ChatError("invalid_provider", "AI_SCENERY_PROVIDER must be mock or openai", 500)
