import assert from "node:assert/strict";
import test from "node:test";

import { parseSSEEvent } from "./sse.js";

test("parseSSEEvent accepts Gin event formatting", () => {
  const event = parseSSEEvent('event:error\r\ndata:{"code":"provider_quota_exceeded"}');
  assert.equal(event.type, "error");
  assert.deepEqual(JSON.parse(event.data), { code: "provider_quota_exceeded" });
});
