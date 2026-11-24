import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class HomePage extends BasePage {
  async expectRendered(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Selamat datang di Domus" })
      .waitFor({ state: "visible", timeout: 10_000 });

    const menus = ["Organisasi", "Kegiatan", "Referensi", "Pengguna"];
    for (const label of menus) {
      await this.page
        .getByRole("link", { name: new RegExp(label) })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectVisibleTiles(labels: string[]): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Selamat datang di Domus" })
      .waitFor({ state: "visible", timeout: 10_000 });
    for (const label of labels) {
      await this.page
        .getByRole("link", { name: new RegExp(label) })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectHiddenTiles(labels: string[]): Promise<void> {
    for (const label of labels) {
      const count = await this.page
        .getByRole("link", { name: new RegExp(label) })
        .count();
      assert.equal(count, 0, `Expected launcher tile "${label}" to be hidden`);
    }
  }

  async openTile(label: string): Promise<void> {
    await this.page.getByRole("link", { name: new RegExp(label) }).click();
  }
}
