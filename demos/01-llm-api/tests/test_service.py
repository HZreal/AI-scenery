import os
import unittest
from unittest.mock import MagicMock, patch

from llm_api.service import ChatError, MockLLMProvider, OpenAIResponsesProvider, create_chat_response, request_from_payload


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

    def test_openai_provider_reads_text_from_raw_response_content(self):
        request = request_from_payload({"message": "hello"})
        raw_response = {"output": [{"content": [{"type": "output_text", "text": "world"}]}], "usage": {"input_tokens": 3, "output_tokens": 2}}
        response = MagicMock()
        response.read.return_value = __import__("json").dumps(raw_response).encode()

        with patch("llm_api.service.urllib.request.urlopen") as urlopen:
            urlopen.return_value.__enter__.return_value = response
            result = OpenAIResponsesProvider("test-key").generate(request)

        self.assertEqual(result.result, "world")
        self.assertEqual(result.input_tokens, 3)
