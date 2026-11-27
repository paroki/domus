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
    await section.getByRole("link", { name: "Kelola" }).click();
    await this.page
      .getByRole("button", { name: "Hapus", exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page.getByRole("button", { name: "Hapus", exact: true }).click();
    const dialog = this.page.getByRole("dialog");
    await dialog.waitFor({ state: "visible", timeout: 10_000 });
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/agent/providers/") &&
          r.request().method() === "DELETE" &&
          r.status() === 409,
      ),
      dialog.getByRole("button", { name: "Hapus", exact: true }).click(),
    ]);
    await this.page
      .getByRole("alert")
      .filter({ hasText: "sedang digunakan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }
}
