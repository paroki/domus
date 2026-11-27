import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class CantusPage extends BasePage {
  async expectDirectoryLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Katalog Nyanyian & Mazmur" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectCreateButtonVisible(visible: boolean): Promise<void> {
    const link = this.page.getByRole("link", { name: "Tambah Nyanyian" });
    if (visible) {
      await link.first().waitFor({ state: "visible", timeout: 10_000 });
    } else {
      assert.equal(
        await link.count(),
        0,
        "Expected Tambah Nyanyian CTA to be hidden",
      );
    }
  }

  async expectForbidden(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Akses Ditolak" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByText(/Admin Nyanyian \(cantus:admin\)/)
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async openCreateFromDirectory(): Promise<void> {
    await this.page
      .getByRole("link", { name: "Tambah Nyanyian" })
      .first()
      .click();
    await this.page
      .getByRole("heading", { name: "Tambah Nyanyian", exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async createLaguViaWizard(input: {
    judul: string;
    sumber?: string;
    nomor?: string;
    kategori?: string;
    nadaDasar?: string;
    tags?: string[];
    lirik?: string;
    birama?: string;
    tempo?: string;
    kreditLagu?: string;
    media?: Array<{
      type: "IMAGE" | "PDF" | "AUDIO" | "VIDEO";
      title: string;
      viewUrl: string;
    }>;
  }): Promise<string> {
    await this.openCreateFromDirectory();
    await this.page.getByLabel("Judul *").fill(input.judul);
    if (input.sumber) {
      await this.page.getByLabel("Sumber").fill(input.sumber);
    }
    if (input.nomor) {
      await this.page.getByLabel("Nomor Lagu / Mazmur").fill(input.nomor);
    }
    if (input.kategori) {
      await this.page.getByLabel("Kategori").fill(input.kategori);
    }
    if (input.nadaDasar) {
      await this.page.getByLabel("Nada Dasar").fill(input.nadaDasar);
    }
    if (input.tags?.length) {
      const tagInput = this.page.getByPlaceholder(
        "Ketik tag lalu tekan Enter atau Tambah Tag",
      );
      for (const tg of input.tags) {
        await tagInput.fill(tg);
        await this.page.getByRole("button", { name: "Tambah Tag" }).click();
      }
    }

    await this.page
      .getByRole("button", { name: "Lanjut ke Lirik / Mazmur" })
      .click();

    if (input.lirik) {
      await this.page.getByLabel("Lirik Lagu").fill(input.lirik);
    }
    if (input.birama) {
      await this.page.getByLabel("Birama").fill(input.birama);
    }
    if (input.tempo) {
      await this.page.getByLabel("Tempo").fill(input.tempo);
    }

    await this.page
      .getByRole("button", { name: "Lanjut ke Kredit & Media" })
      .click();

    if (input.kreditLagu) {
      await this.page.getByLabel("Pencipta Lagu").fill(input.kreditLagu);
    }
    if (input.media?.length) {
      for (let i = 0; i < input.media.length; i++) {
        const m = input.media[i];
        if (!m) continue;
        await this.page.getByRole("button", { name: "Tambah Media" }).click();
        await this.page.getByLabel("Tipe").nth(i).selectOption(m.type);
        await this.page.getByLabel("Judul Media").nth(i).fill(m.title);
        await this.page
          .getByLabel("URL Tampilan (http/https)")
          .nth(i)
          .fill(m.viewUrl);
      }
    }

    await this.page
      .getByRole("button", { name: "Simpan & Buat Nyanyian" })
      .click();
    await this.expectDetailLoaded(input.judul);
    return this.page.url().split("/").pop() ?? "";
  }

  async createMazmurViaWizard(input: {
    judul: string;
    sumber?: string;
    nomor?: string;
    tags?: string[];
    perayaan?: string;
    siklusTahun?: string;
    masa?: string;
    refAlkitab?: string;
    ulangan?: string;
    bait?: string;
    bpi?: string;
  }): Promise<string> {
    await this.page.goto("/cantus/new");
    await this.page
      .getByRole("heading", { name: "Tambah Nyanyian", exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });

    await this.page.getByLabel("Jenis Nyanyian *").selectOption("mazmur");
    await this.page.getByLabel("Judul *").fill(input.judul);
    if (input.sumber) {
      await this.page.getByLabel("Sumber").fill(input.sumber);
    }
    if (input.nomor) {
      await this.page.getByLabel("Nomor Lagu / Mazmur").fill(input.nomor);
    }
    if (input.tags?.length) {
      const tagInput = this.page.getByPlaceholder(
        "Ketik tag lalu tekan Enter atau Tambah Tag",
      );
      for (const tg of input.tags) {
        await tagInput.fill(tg);
        await this.page.getByRole("button", { name: "Tambah Tag" }).click();
      }
    }

    await this.page
      .getByRole("button", { name: "Lanjut ke Lirik / Mazmur" })
      .click();

    if (input.perayaan) {
      await this.page.getByLabel("Perayaan / Hari Minggu").fill(input.perayaan);
    }
    if (input.siklusTahun) {
      await this.page
        .getByLabel("Siklus Tahun")
        .selectOption(input.siklusTahun);
    }
    if (input.masa) {
      await this.page.getByLabel("Masa Liturgi").fill(input.masa);
    }
    if (input.refAlkitab) {
      await this.page.getByLabel("Referensi Alkitab").fill(input.refAlkitab);
    }
    if (input.ulangan) {
      await this.page.getByLabel("Ulangan (Refrain)").fill(input.ulangan);
    }
    if (input.bait) {
      await this.page.getByLabel("Bait Mazmur").fill(input.bait);
    }
    if (input.bpi) {
      await this.page.getByLabel("Bait Pengantar Injil (BPI)").fill(input.bpi);
    }

    await this.page
      .getByRole("button", { name: "Simpan & Buat Nyanyian" })
      .click();
    await this.expectDetailLoaded(input.judul);
    return this.page.url().split("/").pop() ?? "";
  }

  async expectDetailLoaded(judul: string): Promise<void> {
    await this.page
      .getByRole("heading", { name: judul })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectInlineMedia(options: {
    imageAlt?: string;
    audioTitle?: string;
    pdfTitle?: string;
  }): Promise<void> {
    if (options.imageAlt) {
      await this.page
        .locator(`img[alt="${options.imageAlt}"]`)
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options.audioTitle) {
      await this.page
        .getByText(options.audioTitle, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
      await this.page
        .locator("audio[controls]")
        .first()
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options.pdfTitle) {
      await this.page
        .getByText(options.pdfTitle, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectMazmurSections(input: {
    ulangan: string;
    bait: string;
    bpi: string;
  }): Promise<void> {
    await this.page
      .getByText(input.ulangan)
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByText(input.bait)
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByText(input.bpi)
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async editIdentityNadaDasar(nadaDasar: string): Promise<void> {
    await this.page.getByRole("button", { name: "Ubah Identitas" }).click();
    await this.page.getByLabel("Nada Dasar").fill(nadaDasar);
    await this.page
      .getByRole("button", { name: "Simpan", exact: true })
      .click();
    await this.page
      .getByText("Identitas nyanyian diperbarui.")
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async editLirik(lirik: string): Promise<void> {
    await this.page
      .getByRole("button", { name: "Ubah Lirik / Mazmur" })
      .click();
    await this.page.getByLabel("Lirik Lagu").fill(lirik);
    await this.page
      .getByRole("button", { name: "Simpan", exact: true })
      .click();
    await this.page
      .getByText("Lirik dan metadata musik diperbarui.")
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectReadOnlyDetail(): Promise<void> {
    for (const label of [
      "Ubah Identitas",
      "Ubah Lirik / Mazmur",
      "Ubah Kredit",
      "Ubah Media",
      "Hapus Nyanyian",
    ]) {
      assert.equal(
        await this.page.getByRole("button", { name: label }).count(),
        0,
        `Expected "${label}" button to be hidden in read-only mode`,
      );
    }
  }

  async searchDirectory(query: string): Promise<void> {
    await this.page
      .getByRole("searchbox", {
        name: "Cari judul, nomor lagu, lirik, atau pengarang",
      })
      .fill(query);
  }

  async filterJenis(label: string): Promise<void> {
    await this.page.getByRole("button", { name: label, exact: true }).click();
  }

  async filterSumber(sumber: string): Promise<void> {
    await this.page.getByLabel("Filter sumber").selectOption(sumber);
  }

  async filterTag(tag: string): Promise<void> {
    await this.page.getByLabel("Filter tag").fill(tag);
  }

  async expectSongVisible(judul: string): Promise<void> {
    await this.page
      .locator("div.divide-y > a")
      .filter({ hasText: judul })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectSongNotVisible(judul: string): Promise<void> {
    await this.page
      .locator("div.divide-y > a")
      .filter({ hasText: judul })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async openSongDetail(judul: string): Promise<void> {
    await this.page
      .locator("div.divide-y > a")
      .filter({ hasText: judul })
      .first()
      .click();
    await this.expectDetailLoaded(judul);
  }

  async deactivateSong(): Promise<void> {
    await this.page
      .getByRole("button", { name: "Nonaktifkan Nyanyian" })
      .click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Nonaktifkan" })
      .click();
    await this.page
      .getByText("Nyanyian dinonaktifkan.")
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteSong(): Promise<void> {
    await this.page.getByRole("button", { name: "Hapus Nyanyian" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Hapus Nyanyian" })
      .click();
    await this.expectDirectoryLoaded();
  }
}
