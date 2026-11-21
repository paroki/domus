import assert from "node:assert/strict";
import type { Locator } from "playwright";
import { BasePage } from "../base.page";

export class SystemSettingsPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Pengaturan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectForbiddenBanner(): Promise<void> {
    await this.page
      .getByRole("alert")
      .filter({
        hasText:
          "Hanya Super Admin yang dapat melihat dan mengubah pengaturan sistem.",
      })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async selectTab(
    tab: "Kanal Komunikasi" | "Penyedia LLM" | "Agen & Embedding",
  ): Promise<void> {
    await this.page.getByRole("tab", { name: tab }).click();
  }

  async expectChannelGroupsVisible(groupNames: string[]): Promise<void> {
    for (const name of groupNames) {
      await this.page
        .getByRole("heading", { name })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectNoRawAgentKeysInChannels(): Promise<void> {
    const count = await this.page.locator('[id^="setting-agent."]').count();
    assert.equal(count, 0, "Expected no agent.* inputs in Channels tab");
  }

  private providerSection(providerLabel: string): Locator {
    return this.page
      .locator("section")
      .filter({
        has: this.page.getByRole("heading", { name: providerLabel }),
      })
      .first();
  }

  async configureProvider(
    providerId: string,
    input: { apiKey: string; baseUrl?: string },
  ): Promise<void> {
    const id = providerId.toLowerCase();
    await this.page
      .locator(`[id="setting-agent.provider.${id}.api_key"]`)
      .fill(input.apiKey);
    if (input.baseUrl !== undefined) {
      await this.page
        .locator(`[id="setting-agent.provider.${id}.base_url"]`)
        .fill(input.baseUrl);
    }
  }

  async syncProviderModels(
    providerId: string,
    providerLabel: string,
  ): Promise<void> {
    const section = this.providerSection(providerLabel);
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r
            .url()
            .includes(
              `/api/agent/providers/${providerId.toLowerCase()}/sync-models`,
            ) &&
          r.request().method() === "POST" &&
          r.status() === 200,
      ),
      this.page.waitForResponse(
        (r) =>
          r.url().endsWith("/api/settings") &&
          r.request().method() === "GET" &&
          r.status() === 200,
      ),
      section.getByRole("button", { name: "Sinkronkan Model" }).click(),
    ]);
    await this.page
      .getByRole("status")
      .filter({ hasText: "berhasil disinkronkan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectProviderModelsCatalog(
    providerLabel: string,
    expectedModelIds: string[],
  ): Promise<void> {
    const section = this.providerSection(providerLabel);
    const details = section.locator("details");
    const isOpen = await details.getAttribute("open");
    if (isOpen === null) {
      await details.locator("summary").click();
    }
    for (const modelId of expectedModelIds) {
      await details
        .getByText(modelId, { exact: true })
        .first()
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  private async setControlValue(
    locator: Locator,
    value: string,
  ): Promise<void> {
    const tagName = await locator.evaluate((el) => el.tagName.toLowerCase());
    if (tagName === "select") {
      await locator.selectOption(value);
    } else {
      await locator.fill(value);
    }
  }

  async configureAgentCard(
    agentName: string,
    input: {
      provider?: string;
      model?: string;
      effort?: string;
      instruction?: string;
    },
  ): Promise<void> {
    const id = agentName.toLowerCase();
    if (input.provider !== undefined) {
      await this.page
        .locator(`[id="setting-agent.${id}.provider"]`)
        .selectOption(input.provider);
    }
    if (input.model !== undefined) {
      await this.setControlValue(
        this.page.locator(`[id="setting-agent.${id}.model"]`),
        input.model,
      );
    }
    if (input.effort !== undefined) {
      await this.page
        .locator(`[id="setting-agent.${id}.effort"]`)
        .selectOption(input.effort);
    }
    if (input.instruction !== undefined) {
      await this.page
        .locator(`[id="setting-agent.${id}.instruction"]`)
        .fill(input.instruction);
    }
  }

  async expectAgentCardValues(
    agentName: string,
    expected: {
      provider?: string;
      model?: string;
      effort?: string;
      instructionIncludes?: string;
    },
  ): Promise<void> {
    const id = agentName.toLowerCase();
    if (expected.provider !== undefined) {
      const val = await this.page
        .locator(`[id="setting-agent.${id}.provider"]`)
        .inputValue();
      assert.equal(val, expected.provider);
    }
    if (expected.model !== undefined) {
      const val = await this.page
        .locator(`[id="setting-agent.${id}.model"]`)
        .inputValue();
      assert.equal(val, expected.model);
    }
    if (expected.effort !== undefined) {
      const val = await this.page
        .locator(`[id="setting-agent.${id}.effort"]`)
        .inputValue();
      assert.equal(val, expected.effort);
    }
    if (expected.instructionIncludes !== undefined) {
      const val = await this.page
        .locator(`[id="setting-agent.${id}.instruction"]`)
        .inputValue();
      assert.ok(val.includes(expected.instructionIncludes));
    }
  }

  async resetAgentInstruction(agentName: string): Promise<void> {
    const section = this.page
      .locator("section")
      .filter({
        has: this.page.getByRole("heading", { name: agentName, exact: true }),
      })
      .first();
    await section
      .getByRole("button", { name: "Reset Instruksi Bawaan" })
      .click();
    await this.page
      .getByRole("status")
      .filter({ hasText: "berhasil dikembalikan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async configureEmbeddingCard(input: {
    provider: string;
    model?: string;
  }): Promise<void> {
    await this.page
      .locator('[id="setting-agent.embedding.provider"]')
      .selectOption(input.provider);
    if (input.model !== undefined) {
      await this.setControlValue(
        this.page.locator('[id="setting-agent.embedding.model"]'),
        input.model,
      );
    }
  }

  async saveAgentConfig(): Promise<void> {
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/settings") &&
          r.request().method() === "PATCH" &&
          r.status() === 200,
      ),
      this.page
        .getByRole("button", { name: "Simpan Konfigurasi Agen" })
        .click(),
    ]);
    await this.page
      .getByRole("status")
      .filter({ hasText: "Pengaturan berhasil disimpan." })
      .waitFor({ state: "visible", timeout: 10_000 });
  }
}
