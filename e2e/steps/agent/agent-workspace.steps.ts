import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import { loginAsSuperAdmin } from "../../support/auth";
import { TestClient } from "../../support/test-client";
import type { DomusWorld } from "../../support/world";

Given(
  "AI provider and agent models are configured with mock LLM",
  async function (this: DomusWorld) {
    assert.ok(this.mockLlmBaseUrl, "Expected mockLlmBaseUrl");
    const adminSession = await loginAsSuperAdmin();
    const adminClient = new TestClient();
    adminClient.setToken(adminSession.rawToken);

    const prov = await adminClient.post<{ data: { id: string } }>(
      "/agent/providers",
      {
        name: "E2E OpenAI",
        provider: "openai",
        apiKey: "e2e-secret",
        baseUrl: this.mockLlmBaseUrl,
      },
    );
    assert.equal(prov.status, 201, "Expected 201 creating AI provider");
    const provId = prov.data.data.id;
    assert.ok(provId, "Expected provider id");

    const syncRes = await adminClient.post(
      `/agent/providers/${provId}/sync-models`,
    );
    assert.equal(syncRes.status, 200, "Expected 200 syncing models");

    const patchRes = await adminClient.patch("/settings", {
      values: {
        "agent.model.gabriel": `${provId}:gpt-4.1-mini`,
        "agent.model.mikael": `${provId}:gpt-4.1`,
        "agent.model.rafael": `${provId}:gpt-4.1`,
        "agent.model.yosef": `${provId}:gpt-4.1`,
        "agent.model.gregorius": `${provId}:gpt-4.1`,
        "agent.model.sesilia": `${provId}:gpt-4.1`,
        "agent.model.memory_extractor": `${provId}:gpt-4.1-mini`,
        "agent.model.embedding": `${provId}:text-embedding-3-small`,
      },
    });
    assert.equal(patchRes.status, 200, "Expected 200 updating agent settings");
  },
);

Then("I see the AI Assistant workspace", async function (this: DomusWorld) {
  await this.pages.agent.workspace.expectLoaded();
});

When(
  "I send AI chat message {string}",
  async function (this: DomusWorld, text: string) {
    await this.pages.agent.workspace.sendMessage(text);
  },
);

Then(
  "I see AI agent handoff from {string} to {string}",
  async function (this: DomusWorld, from: string, to: string) {
    await this.pages.agent.workspace.expectHandoff(from, to);
  },
);

Then(
  "I see AI tool badge {string}",
  async function (this: DomusWorld, toolName: string) {
    await this.pages.agent.workspace.expectToolBadge(toolName);
  },
);

Then(
  "I see AI assistant reply containing {string}",
  async function (this: DomusWorld, text: string) {
    await this.pages.agent.workspace.expectAssistantReply(text);
  },
);

Then(
  "the AI session URL persists across page reload with reply {string}",
  async function (this: DomusWorld, text: string) {
    await this.pages.agent.workspace.expectSessionUrlAndReload(text);
  },
);

When(
  "I send slow AI chat message {string} and click Hentikan",
  async function (this: DomusWorld, text: string) {
    await this.pages.agent.workspace.sendSlowStreamAndAbort(text);
  },
);

When(
  "I rename the active AI session to {string}",
  async function (this: DomusWorld, newTitle: string) {
    await this.pages.agent.workspace.renameActiveSession(newTitle);
  },
);

When(
  "I delete the AI session {string}",
  async function (this: DomusWorld, title: string) {
    await this.pages.agent.workspace.deleteActiveSession(title);
  },
);

When("I open the AI memory drawer", async function (this: DomusWorld) {
  await this.pages.agent.workspace.openMemoryDrawer();
});

When(
  "I add an AI memory {string}",
  async function (this: DomusWorld, content: string) {
    await this.pages.agent.workspace.addMemory(content);
  },
);

When(
  "I search AI memories for {string} expecting {string}",
  async function (this: DomusWorld, query: string, expected: string) {
    await this.pages.agent.workspace.searchMemory(query, expected);
  },
);

When(
  "I edit AI memory {string} to {string}",
  async function (this: DomusWorld, oldText: string, newContent: string) {
    await this.pages.agent.workspace.editMemory(oldText, newContent);
  },
);

When(
  "I delete AI memory {string}",
  async function (this: DomusWorld, text: string) {
    await this.pages.agent.workspace.deleteMemory(text);
  },
);

When(
  "I filter AI memories by category {string}",
  async function (this: DomusWorld, categoryLabel: string) {
    await this.pages.agent.workspace.filterMemoryByCategory(categoryLabel);
  },
);

When(
  "I click AI quick-reply option {string}",
  async function (this: DomusWorld, label: string) {
    await this.pages.agent.workspace.clickQuickReplyOption(label);
  },
);

When("I open the AI bottom dock", async function (this: DomusWorld) {
  await this.pages.agent.workspace.openBottomDock();
});

Then("the AI bottom dock remains open", async function (this: DomusWorld) {
  await this.pages.agent.workspace.expectBottomDockOpen();
});

When(
  "I send AI bottom dock message {string}",
  async function (this: DomusWorld, text: string) {
    await this.pages.agent.workspace.sendDockMessage(text);
  },
);

Then(
  "I see AI bottom dock reply containing {string}",
  async function (this: DomusWorld, text: string) {
    await this.pages.agent.workspace.expectDockReply(text);
  },
);

When(
  "I click Buka Penuh in the AI bottom dock",
  async function (this: DomusWorld) {
    await this.pages.agent.workspace.clickDockOpenFull();
  },
);
