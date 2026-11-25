import { BasePage } from "../base.page";

export class KevikepanSipilPage extends BasePage {
  async createKevikepan(input: {
    keuskupanName: string;
    code: string;
    name: string;
    vikep?: string;
  }): Promise<void> {
    await this.goto("/settings/gerejawi/kevikepan/new");
    await this.page
      .locator("#kevikepan-keuskupan")
      .selectOption({ label: input.keuskupanName });
    await this.page.locator("#kevikepan-nama").fill(input.name);
    await this.page.locator("#kevikepan-kode").fill(input.code);
    if (input.vikep) {
      await this.page.locator("#kevikepan-vikep").fill(input.vikep);
    }
    await this.page.getByRole("button", { name: "Simpan Kevikepan" }).click();
    await this.page.waitForURL(/\/settings\/gerejawi\/kevikepan/, {
      timeout: 10_000,
    });
  }

  async deleteKevikepan(name: string): Promise<void> {
    await this.goto("/settings/gerejawi/kevikepan");
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.getByRole("button", { name: "Hapus Kevikepan" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Ya, Hapus" })
      .click();
  }

  async expectKevikepanVisible(name: string, code: string): Promise<void> {
    await this.goto("/settings/gerejawi/kevikepan");
    await this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .filter({ hasText: code })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectKevikepanNotVisible(name: string): Promise<void> {
    await this.page.waitForURL(/\/settings\/gerejawi\/kevikepan/, {
      timeout: 10_000,
    });
    await this.page
      .getByText(name, { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async createKota(input: { code: string; name: string }): Promise<void> {
    await this.goto("/settings/sipil/kota/new");
    await this.page.locator("#kota-nama").fill(input.name);
    await this.page.locator("#kota-kode").fill(input.code);
    await this.page.getByRole("button", { name: "Simpan Kota" }).click();
    await this.page.waitForURL(/\/settings\/sipil\/kota$/, { timeout: 10_000 });
  }

  async deleteKota(name: string): Promise<void> {
    await this.goto("/settings/sipil/kota");
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.getByRole("button", { name: "Hapus Kota" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Ya, Hapus" })
      .click();
  }

  async expectKotaNotVisible(name: string): Promise<void> {
    await this.page.waitForURL(/\/settings\/sipil\/kota$/, { timeout: 10_000 });
    await this.page
      .getByText(name, { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async createKecamatan(input: {
    kotaName: string;
    code: string;
    name: string;
  }): Promise<void> {
    await this.goto("/settings/sipil/kecamatan/new");
    await this.page
      .locator("#kecamatan-kota")
      .selectOption({ label: input.kotaName });
    await this.page.locator("#kecamatan-nama").fill(input.name);
    await this.page.locator("#kecamatan-kode").fill(input.code);
    await this.page.getByRole("button", { name: "Simpan Kecamatan" }).click();
    await this.page.waitForURL(/\/settings\/sipil\/kecamatan/, {
      timeout: 10_000,
    });
  }

  async deleteKecamatan(name: string): Promise<void> {
    await this.goto("/settings/sipil/kecamatan");
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.getByRole("button", { name: "Hapus Kecamatan" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Ya, Hapus" })
      .click();
  }

  async expectKecamatanNotVisible(name: string): Promise<void> {
    await this.page.waitForURL(/\/settings\/sipil\/kecamatan/, {
      timeout: 10_000,
    });
    await this.page
      .getByText(name, { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async createKelurahan(input: {
    kecamatanName: string;
    code: string;
    name: string;
    kodePos?: string;
  }): Promise<void> {
    await this.goto("/settings/sipil/kelurahan/new");
    await this.page
      .locator("#kelurahan-kecamatan")
      .selectOption({ label: input.kecamatanName });
    await this.page.locator("#kelurahan-nama").fill(input.name);
    await this.page.locator("#kelurahan-kode").fill(input.code);
    if (input.kodePos) {
      await this.page.locator("#kelurahan-kodepos").fill(input.kodePos);
    }
    await this.page.getByRole("button", { name: "Simpan Kelurahan" }).click();
    await this.page.waitForURL(/\/settings\/sipil\/kelurahan/, {
      timeout: 10_000,
    });
  }

  async deleteKelurahan(name: string): Promise<void> {
    await this.goto("/settings/sipil/kelurahan");
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.getByRole("button", { name: "Hapus Kelurahan" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Ya, Hapus" })
      .click();
  }

  async expectKelurahanNotVisible(name: string): Promise<void> {
    await this.page.waitForURL(/\/settings\/sipil\/kelurahan/, {
      timeout: 10_000,
    });
    await this.page
      .getByText(name, { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async expectErrorAlert(message: string): Promise<void> {
    await this.page
      .getByRole("alert")
      .filter({ hasText: message })
      .waitFor({ state: "visible", timeout: 10_000 });
  }
}
