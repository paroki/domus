import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class AIAgentSettingsPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "AI Agent", exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .locator('[id="setting-agent.model.gabriel"]')
      .waitFor({ state: "visible", timeout: 10_000 });
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

  async bindModels(models: {
    gabriel: string;
    mikael: string;
    rafael: string;
    embedding: string;
  }): Promise<void> {
    await this.selectModelBySuffix("agent.model.gabriel", models.gabriel);
    await this.selectModelBySuffix("agent.model.mikael", models.mikael);
    await this.selectModelBySuffix("agent.model.rafael", models.rafael);
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

  private gabrielPromptBlock() {
    return this.page
      .locator("div.flex.flex-col")
      .filter({ has: this.page.locator('[id="setting-agent.prompt.gabriel"]') })
      .last();
  }

  async customizeGabrielPrompt(text: string): Promise<void> {
    await this.page.locator('[id="setting-agent.prompt.gabriel"]').fill(text);
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/settings") &&
          r.request().method() === "PATCH" &&
          r.status() === 200,
      ),
      this.page.getByRole("button", { name: "Simpan Prompt Persona" }).click(),
    ]);
    await this.gabrielPromptBlock()
      .getByText("Overridden", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async resetGabrielPromptToDefault(): Promise<void> {
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/settings") &&
          r.request().method() === "PATCH" &&
          r.status() === 200,
      ),
      this.gabrielPromptBlock()
        .getByRole("button", { name: "Reset ke Default" })
        .click(),
    ]);
    await this.gabrielPromptBlock()
      .getByText("Overridden", { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
    const val = await this.page
      .locator('[id="setting-agent.prompt.gabriel"]')
      .inputValue();
    assert.ok(
      val.includes("Gabriel"),
      "Expected default Gabriel prompt restored",
    );
  }
}
