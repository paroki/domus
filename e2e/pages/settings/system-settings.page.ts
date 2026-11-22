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

  async openLauncherItem(label: string): Promise<void> {
    await this.page.getByRole("link", { name: label }).click();
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

  private providerIds = new Map<string, string>();

  private resolveProviderValue(provider: string): string {
    return this.providerIds.get(provider.toLowerCase()) ?? provider;
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
    input: { apiKey: string; baseUrl?: string; name?: string },
  ): Promise<void> {
    const id = providerId.toLowerCase();
    const displayName = input.name ?? (id === "openai" ? "OpenAI" : providerId);

    await this.page.getByRole("button", { name: "Tambah Konfigurasi" }).click();
    await this.page.locator("#provider-name").fill(displayName);
    await this.page.locator("#provider-type").selectOption(id);
    await this.page.locator("#provider-apikey").fill(input.apiKey);
    if (input.baseUrl !== undefined) {
      await this.page.locator("#provider-baseurl").fill(input.baseUrl);
    }

    const [createRes] = await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().endsWith("/api/agent/providers") &&
          r.request().method() === "POST" &&
          r.status() === 201,
      ),
      this.page.getByRole("button", { name: "Simpan Konfigurasi" }).click(),
    ]);
    const body = (await createRes.json()) as { data?: { id?: string } };
    if (body.data?.id) {
      this.providerIds.set(id, body.data.id);
    }
    await this.page
      .getByRole("status")
      .filter({ hasText: "Pengaturan berhasil disimpan." })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async syncProviderModels(
    providerId: string,
    providerLabel: string,
  ): Promise<void> {
    const uuid = this.resolveProviderValue(providerId);
    const section = this.providerSection(providerLabel);
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes(`/api/agent/providers/${uuid}/sync-models`) &&
          r.request().method() === "POST" &&
          r.status() === 200,
      ),
      this.page.waitForResponse(
        (r) =>
          r.url().endsWith("/api/agent/providers") &&
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
        .selectOption(this.resolveProviderValue(input.provider));
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
      assert.equal(val, this.resolveProviderValue(expected.provider));
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
      .selectOption(this.resolveProviderValue(input.provider));
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

  async saveEmbeddingConfig(): Promise<void> {
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/settings") &&
          r.request().method() === "PATCH" &&
          r.status() === 200,
      ),
      this.page
        .getByRole("button", { name: "Simpan Konfigurasi Embedding" })
        .click(),
    ]);
    await this.page
      .getByRole("status")
      .filter({ hasText: "Pengaturan berhasil disimpan." })
      .waitFor({ state: "visible", timeout: 10_000 });
  }
}
