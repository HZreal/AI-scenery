import json
import os
import unittest
from unittest.mock import patch

from llm_api import server


class ChatAPITest(unittest.TestCase):
    def setUp(self):
        self.environment = patch.dict(os.environ, {"AI_SCENERY_PROVIDER": "mock"})
        self.environment.start()
        self.addCleanup(self.environment.stop)

        app_factory = getattr(server, "create_app", None)
        self.assertIsNotNone(app_factory, "the HTTP server must expose a Flask app factory")
        self.client = app_factory().test_client()

    def test_health_endpoint_returns_ok(self):
        response = self.client.get("/health")

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.get_json(), {"status": "ok"})

    def test_text_mode_returns_response_contract(self):
        response = self.client.post(
            "/api/chat",
            json={"message": "解释什么是 AI Agent", "mode": "text"},
        )

        payload = response.get_json()
        self.assertEqual(response.status_code, 200)
        self.assertIn("AI Agent", payload["result"])
        self.assertEqual(payload["metadata"]["model"], "mock-llm")
        self.assertTrue(payload["metadata"]["trace_id"])

    def test_json_mode_returns_parseable_result(self):
        response = self.client.post(
            "/api/chat",
            json={"message": "列出学习点", "mode": "json"},
        )

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.get_json()["result"]["topic"], "LLM API")

    def test_stream_mode_returns_sse_events(self):
        response = self.client.post(
            "/api/chat",
            json={"message": "解释流式输出", "mode": "text", "stream": True},
        )

        self.assertEqual(response.status_code, 200)
        self.assertTrue(response.content_type.startswith("text/event-stream"))
        self.assertIn(b"event: completed", response.data)

    def test_empty_message_returns_bad_request(self):
        response = self.client.post("/api/chat", json={"message": "   ", "mode": "text"})

        self.assertEqual(response.status_code, 400)
        self.assertEqual(response.get_json()["code"], "empty_message")
        self.assertTrue(response.get_json()["metadata"]["trace_id"])

    def test_messages_input_is_documented_by_openapi(self):
        response = self.client.post(
            "/api/chat",
            json={
                "messages": [
                    {"role": "system", "content": "回答要简洁"},
                    {"role": "user", "content": "解释 AI Agent"},
                ],
                "mode": "text",
            },
        )
        specification = self.client.get("/openapi.json")

        self.assertEqual(response.status_code, 200)
        self.assertEqual(specification.status_code, 200)
        self.assertIn("/api/chat", specification.get_json()["paths"])


if __name__ == "__main__":
    unittest.main()
