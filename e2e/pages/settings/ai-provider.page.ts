import type { Locator } from "playwright";
import { BasePage } from "../base.page";

export class AIProviderSettingsPage extends BasePage {
  private providerSection(providerLabel: string): Locator {
    return this.page
      .locator("section")
      .filter({
        has: this.page.getByRole("heading", { name: providerLabel }),
      })
      .first();
  }

  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "AI Provider" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteProviderExpectConflict(providerLabel: string): Promise<void> {
    const section = this.providerSection(providerLabel);
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/agent/providers/") &&
          r.request().method() === "DELETE" &&
          r.status() === 409,
      ),
      section.getByRole("button", { name: "Hapus" }).click(),
    ]);
    await this.page
      .getByRole("alert")
      .filter({ hasText: "sedang digunakan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }
}
