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
  "I open settings launcher {string}",
  async function (this: DomusWorld, label: string) {
    await this.pages.settings.openLauncherItem(label);
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
