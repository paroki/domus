import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class SettingsPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Pengaturan Paroki" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectVisibleTiles(labels: string[]): Promise<void> {
    const nav = this.page.getByRole("navigation", {
      name: "Pengaturan Paroki",
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
      name: "Pengaturan Paroki",
    });
    for (const label of labels) {
      const count = await nav.getByRole("link", { name: label }).count();
      assert.equal(count, 0, `Expected tile "${label}" to be hidden`);
    }
  }

  async filterByQuery(q: string): Promise<void> {
    await this.page
      .getByRole("searchbox", { name: "Cari pengaturan…" })
      .fill(q);
  }

  async openTile(label: string): Promise<void> {
    await this.page
      .getByRole("navigation", { name: "Pengaturan Paroki" })
      .getByRole("link", { name: label })
      .click();
  }

  async expectParishAdminRestrictionBanner(): Promise<void> {
    await this.page
      .getByText("Hanya Admin Paroki yang dapat mengubah profil paroki ini.")
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectEmptySettings(): Promise<void> {
    await this.page
      .getByText("Tidak ada menu pengaturan yang tersedia.")
      .waitFor({ state: "visible", timeout: 10_000 });
  }
}
