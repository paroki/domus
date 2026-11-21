import assert from "node:assert/strict";
import { Then, When } from "@cucumber/cucumber";
import type { DomusWorld } from "../../support/world";

Then(
  "I see the system settings forbidden banner",
  async function (this: DomusWorld) {
    await this.pages.settings.expectForbiddenBanner();
  },
);

Then("I see the system settings page", async function (this: DomusWorld) {
  await this.pages.settings.expectLoaded();
});

Then(
  "I see channel groups {string} without raw agent keys",
  async function (this: DomusWorld, rawGroups: string) {
    const names = rawGroups.split(",").map((s) => s.trim());
    await this.pages.settings.expectChannelGroupsVisible(names);
    await this.pages.settings.expectNoRawAgentKeysInChannels();
  },
);

When(
  "I switch settings tab to {string}",
  async function (this: DomusWorld, tab: string) {
    await this.pages.settings.selectTab(
      tab as "Kanal Komunikasi" | "Penyedia LLM" | "Agen & Embedding",
    );
  },
);

When(
  "I configure provider {string} with key {string} and mock LLM base URL",
  async function (this: DomusWorld, providerId: string, apiKey: string) {
    assert.ok(this.mockLlmBaseUrl, "Expected mockLlmBaseUrl");
    await this.pages.settings.configureProvider(providerId, {
      apiKey,
      baseUrl: this.mockLlmBaseUrl,
    });
  },
);

When(
  "I sync models for provider {string} labeled {string}",
  async function (this: DomusWorld, providerId: string, label: string) {
    await this.pages.settings.syncProviderModels(providerId, label);
  },
);

Then(
  "I see provider {string} model catalog containing {string}",
  async function (this: DomusWorld, label: string, rawIds: string) {
    const ids = rawIds.split(",").map((s) => s.trim());
    await this.pages.settings.expectProviderModelsCatalog(label, ids);
  },
);

When(
  "I configure agent {string} with provider {string}, model {string}, effort {string}, and instruction {string}",
  async function (
    this: DomusWorld,
    agentName: string,
    provider: string,
    model: string,
    effort: string,
    instruction: string,
  ) {
    await this.pages.settings.configureAgentCard(agentName, {
      provider,
      model,
      effort,
      instruction,
    });
  },
);

When(
  "I configure agent {string} with provider {string}, model {string}, and effort {string}",
  async function (
    this: DomusWorld,
    agentName: string,
    provider: string,
    model: string,
    effort: string,
  ) {
    await this.pages.settings.configureAgentCard(agentName, {
      provider,
      model,
      effort,
    });
  },
);

When(
  "I configure embedding with provider {string} and model {string}",
  async function (this: DomusWorld, provider: string, model: string) {
    await this.pages.settings.configureEmbeddingCard({ provider, model });
  },
);

When("I save agent configuration", async function (this: DomusWorld) {
  await this.pages.settings.saveAgentConfig();
});

Then(
  "I see agent {string} configured with provider {string}, model {string}, effort {string}, and instruction {string}",
  async function (
    this: DomusWorld,
    agentName: string,
    provider: string,
    model: string,
    effort: string,
    instruction: string,
  ) {
    await this.pages.settings.expectAgentCardValues(agentName, {
      provider,
      model,
      effort,
      instructionIncludes: instruction,
    });
  },
);

When(
  "I reset instruction for agent {string}",
  async function (this: DomusWorld, agentName: string) {
    await this.pages.settings.resetAgentInstruction(agentName);
  },
);

Then(
  "I see agent {string} instruction restored from default",
  async function (this: DomusWorld, agentName: string) {
    await this.pages.settings.expectAgentCardValues(agentName, {
      instructionIncludes: "Gabriel",
    });
  },
);
