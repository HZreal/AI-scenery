import os
import unittest
from unittest.mock import patch

from llm_api.service import ChatError, MockLLMProvider, create_chat_response, request_from_payload


class ChatServiceTest(unittest.TestCase):
    def test_message_shorthand_becomes_user_message(self):
        request = request_from_payload({"message": "解释 AI Agent", "mode": "text"})

        self.assertEqual(request.messages[0].role, "user")
        self.assertEqual(request.messages[0].content, "解释 AI Agent")

    def test_json_mode_returns_json_object(self):
        request = request_from_payload({"messages": [{"role": "user", "content": "列出学习点"}], "mode": "json"})

        response = create_chat_response(request, MockLLMProvider())

        self.assertEqual(response["result"]["topic"], "LLM API")
        self.assertIsNone(response["metadata"]["cost_estimate_usd"])

    def test_context_budget_rejects_oversized_input(self):
        request = request_from_payload({"message": "abcdef"})
        with patch.dict(os.environ, {"MAX_INPUT_CHARS": "5"}):
            with self.assertRaises(ChatError) as raised:
                create_chat_response(request, MockLLMProvider())

        self.assertEqual(raised.exception.code, "context_limit_exceeded")
