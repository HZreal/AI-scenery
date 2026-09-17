from __future__ import annotations

import json
from pathlib import Path
from typing import Iterator

from flask import Flask, Response, jsonify, stream_with_context
from flask_smorest import Api, Blueprint

from .schemas import ChatInputSchema
from .service import (
    ChatError,
    ProviderResult,
    create_chat_response,
    error_response,
    load_env_file,
    metadata,
    new_trace_id,
    provider_from_env,
    request_from_payload,
)

blp = Blueprint("llm", __name__, description="LLM application basics")


@blp.route("/health", methods=["GET"])
def health() -> tuple[Response, int]:
    return jsonify({"status": "ok"}), 200


@blp.route("/api/chat", methods=["POST"])
@blp.arguments(ChatInputSchema)
@blp.doc(responses={200: {"description": "JSON response or SSE stream"}, 400: {"description": "Invalid request"}, 500: {"description": "Provider error"}})
def chat(payload: dict) -> tuple[Response, int] | Response:
    trace_id = new_trace_id()
    try:
        request = request_from_payload(payload)
        provider = provider_from_env()
        if payload.get("stream"):
            return Response(stream_with_context(sse_events(provider, request, trace_id)), content_type="text/event-stream; charset=utf-8", headers={"Cache-Control": "no-cache"})
        return jsonify(create_chat_response(request, provider, trace_id))
    except ChatError as exc:
        return jsonify(error_response(exc, trace_id)), exc.status_code


def sse_events(provider, request, trace_id: str) -> Iterator[str]:
    # 本服务的事件名保持稳定，上游 Provider 的事件细节不直接暴露给客户端。
    yield sse("metadata", {"trace_id": trace_id, "model": getattr(provider, "model", "unknown")})
    try:
        for delta in provider.stream(request):
            yield sse("delta", {"delta": delta})
        yield sse("completed", {"metadata": metadata(ProviderResult("", getattr(provider, "model", "unknown")), request.input_characters, trace_id)})
    except ChatError as exc:
        yield sse("error", error_response(exc, trace_id))


def sse(event: str, data: dict) -> str:
    return f"event: {event}\ndata: {json.dumps(data, ensure_ascii=False)}\n\n"


def create_app() -> Flask:
    app = Flask(__name__)
    app.config.update(
        API_TITLE="AI Scenery LLM API",
        API_VERSION="1.0.0",
        OPENAPI_VERSION="3.0.3",
        OPENAPI_URL_PREFIX="/",
        OPENAPI_JSON_PATH="openapi.json",
        OPENAPI_SWAGGER_UI_PATH="/docs",
        OPENAPI_SWAGGER_UI_URL="https://cdn.jsdelivr.net/npm/swagger-ui-dist/",
    )
    api = Api(app)
    api.register_blueprint(blp)

    @app.errorhandler(422)
    def invalid_schema(error):
        trace_id = new_trace_id()
        return jsonify({"error": "request validation failed", "code": "invalid_request", "metadata": {"trace_id": trace_id}}), 422

    return app


def run(host: str = "127.0.0.1", port: int = 8001) -> None:
    load_env_file(str(Path(__file__).resolve().parents[4] / ".env.local"))
    app = create_app()
    print(f"Serving LLM API demo on http://{host}:{port}")
    app.run(host=host, port=port, debug=False)


if __name__ == "__main__":
    run()
