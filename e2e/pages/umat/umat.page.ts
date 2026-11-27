import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class UmatPage extends BasePage {
  async expectDirectoryLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Data Umat", exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectForbidden(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Akses Data Umat Diperlukan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectErrorAlert(message: string): Promise<void> {
    await this.page
      .getByRole("alert")
      .filter({ hasText: message })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  // --- Kartu Keluarga ---

  async openKeluargaList(): Promise<void> {
    await this.page.goto("/umat/keluarga");
    await this.page
      .getByRole("heading", { name: "Kartu Keluarga", exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async createKeluarga(input: {
    namaKeluarga: string;
    nomorKk?: string;
    alamat?: string;
  }): Promise<void> {
    await this.page
      .getByRole("link", { name: "Tambah Kartu Keluarga" })
      .first()
      .click();
    await this.page
      .getByRole("heading", { name: "Tambah Kartu Keluarga" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page.getByLabel("Nama Keluarga *").fill(input.namaKeluarga);
    if (input.nomorKk) {
      await this.page.getByLabel("Nomor KK").fill(input.nomorKk);
    }
    if (input.alamat) {
      await this.page
        .getByRole("button", { name: "Lanjut ke Alamat & Wilayah Sipil →" })
        .click();
      await this.page.getByLabel("Alamat Lengkap").fill(input.alamat);
    }
    await this.page
      .getByRole("button", { name: "Simpan Kartu Keluarga" })
      .click();
    await this.page
      .getByRole("heading", { name: input.namaKeluarga })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async searchKeluargaByNomorKk(nomorKk: string): Promise<void> {
    await this.page.getByLabel("Nomor KK (Tepat)").fill(nomorKk);
  }

  async expectKeluargaRowMasked(
    namaKeluarga: string,
    masked: string,
    rawNomorKk: string,
  ): Promise<void> {
    const row = this.page
      .locator("div.divide-y > div")
      .filter({ hasText: namaKeluarga });
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row.getByText(masked, { exact: true }).waitFor({ state: "visible" });
    const text = (await row.textContent()) ?? "";
    assert.equal(
      text.includes(rawNomorKk),
      false,
      "Keluarga list row must not expose raw nomorKk",
    );
  }

  async openKeluargaDetail(namaKeluarga: string): Promise<void> {
    await this.page
      .locator("div.divide-y > div")
      .filter({ hasText: namaKeluarga })
      .getByRole("link", { name: "Kelola KK →" })
      .click();
    await this.page
      .getByRole("heading", { name: namaKeluarga })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectKeluargaDetailNomorKk(rawNomorKk: string): Promise<void> {
    await this.page
      .locator("dd", { hasText: rawNomorKk })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteKeluarga(): Promise<void> {
    await this.page.getByRole("button", { name: "Hapus KK" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Hapus KK" })
      .click();
  }

  // --- Umat Wizard & Directory ---

  async createUmatViaWizard(input: {
    nama: string;
    nik?: string;
    jenisKelamin?: "Pria" | "Wanita";
    keluargaName?: string;
    agamaLabel?: string;
    telepon?: string;
    email?: string;
    jurusan?: string;
  }): Promise<void> {
    await this.page.goto("/umat/new");
    await this.page
      .getByRole("heading", { name: "Tambah Data Umat" })
      .waitFor({ state: "visible", timeout: 10_000 });

    await this.page.getByLabel("Nama Lengkap *").fill(input.nama);
    if (input.nik) {
      await this.page.getByLabel("NIK", { exact: true }).fill(input.nik);
    }
    if (input.jenisKelamin) {
      await this.page
        .getByLabel("Jenis Kelamin *")
        .selectOption({ label: input.jenisKelamin });
    }
    if (input.agamaLabel) {
      const agamaSelect = this.page.getByLabel("Agama");
      await agamaSelect
        .locator("option", { hasText: input.agamaLabel })
        .waitFor({ state: "attached", timeout: 10_000 });
      await agamaSelect.selectOption({ label: input.agamaLabel });
    }
    if (input.keluargaName) {
      const select = this.page.getByLabel("Kartu Keluarga");
      await select
        .locator("option", { hasText: input.keluargaName })
        .waitFor({ state: "attached", timeout: 10_000 });
      const val = await select
        .locator("option", { hasText: input.keluargaName })
        .first()
        .getAttribute("value");
      if (val) {
        await select.selectOption(val);
      }
    }

    await this.page
      .getByRole("button", { name: "Lanjut ke Kontak & Domisili →" })
      .click();
    if (input.telepon) {
      await this.page.getByLabel("Telepon").fill(input.telepon);
    }
    if (input.email) {
      await this.page.getByLabel("Email").fill(input.email);
    }

    await this.page
      .getByRole("button", { name: "Lanjut ke Pendidikan & Sosial →" })
      .click();
    if (input.jurusan) {
      await this.page.getByLabel("Jurusan").fill(input.jurusan);
    }

    await this.page.getByRole("button", { name: "Simpan & Buat Umat" }).click();
    await this.page
      .getByRole("heading", { name: input.nama })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async searchUmatByNik(nik: string): Promise<void> {
    await this.page.goto("/umat");
    await this.expectDirectoryLoaded();
    const nikInput = this.page.getByLabel("NIK (Tepat)");
    if (!(await nikInput.isVisible())) {
      await this.page.getByRole("button", { name: /Filter Lanjutan/ }).click();
    }
    await nikInput.fill(nik);
  }

  async expectUmatRowMasked(
    nama: string,
    masked: string,
    rawNik: string,
  ): Promise<void> {
    const row = this.page.locator("div.divide-y > a").filter({ hasText: nama });
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row.getByText(masked, { exact: true }).waitFor({ state: "visible" });
    const text = (await row.textContent()) ?? "";
    assert.equal(
      text.includes(rawNik),
      false,
      "Umat list row must not expose raw NIK",
    );
  }

  async expectUmatRowHasNomorInduk(nama: string): Promise<void> {
    const row = this.page.locator("div.divide-y > a").filter({ hasText: nama });
    await row.waitFor({ state: "visible", timeout: 10_000 });
    const badges = row.locator("span.font-mono");
    const count = await badges.count();
    assert.ok(
      count >= 2,
      `Expected nomorInduk and masked NIK badges for ${nama}, got ${count}`,
    );
    const nomorIndukText = ((await badges.first().textContent()) ?? "").trim();
    assert.ok(
      nomorIndukText.length > 0 && !nomorIndukText.startsWith("*"),
      `Expected non-empty nomorInduk badge, got "${nomorIndukText}"`,
    );
  }

  async openUmatDetail(nama: string): Promise<void> {
    await this.page
      .locator("div.divide-y > a")
      .filter({ hasText: nama })
      .click();
    await this.page
      .getByRole("heading", { name: nama })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async editUmatNamaBaptis(namaBaptis: string): Promise<void> {
    const card = this.page.locator("div.flex.flex-col.gap-5 > div").filter({
      has: this.page.getByRole("heading", {
        name: "Identitas Pribadi & Kependudukan",
      }),
    });
    await card.getByRole("button", { name: "Ubah" }).click();
    await this.page.getByLabel("Nama Baptis").fill(namaBaptis);
    await this.page.getByRole("button", { name: "Simpan Perubahan" }).click();
    await this.page
      .locator("dd", { hasText: namaBaptis })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async linkUserAccount(emailOrName: string): Promise<void> {
    const card = this.page.locator("div.flex.flex-col.gap-5 > div").filter({
      has: this.page.getByRole("heading", {
        name: "Tautan Akun Pengguna Domus",
      }),
    });
    await card.getByRole("button", { name: "Ubah" }).click();
    const select = this.page.getByLabel("Anggota Paroki Terdaftar");
    await select
      .locator("option", { hasText: emailOrName })
      .waitFor({ state: "attached", timeout: 10_000 });
    const val = await select
      .locator("option", { hasText: emailOrName })
      .first()
      .getAttribute("value");
    if (val) {
      await select.selectOption(val);
    }
    await this.page.getByRole("button", { name: "Simpan Perubahan" }).click();
    await this.page
      .locator("dd", { hasText: emailOrName })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  // --- Sakramen ---

  async addOrUpdateSakramen(input: {
    jenisLabel: string;
    tanggal?: string;
    buku?: string;
    nomor?: string;
  }): Promise<void> {
    await this.page.getByRole("link", { name: "Catat Sakramen" }).click();
    await this.page
      .getByRole("heading", { name: "Catat Sakramen" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByLabel("Jenis Sakramen *")
      .selectOption({ label: input.jenisLabel });
    if (input.tanggal) {
      await this.page.getByLabel("Tanggal Sakramen").fill(input.tanggal);
    }
    if (input.buku) {
      await this.page.getByLabel("Buku Register").fill(input.buku);
    }
    if (input.nomor) {
      await this.page.getByLabel("Nomor Register").fill(input.nomor);
    }
    await this.page.getByRole("button", { name: "Simpan Sakramen" }).click();
    await this.page
      .getByRole("heading", { name: /Riwayat Sakramen/ })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async editSakramenNomor(jenisLabel: string, nomor: string): Promise<void> {
    const row = this.page
      .locator("div.divide-y > div")
      .filter({ hasText: jenisLabel });
    await row.getByRole("link", { name: "Ubah" }).click();
    await this.page
      .getByRole("heading", { name: "Ubah Catatan Sakramen" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page.getByLabel("Nomor Register").fill(nomor);
    await this.page.getByRole("button", { name: "Simpan Sakramen" }).click();
    await this.page
      .getByText(`No. ${nomor}`)
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteSakramen(jenisLabel: string): Promise<void> {
    const row = this.page
      .locator("div.divide-y > div")
      .filter({ hasText: jenisLabel });
    await row.getByRole("button", { name: "Hapus" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Hapus", exact: true })
      .click();
    await row.waitFor({ state: "detached", timeout: 10_000 });
  }

  async deleteUmat(): Promise<void> {
    await this.page.getByRole("button", { name: "Hapus Umat" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Hapus Umat" })
      .click();
    await this.expectDirectoryLoaded();
  }

  // --- Kamus Referensi Umat ---

  async createReferensi(input: {
    category: string;
    code: string;
    label: string;
    sortOrder?: number;
  }): Promise<void> {
    await this.goto(
      `/umat/referensi/new?category=${encodeURIComponent(input.category)}`,
    );
    await this.page.locator("#referensi-category").selectOption(input.category);
    await this.page.locator("#referensi-label").fill(input.label);
    await this.page.locator("#referensi-code").fill(input.code);
    if (input.sortOrder !== undefined) {
      await this.page.locator("#referensi-sort").fill(String(input.sortOrder));
    }
    await this.page.getByRole("button", { name: "Simpan Item" }).click();
    await this.page.waitForURL(/\/umat\/referensi\?category=/, {
      timeout: 10_000,
    });
  }

  async expectReferensiVisible(
    label: string,
    scopeLabel: string,
  ): Promise<void> {
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ has: this.page.getByText(label, { exact: true }) })
      .first();
    await card.waitFor({ state: "visible", timeout: 10_000 });
    const scopeText = await card.getByText(scopeLabel, { exact: true }).count();
    assert.ok(scopeText > 0, `Expected scope badge "${scopeLabel}"`);
  }

  async expectReferensiNotVisible(label: string): Promise<void> {
    await this.page
      .getByText(label, { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async expectPickerHasOption(
    path: string,
    fieldLabel: string,
    optionLabel: string,
  ): Promise<void> {
    await this.goto(path);
    const select = this.page.getByLabel(fieldLabel);
    await select
      .locator("option", { hasText: optionLabel })
      .waitFor({ state: "attached", timeout: 10_000 });
  }

  async editReferensi(
    currentLabel: string,
    input: { label: string; sortOrder?: number },
  ): Promise<void> {
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ has: this.page.getByText(currentLabel, { exact: true }) })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.locator("#referensi-label").fill(input.label);
    if (input.sortOrder !== undefined) {
      await this.page.locator("#referensi-sort").fill(String(input.sortOrder));
    }
    await this.page.getByRole("button", { name: "Simpan Item" }).click();
    await this.page.waitForURL(/\/umat\/referensi\?category=/, {
      timeout: 10_000,
    });
  }

  async toggleReferensiActive(label: string): Promise<void> {
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ has: this.page.getByText(label, { exact: true }) })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.getByLabel("Item Aktif").click();
    await this.page.getByRole("button", { name: "Simpan Item" }).click();
    await this.page.waitForURL(/\/umat\/referensi\?category=/, {
      timeout: 10_000,
    });
  }

  async tryDeleteReferensi(label: string): Promise<void> {
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
  }

  async deleteReferensi(label: string): Promise<void> {
    await this.tryDeleteReferensi(label);
    await this.page.waitForURL(/\/umat\/referensi\?category=/, {
      timeout: 10_000,
    });
  }

  async expectReferensiListReadOnly(): Promise<void> {
    await this.page
      .getByRole("heading", { name: /Kamus:/ })
      .waitFor({ state: "visible", timeout: 10_000 });
    assert.equal(
      await this.page.getByRole("link", { name: "Tambah Item" }).count(),
      0,
      "Expected plain user not to see Tambah Item link",
    );
    assert.equal(
      await this.page.getByRole("link", { name: "Ubah" }).count(),
      0,
      "Expected plain user not to see Ubah links",
    );
  }
}
