import assert from "node:assert/strict";
import test from "node:test";
import {
  createSSEParser,
  formatAgentLabel,
  isSuperAdmin,
  safeUrlTransform,
  validateChatMessage,
  validateMemoryContent,
} from "./model";

test("safeUrlTransform allows http/https/mailto and blocks dangerous schemes", () => {
  assert.equal(
    safeUrlTransform("https://domus.paroki.org"),
    "https://domus.paroki.org",
  );
  assert.equal(
    safeUrlTransform("http://localhost:3000/a"),
    "http://localhost:3000/a",
  );
  assert.equal(
    safeUrlTransform("mailto:sekretariat@paroki.org"),
    "mailto:sekretariat@paroki.org",
  );
  assert.equal(safeUrlTransform("javascript:alert(1)"), "");
  assert.equal(safeUrlTransform("JAVASCRIPT:alert(1)"), "");
  assert.equal(
    safeUrlTransform("data:text/html,<script>alert(1)</script>"),
    "",
  );
  assert.equal(safeUrlTransform("vbscript:msgbox(1)"), "");
  assert.equal(safeUrlTransform("/relative/path"), "");
});

test("createSSEParser handles split chunks, keepalive pings, and typed frames", () => {
  const parser = createSSEParser();
  const frames = [
    ...parser.push(": ping\n\nevent: ses"),
    ...parser.push(
      'sion\ndata: {"sessionId":"s1","title":"Halo","isNew":true}\n\n',
    ),
    ...parser.push(
      'event: agent_transfer\ndata: {"from":"gabriel","to":"mikael"}\n\n',
    ),
    ...parser.push(
      'event: tool_start\ndata: {"agent":"mikael","toolCallId":"c1","toolName":"event_list"}\n\n',
    ),
    ...parser.push(
      'event: tool_end\ndata: {"agent":"mikael","toolCallId":"c1","toolName":"event_list","isError":false}\n\n',
    ),
    ...parser.push(
      'event: token\ndata: {"agent":"gabriel","delta":"Halo "}\n\n',
    ),
    ...parser.push(
      'event: done\ndata: {"sessionId":"s1","messageId":"m1","agent":"gabriel"}\n\n',
    ),
    ...parser.flush(),
  ];

  assert.equal(frames.length, 6);
  assert.equal(frames[0]?.event, "session");
  assert.equal(frames[1]?.event, "agent_transfer");
  assert.equal(frames[2]?.event, "tool_start");
  assert.equal(frames[3]?.event, "tool_end");
  assert.equal(frames[4]?.event, "token");
  assert.equal(frames[5]?.event, "done");
});

test("validators and helpers enforce rune limits and roles", () => {
  assert.equal(validateMemoryContent("   ") !== null, true);
  assert.equal(validateMemoryContent("a".repeat(2000)), null);
  assert.equal(validateMemoryContent("a".repeat(2001)) !== null, true);

  assert.equal(validateChatMessage("a".repeat(16000)), null);
  assert.equal(validateChatMessage("a".repeat(16001)) !== null, true);

  assert.equal(formatAgentLabel("gabriel"), "Gabriel");
  assert.equal(formatAgentLabel("mikael"), "Mikael");
  assert.equal(formatAgentLabel("rafael"), "Rafael");
  assert.equal(isSuperAdmin(["user:super-admin"]), true);
  assert.equal(isSuperAdmin(["user:user"]), false);
});
