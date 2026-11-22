import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import { loginAsSuperAdmin } from "../../support/auth";
import { TestClient } from "../../support/test-client";
import type { DomusWorld } from "../../support/world";

Then("I see the AI Assistant workspace", async function (this: DomusWorld) {
  await this.pages.agent.expectLoaded();
});

Given(
  "I have two agent sessions {string} and {string}",
  async function (this: DomusWorld, title1: string, title2: string) {
    const r1 = await this.client.post("/agent/sessions", { title: title1 });
    assert.equal(r1.status, 201);
    const r2 = await this.client.post("/agent/sessions", { title: title2 });
    assert.equal(r2.status, 201);
  },
);

Then(
  "I see agent session {string} in sidebar",
  async function (this: DomusWorld, title: string) {
    await this.pages.agent.expectSessionInSidebar(title);
  },
);

Then(
  "I do not see agent session {string} in sidebar",
  async function (this: DomusWorld, title: string) {
    await this.pages.agent.expectSessionNotInSidebar(title);
  },
);

When(
  "I rename agent session {string} to {string}",
  async function (this: DomusWorld, oldTitle: string, newTitle: string) {
    await this.pages.agent.renameSession(oldTitle, newTitle);
  },
);

When(
  "I search agent sessions for {string}",
  async function (this: DomusWorld, q: string) {
    await this.pages.agent.searchSessions(q);
  },
);

When(
  "I delete agent session {string}",
  async function (this: DomusWorld, title: string) {
    await this.pages.agent.deleteSession(title);
  },
);

Given(
  "the parish has a shared parish memory {string}",
  async function (this: DomusWorld, content: string) {
    assert.ok(this.parishAdminToken && this.activeParishId);
    const adminClient = new TestClient();
    adminClient.setToken(this.parishAdminToken);
    adminClient.setParishId(this.activeParishId);
    const res = await adminClient.post("/agent/memories", {
      scope: "parish",
      category: "parish_faq",
      content,
      isPinned: true,
    });
    assert.equal(res.status, 201);
  },
);

When("I open the AI Memory Drawer", async function (this: DomusWorld) {
  await this.pages.agent.openMemoryDrawer();
});

When(
  "I create a personal memory in category {string} with content {string} and pinned true",
  async function (this: DomusWorld, category: string, content: string) {
    await this.pages.agent.createMemory({
      category,
      content,
      isPinned: true,
    });
  },
);

Then(
  "I see memory {string}",
  async function (this: DomusWorld, content: string) {
    await this.pages.agent.expectMemoryVisible(content);
  },
);

Then(
  "I do not see memory {string}",
  async function (this: DomusWorld, content: string) {
    await this.pages.agent.expectMemoryNotVisible(content);
  },
);

When(
  "I edit memory {string} to {string}",
  async function (this: DomusWorld, oldContent: string, newContent: string) {
    await this.pages.agent.editMemory(oldContent, { content: newContent });
  },
);

When(
  "I switch memory scope tab to {string}",
  async function (this: DomusWorld, tabName: string) {
    await this.pages.agent.switchMemoryScopeTab(
      tabName as "Memori Pribadi" | "FAQ & Pengetahuan Paroki",
    );
  },
);

Then(
  "I see the parish memory read-only notice",
  async function (this: DomusWorld) {
    await this.pages.agent.expectParishMemoryReadOnlyNotice();
  },
);

When(
  "I delete memory {string}",
  async function (this: DomusWorld, content: string) {
    await this.pages.agent.deleteMemory(content);
  },
);

Given(
  "AI agent OpenAI provider is configured against the mock LLM server",
  async function (this: DomusWorld) {
    assert.ok(this.mockLlmBaseUrl, "Expected mockLlmBaseUrl");
    const superSession = await loginAsSuperAdmin();
    const superClient = new TestClient();
    superClient.setToken(superSession.rawToken);
    const provRes = await superClient.post<{ data: { id: string } }>(
      "/agent/providers",
      {
        name: "OpenAI E2E",
        provider: "openai",
        apiKey: "e2e-mock-key",
        baseUrl: this.mockLlmBaseUrl,
      },
    );
    assert.equal(provRes.status, 201);
    const providerId = provRes.data.data.id;
    const res = await superClient.patch("/settings", {
      values: {
        "agent.gabriel.provider_id": providerId,
        "agent.gabriel.model": "gpt-4.1-mini",
        "agent.mikael.provider_id": providerId,
        "agent.mikael.model": "gpt-4.1-mini",
        "agent.rafael.provider_id": providerId,
        "agent.rafael.model": "gpt-4.1-mini",
        "agent.embedding.provider_id": "none",
      },
    });
    assert.equal(res.status, 200);
  },
);

When(
  "I send AI chat prompt {string}",
  async function (this: DomusWorld, prompt: string) {
    await this.pages.agent.sendMessage(prompt);
  },
);

Then(
  "I see user chat message {string}",
  async function (this: DomusWorld, text: string) {
    await this.pages.agent.expectUserMessageVisible(text);
  },
);

Then(
  "I see thinking indicator for {string}",
  async function (this: DomusWorld, agentLabel: string) {
    await this.pages.agent.expectThinkingBlockVisible(agentLabel);
  },
);

Then(
  "I see expandable tool execution badge for {string}",
  async function (this: DomusWorld, toolName: string) {
    await this.pages.agent.expectToolExecutionBadge(toolName);
  },
);

Then(
  "I see assistant reply containing {string}",
  async function (this: DomusWorld, text: string) {
    await this.pages.agent.expectAssistantMessageContaining(text);
  },
);
