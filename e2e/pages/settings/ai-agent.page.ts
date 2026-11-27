import assert from "node:assert/strict";
import type { Locator } from "playwright";
import { BasePage } from "../base.page";

export class AIAgentSettingsPage extends BasePage {
  private modelRoutingSection(): Locator {
    return this.page
      .locator("section")
      .filter({
        has: this.page.getByRole("heading", {
          name: "Routing Model AI Agent",
        }),
      })
      .first();
  }

  private promptCard(agentKey: string): Locator {
    return this.page.locator(`[data-agent-prompt="agent.prompt.${agentKey}"]`);
  }

  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "AI Agent", exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.modelRoutingSection().waitFor({
      state: "visible",
      timeout: 10_000,
    });
  }

  private async selectModelBySuffix(
    settingKey: string,
    modelId: string,
  ): Promise<void> {
    const select = this.page.locator(`[id="setting-${settingKey}"]`);
    const optionVal = await select.evaluate((el, suffix) => {
      const s = el as HTMLSelectElement;
      const opt = Array.from(s.options).find((o) =>
        o.value.endsWith(`:${suffix}`),
      );
      return opt?.value ?? "";
    }, modelId);
    assert.ok(
      optionVal,
      `Expected option ending with :${modelId} on ${settingKey}`,
    );
    await select.selectOption(optionVal);
  }

  async expectAllAgentControlsVisible(): Promise<void> {
    const modelAgents = [
      "gabriel",
      "mikael",
      "rafael",
      "yosef",
      "gregorius",
      "sesilia",
      "memory_extractor",
      "embedding",
    ];
    await this.modelRoutingSection()
      .getByRole("button", { name: "Ubah", exact: true })
      .click();
    for (const agent of modelAgents) {
      await this.page
        .locator(`[id="setting-agent.model.${agent}"]`)
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    await this.modelRoutingSection()
      .getByRole("button", { name: "Batal", exact: true })
      .click();

    const promptAgents = [
      "gabriel",
      "mikael",
      "rafael",
      "yosef",
      "gregorius",
      "sesilia",
      "memory_extractor",
    ];
    for (const agent of promptAgents) {
      await this.promptCard(agent)
        .getByRole("button", { name: "Ubah Prompt", exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async bindModels(models: {
    gabriel: string;
    mikael: string;
    rafael: string;
    yosef?: string;
    gregorius?: string;
    sesilia?: string;
    memory_extractor?: string;
    embedding: string;
  }): Promise<void> {
    await this.modelRoutingSection()
      .getByRole("button", { name: "Ubah", exact: true })
      .click();
    await this.selectModelBySuffix("agent.model.gabriel", models.gabriel);
    await this.selectModelBySuffix("agent.model.mikael", models.mikael);
    await this.selectModelBySuffix("agent.model.rafael", models.rafael);
    if (models.yosef) {
      await this.selectModelBySuffix("agent.model.yosef", models.yosef);
    }
    if (models.gregorius) {
      await this.selectModelBySuffix("agent.model.gregorius", models.gregorius);
    }
    if (models.sesilia) {
      await this.selectModelBySuffix("agent.model.sesilia", models.sesilia);
    }
    if (models.memory_extractor) {
      await this.selectModelBySuffix(
        "agent.model.memory_extractor",
        models.memory_extractor,
      );
    }
    await this.selectModelBySuffix("agent.model.embedding", models.embedding);

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

  async customizeGabrielPrompt(text: string): Promise<void> {
    const card = this.promptCard("gabriel");
    await card
      .getByRole("button", { name: "Ubah Prompt", exact: true })
      .click();
    await this.page.locator('[id="setting-agent.prompt.gabriel"]').fill(text);
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/settings") &&
          r.request().method() === "PATCH" &&
          r.status() === 200,
      ),
      card
        .getByRole("button", { name: "Simpan Prompt Persona", exact: true })
        .click(),
    ]);
    await card
      .getByText("Overridden", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async resetGabrielPromptToDefault(): Promise<void> {
    const card = this.promptCard("gabriel");
    await card
      .getByRole("button", { name: "Reset ke Default", exact: true })
      .click();
    const dialog = this.page.getByRole("dialog");
    await dialog.waitFor({ state: "visible", timeout: 10_000 });
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/settings") &&
          r.request().method() === "PATCH" &&
          r.status() === 200,
      ),
      dialog
        .getByRole("button", { name: "Reset ke Default", exact: true })
        .click(),
    ]);
    await card
      .getByText("Default", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
    const preview = await card.locator("p").first().textContent();
    assert.ok(
      preview?.includes("Gabriel"),
      "Expected default Gabriel prompt restored",
    );
  }
}
