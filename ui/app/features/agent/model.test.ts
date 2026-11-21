import assert from "node:assert/strict";
import test from "node:test";
import {
  agentDisplayMeta,
  buildTimelineItems,
  canMutateMemory,
  isParishAdmin,
  parseSSEBuffer,
  sortMessagesChronologically,
  toolDisplayMeta,
} from "./model";

test("isParishAdmin and canMutateMemory enforce scope ownership", () => {
  assert.equal(isParishAdmin(["ref:parish-admin"]), true);
  assert.equal(isParishAdmin(["user:user"]), false);

  assert.equal(
    canMutateMemory(
      { scope: "user", userId: "u1" },
      { id: "u1", roles: ["user:user"] },
    ),
    true,
  );
  assert.equal(
    canMutateMemory(
      { scope: "user", userId: "u2" },
      { id: "u1", roles: ["user:user"] },
    ),
    false,
  );
  assert.equal(
    canMutateMemory({ scope: "parish" }, { id: "u1", roles: ["user:user"] }),
    false,
  );
  assert.equal(
    canMutateMemory(
      { scope: "parish" },
      { id: "u1", roles: ["ref:parish-admin"] },
    ),
    true,
  );
});

test("agentDisplayMeta and toolDisplayMeta return expected labels", () => {
  assert.equal(agentDisplayMeta("gabriel").name, "Gabriel");
  assert.equal(agentDisplayMeta("mikael").tone, "success");
  assert.equal(agentDisplayMeta("rafael").tone, "warning");

  assert.equal(toolDisplayMeta("org_list").category, "Organisasi");
  assert.equal(toolDisplayMeta("event_create").category, "Kegiatan");
  assert.equal(toolDisplayMeta("memory_save").category, "Memori");
  assert.equal(
    toolDisplayMeta("delegate_specialist").category,
    "Delegasi Spesialis",
  );
});

test("sortMessagesChronologically and buildTimelineItems pair tool calls and results", () => {
  const messages = [
    {
      id: "m3",
      role: "assistant",
      agentName: "gabriel",
      content: "Ada 1 kegiatan.",
      createdAt: "2026-11-22T01:02:00Z",
    },
    {
      id: "m2",
      role: "tool",
      agentName: "mikael",
      toolCallId: "tc1",
      toolName: "event_list",
      content: '{"items":[{"title":"Misa"}]}',
      createdAt: "2026-11-22T01:01:05Z",
    },
    {
      id: "m1",
      role: "assistant",
      agentName: "mikael",
      content: "",
      thinking: "Cek jadwal",
      toolCalls: [{ id: "tc1", name: "event_list", arguments: "{}" }],
      createdAt: "2026-11-22T01:01:00Z",
    },
    {
      id: "m0",
      role: "user",
      agentName: "gabriel",
      content: "Jadwal misa?",
      createdAt: "2026-11-22T01:00:00Z",
    },
  ];

  const sorted = sortMessagesChronologically(messages);
  assert.deepEqual(
    sorted.map((m) => m.id),
    ["m0", "m1", "m2", "m3"],
  );

  const timeline = buildTimelineItems(messages);
  assert.equal(timeline.length, 3);
  assert.equal(timeline[1]?.thinking, "Cek jadwal");
  assert.equal(timeline[1]?.toolSteps.length, 1);
  assert.equal(timeline[1]?.toolSteps[0]?.status, "done");
  assert.equal(
    timeline[1]?.toolSteps[0]?.result,
    '{"items":[{"title":"Misa"}]}',
  );
});

test("parseSSEBuffer parses complete frames and retains partial tail", () => {
  const chunk1 =
    ': keepalive\n\nevent: session\ndata: {"event":"session","sessionId":"s1"}\n\nevent: delta\ndata: {"event":"delta","content":"Halo';
  const r1 = parseSSEBuffer(chunk1);
  assert.equal(r1.events.length, 1);
  assert.equal(r1.events[0]?.event, "session");
  assert.equal(r1.events[0]?.sessionId, "s1");
  assert.equal(
    r1.remainder,
    'event: delta\ndata: {"event":"delta","content":"Halo',
  );

  const r2 = parseSSEBuffer(
    `${r1.remainder} Kak!"}\r\n\r\nevent: done\ndata: {"event":"done"}\r\n\r\n`,
  );
  assert.equal(r2.events.length, 2);
  assert.equal(r2.events[0]?.content, "Halo Kak!");
  assert.equal(r2.events[1]?.event, "done");
  assert.equal(r2.remainder, "");
});
