import assert from "node:assert/strict";
import test from "node:test";
import {
  createSSEParser,
  extractNavigationMeta,
  extractQuestionMeta,
  formatAgentLabel,
  formatToolJSON,
  getDockSessionKey,
  isSuperAdmin,
  isToolErrorContent,
  safeUrlTransform,
  validateChatMessage,
  validateMemoryContent,
} from "./model";

test("safeUrlTransform allows http/https/mailto and internal paths, blocks dangerous schemes", () => {
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
  assert.equal(safeUrlTransform("/org/123"), "/org/123");
  assert.equal(safeUrlTransform("/org/123#invites"), "/org/123#invites");
  assert.equal(safeUrlTransform("//evil.com"), "");
  assert.equal(safeUrlTransform("javascript:alert(1)"), "");
  assert.equal(safeUrlTransform("JAVASCRIPT:alert(1)"), "");
  assert.equal(
    safeUrlTransform("data:text/html,<script>alert(1)</script>"),
    "",
  );
  assert.equal(safeUrlTransform("vbscript:msgbox(1)"), "");
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
      'event: question\ndata: {"question":"Pilih jenis acara?","options":[{"label":"Misa","value":"Misa"},{"label":"Rapat","value":"Rapat"}],"multiSelect":false}\n\n',
    ),
    ...parser.push(
      'event: navigate\ndata: {"path":"/org/123","label":"Buka Organisasi","autoRedirect":true}\n\n',
    ),
    ...parser.push(
      'event: done\ndata: {"sessionId":"s1","messageId":"m1","agent":"gabriel"}\n\n',
    ),
    ...parser.flush(),
  ];

  assert.equal(frames.length, 8);
  assert.equal(frames[0]?.event, "session");
  assert.equal(frames[1]?.event, "agent_transfer");
  assert.equal(frames[2]?.event, "tool_start");
  assert.equal(frames[3]?.event, "tool_end");
  assert.equal(frames[4]?.event, "token");
  assert.equal(frames[5]?.event, "question");
  assert.equal(frames[6]?.event, "navigate");
  assert.equal(frames[7]?.event, "done");
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
  assert.equal(formatAgentLabel("yosef"), "Yosef");
  assert.equal(formatAgentLabel("gregorius"), "Gregorius");
  assert.equal(formatAgentLabel("sesilia"), "Sesilia");
  assert.equal(formatAgentLabel("unknown"), "Gabriel");
  assert.equal(getDockSessionKey("p-123"), "domus:agent:dockSession:p-123");
  assert.equal(isSuperAdmin(["user:super-admin"]), true);
  assert.equal(isSuperAdmin(["user:user"]), false);

  assert.equal(
    isToolErrorContent(
      "Tool event_list error: Invalid 'from' timestamp; expected RFC3339.",
    ),
    true,
  );
  assert.equal(isToolErrorContent("Error: connection refused"), true);
  assert.equal(isToolErrorContent('{"items":[]}'), false);

  assert.equal(formatToolJSON('{"a":1}'), '{\n  "a": 1\n}');
  assert.equal(formatToolJSON("raw-text"), "raw-text");
  assert.equal(formatToolJSON(""), "");
});

test("extractQuestionMeta and extractNavigationMeta parse and validate payloads", () => {
  assert.equal(extractQuestionMeta(undefined), null);
  assert.equal(extractQuestionMeta({}), null);
  assert.equal(
    extractQuestionMeta({ question: { question: "Pilih?", options: [] } }),
    null,
  );
  assert.deepEqual(
    extractQuestionMeta({
      question: {
        question: "Pilih bidang?",
        options: [
          { label: "Lektor", value: "Lektor" },
          { label: "Mazmur", value: "Mazmur" },
        ],
        multiSelect: true,
      },
    }),
    {
      question: "Pilih bidang?",
      options: [
        { label: "Lektor", value: "Lektor" },
        { label: "Mazmur", value: "Mazmur" },
      ],
      multiSelect: true,
    },
  );

  assert.equal(extractNavigationMeta(undefined), null);
  assert.equal(
    extractNavigationMeta({ navigation: { path: "//evil.com", label: "Go" } }),
    null,
  );
  assert.equal(
    extractNavigationMeta({
      navigation: { path: "javascript:alert(1)", label: "Go" },
    }),
    null,
  );
  assert.deepEqual(
    extractNavigationMeta({
      navigation: {
        path: "/org/123#invites",
        label: "Buka Undangan",
        description: "Kelola undangan anggota",
        autoRedirect: false,
      },
    }),
    {
      path: "/org/123#invites",
      label: "Buka Undangan",
      description: "Kelola undangan anggota",
      autoRedirect: false,
    },
  );
});
