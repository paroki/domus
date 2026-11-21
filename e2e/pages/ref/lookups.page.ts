import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class LookupsPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Kamus Referensi" })
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
    await this.page
      .getByPlaceholder("Kategori (mis. sacrament)")
      .fill(input.category);
    await this.page.getByPlaceholder("Kode (mis. kaj)").fill(input.code);
    await this.page.getByPlaceholder("Label tampilan").fill(input.label);
    if (input.sortOrder !== undefined) {
      await this.page.getByPlaceholder("Urutan").fill(String(input.sortOrder));
    }
    await this.page.getByRole("button", { name: "Tambah Referensi" }).click();
  }

  async editLookup(
    currentLabel: string,
    input: { label: string; sortOrder?: number },
  ): Promise<void> {
    await this.page
      .getByRole("button", { name: `Ubah ${currentLabel}` })
      .click();
    await this.page.getByPlaceholder("Label tampilan").fill(input.label);
    if (input.sortOrder !== undefined) {
      await this.page.getByPlaceholder("Urutan").fill(String(input.sortOrder));
    }
    await this.page.getByRole("button", { name: "Simpan Perubahan" }).click();
  }

  async toggleLookupActive(label: string): Promise<void> {
    const row = this.page
      .locator("section > div.divide-y > div")
      .filter({ has: this.page.getByText(label, { exact: true }) });
    const btn = row.locator("button[title]").first();
    const prevTitle = await btn.getAttribute("title");
    await btn.click();
    const expectedTitle = prevTitle === "Aktif" ? "Nonaktif" : "Aktif";
    await this.page.waitForFunction(
      ([el, expected]) => el?.getAttribute("title") === expected,
      [await btn.elementHandle(), expectedTitle] as const,
      { timeout: 10_000 },
    );
  }

  async deleteLookup(label: string): Promise<void> {
    await this.page.getByRole("button", { name: `Hapus ${label}` }).click();
  }

  async expectLookupVisible(label: string, scopeLabel: string): Promise<void> {
    const row = this.page
      .locator("section > div.divide-y > div")
      .filter({ has: this.page.getByText(label, { exact: true }) });
    await row.waitFor({ state: "visible", timeout: 10_000 });
    const scopeText = await row.getByText(scopeLabel, { exact: true }).count();
    assert.ok(scopeText > 0, `Expected scope badge "${scopeLabel}"`);
  }

  async expectLookupNotVisible(label: string): Promise<void> {
    const row = this.page
      .locator("section > div.divide-y > div")
      .filter({ has: this.page.getByText(label, { exact: true }) });
    await row.waitFor({ state: "detached", timeout: 10_000 });
  }

  async expectLookupReadOnly(label: string): Promise<void> {
    const row = this.page
      .locator("section > div.divide-y > div")
      .filter({ has: this.page.getByText(label, { exact: true }) });
    await row.waitFor({ state: "visible", timeout: 10_000 });
    assert.equal(
      await row.getByRole("button", { name: `Ubah ${label}` }).count(),
      0,
      `Expected read-only lookup "${label}" to have no edit button`,
    );
    assert.equal(
      await row.getByRole("button", { name: `Hapus ${label}` }).count(),
      0,
      `Expected read-only lookup "${label}" to have no delete button`,
    );
  }
}
