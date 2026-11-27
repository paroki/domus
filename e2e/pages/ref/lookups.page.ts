import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class LookupsPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: /Kamus/ })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectForbiddenBanner(): Promise<void> {
    await this.page
      .getByText(
        "Hanya Admin Paroki atau Super Admin yang dapat mengelola kamus referensi.",
      )
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async createLookup(input: {
    category: string;
    code: string;
    label: string;
    sortOrder?: number;
  }): Promise<void> {
    await this.goto(
      `/settings/lookups/new?category=${encodeURIComponent(input.category)}`,
    );
    await this.page.locator("#lookup-category").selectOption(input.category);
    await this.page.locator("#lookup-label").fill(input.label);
    await this.page.locator("#lookup-code").fill(input.code);
    if (input.sortOrder !== undefined) {
      await this.page.locator("#lookup-sort").fill(String(input.sortOrder));
    }
    await this.page.getByRole("button", { name: "Simpan Item" }).click();
    await this.page.waitForURL(/\/settings\/lookups\?category=/, {
      timeout: 10_000,
    });
  }

  async editLookup(
    currentLabel: string,
    input: { label: string; sortOrder?: number },
  ): Promise<void> {
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ has: this.page.getByText(currentLabel, { exact: true }) })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.locator("#lookup-label").fill(input.label);
    if (input.sortOrder !== undefined) {
      await this.page.locator("#lookup-sort").fill(String(input.sortOrder));
    }
    await this.page.getByRole("button", { name: "Simpan Item" }).click();
    await this.page.waitForURL(/\/settings\/lookups\?category=/, {
      timeout: 10_000,
    });
  }

  async toggleLookupActive(label: string): Promise<void> {
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ has: this.page.getByText(label, { exact: true }) })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.getByLabel("Item Aktif").click();
    await this.page.getByRole("button", { name: "Simpan Item" }).click();
    await this.page.waitForURL(/\/settings\/lookups\?category=/, {
      timeout: 10_000,
    });
  }

  async deleteLookup(label: string): Promise<void> {
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ has: this.page.getByText(label, { exact: true }) })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.getByRole("button", { name: "Hapus Item" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Ya, Hapus" })
      .click();
    await this.page.waitForURL(/\/settings\/lookups\?category=/, {
      timeout: 10_000,
    });
  }

  async expectLookupVisible(label: string, scopeLabel: string): Promise<void> {
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ has: this.page.getByText(label, { exact: true }) })
      .first();
    await card.waitFor({ state: "visible", timeout: 10_000 });
    const scopeText = await card.getByText(scopeLabel, { exact: true }).count();
    assert.ok(scopeText > 0, `Expected scope badge "${scopeLabel}"`);
  }

  async expectLookupNotVisible(label: string): Promise<void> {
    await this.page
      .getByText(label, { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async expectLookupReadOnly(label: string): Promise<void> {
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ has: this.page.getByText(label, { exact: true }) })
      .first();
    await card.waitFor({ state: "visible", timeout: 10_000 });
    assert.equal(
      await card.getByRole("link", { name: "Ubah" }).count(),
      0,
      `Expected read-only lookup "${label}" to have no edit link`,
    );
  }

  async expectCategoryNotVisible(label: string): Promise<void> {
    await this.expectLoaded();
    assert.equal(
      await this.page.getByText(label, { exact: true }).count(),
      0,
      `Expected category "${label}" not to be listed on /settings/lookups`,
    );
  }
}
