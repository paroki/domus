import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class SettingsPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Pengaturan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectVisibleTiles(labels: string[]): Promise<void> {
    const nav = this.page.getByRole("navigation", {
      name: "Pengaturan",
    });
    await nav.waitFor({ state: "visible", timeout: 10_000 });
    for (const label of labels) {
      await nav
        .getByRole("link", { name: label })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectHiddenTiles(labels: string[]): Promise<void> {
    const nav = this.page.getByRole("navigation", {
      name: "Pengaturan",
    });
    for (const label of labels) {
      const count = await nav.getByRole("link", { name: label }).count();
      assert.equal(count, 0, `Expected hub "${label}" to be hidden`);
    }
  }

  async openTile(label: string): Promise<void> {
    await this.page
      .getByRole("navigation", { name: "Pengaturan" })
      .getByRole("link", { name: label })
      .click();
  }

  async expectParokiAdminRestrictionBanner(): Promise<void> {
    await this.page
      .getByText(/Hanya Admin Paroki/)
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectEmptySettings(): Promise<void> {
    await this.page
      .getByText("Anda tidak memiliki izin untuk mengakses pengaturan.")
      .waitFor({ state: "visible", timeout: 10_000 });
  }
}
