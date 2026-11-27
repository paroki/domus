import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class ParokiPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: /^(Keuskupan|Paroki|Struktur Gerejawi)$/ })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectForbiddenBanner(): Promise<void> {
    await this.page
      .getByText(/Hanya Super Admin/)
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async createKeuskupan(input: {
    code: string;
    name: string;
    bishopName?: string;
    telepon?: string;
    email?: string;
    website?: string;
    alamat?: string;
  }): Promise<void> {
    await this.goto("/settings/gerejawi/keuskupan/new");
    await this.page.locator("#keuskupan-kode").fill(input.code);
    await this.page.locator("#keuskupan-nama").fill(input.name);
    if (input.bishopName) {
      await this.page.locator("#keuskupan-uskup").fill(input.bishopName);
    }
    if (input.telepon) {
      await this.page.locator("#keuskupan-telepon").fill(input.telepon);
    }
    if (input.email) {
      await this.page.locator("#keuskupan-email").fill(input.email);
    }
    if (input.website) {
      await this.page.locator("#keuskupan-website").fill(input.website);
    }
    if (input.alamat) {
      await this.page.locator("#keuskupan-alamat").fill(input.alamat);
    }
    await this.page.getByRole("button", { name: "Simpan Keuskupan" }).click();
    await this.page.waitForURL(/\/settings\/gerejawi\/keuskupan$/, {
      timeout: 10_000,
    });
  }

  async filterKeuskupanStatus(status: "" | "true" | "false"): Promise<void> {
    await this.page.locator("#keuskupan-status-filter").selectOption(status);
  }

  async expectKeuskupanVisible(name: string, code: string): Promise<void> {
    await this.goto("/settings/gerejawi/keuskupan");
    await this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .filter({ hasText: code })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteKeuskupan(name: string): Promise<void> {
    await this.goto("/settings/gerejawi/keuskupan");
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.getByRole("button", { name: "Hapus Keuskupan" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Ya, Hapus" })
      .click();
    await this.page.waitForURL(/\/settings\/gerejawi\/keuskupan$/, {
      timeout: 10_000,
    });
  }

  async expectKeuskupanNotVisible(name: string): Promise<void> {
    await this.page
      .getByText(name, { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async createParoki(input: {
    keuskupanName?: string;
    kevikepanName?: string;
    code: string;
    name: string;
    timezone?: string;
  }): Promise<void> {
    await this.goto("/settings/gerejawi/paroki/new");
    if (input.keuskupanName) {
      await this.page
        .locator("#wiz-keuskupan")
        .selectOption({ label: input.keuskupanName });
    }
    if (input.kevikepanName) {
      await this.page
        .locator("#wiz-kevikepan")
        .selectOption({ label: input.kevikepanName });
    }
    await this.page.locator("#wiz-nama").fill(input.name);
    await this.page.locator("#wiz-kode").fill(input.code);
    await this.page
      .getByRole("button", { name: "Lanjut ke Identitas →" })
      .click();
    if (input.timezone) {
      await this.page
        .getByRole("button", { name: "Lanjut ke Domisili →" })
        .click();
      await this.page.locator("#wiz-timezone").selectOption(input.timezone);
    }
    await this.page
      .getByRole("button", { name: "Simpan & Buat Paroki" })
      .click();
    await this.page.waitForURL(/\/settings\/gerejawi\/paroki\/[^/]+$/, {
      timeout: 10_000,
    });
  }

  async expectParokiUsersShortcut(): Promise<void> {
    const shortcut = this.page.getByRole("link", {
      name: /Kelola Admin & Anggota Paroki di Direktori Pengguna/,
    });
    await shortcut.waitFor({ state: "visible", timeout: 10_000 });
    assert.equal(await shortcut.getAttribute("href"), "/users");
  }

  async editParokiStruktur(input: { kodeResmi?: string }): Promise<void> {
    const card = this.page
      .locator("div.rounded-xl")
      .filter({
        has: this.page.getByRole("heading", {
          name: "Struktur & Status Gerejawi",
        }),
      })
      .first();
    await card.getByRole("button", { name: "Ubah" }).click();
    if (input.kodeResmi !== undefined) {
      await this.page.locator("#sec-koderesmi").fill(input.kodeResmi);
    }
    await this.page.getByRole("button", { name: "Simpan Perubahan" }).click();
    await this.page
      .getByRole("status")
      .filter({ hasText: "Data paroki berhasil diperbarui." })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectParokiVisible(name: string, code: string): Promise<void> {
    await this.goto("/settings/gerejawi/paroki");
    await this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .filter({ hasText: code })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteParoki(name: string): Promise<void> {
    await this.goto("/settings/gerejawi/paroki");
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .first();
    await card.getByRole("link", { name: /Ubah/ }).click();
    await this.page.getByRole("button", { name: "Hapus Paroki" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Ya, Hapus" })
      .click();
    await this.page.waitForURL(/\/settings\/gerejawi\/paroki$/, {
      timeout: 10_000,
    });
  }

  async expectParokiNotVisible(name: string): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Paroki", exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByText(name, { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
  }
}
