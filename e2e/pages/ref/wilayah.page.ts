import { BasePage } from "../base.page";

export class WilayahPage extends BasePage {
  private wilayahSection() {
    return this.page.locator("section").filter({
      has: this.page.getByRole("heading", { name: "Wilayah / Stasi" }),
    });
  }

  private lingkunganSection() {
    return this.page.locator("section").filter({
      has: this.page.getByRole("heading", { name: "Lingkungan" }),
    });
  }

  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Wilayah & Lingkungan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async createWilayah(input: {
    jenis?: "wilayah" | "stasi";
    code: string;
    name: string;
    coordinatorName?: string;
    coordinatorPhone?: string;
  }): Promise<void> {
    const section = this.wilayahSection();
    if (input.jenis) {
      await section.locator("form select").selectOption(input.jenis);
    }
    await section.getByPlaceholder("Kode (mis. wil-01)").fill(input.code);
    await section.getByPlaceholder("Nama wilayah / stasi").fill(input.name);
    if (input.coordinatorName) {
      await section
        .getByPlaceholder("Koordinator (opsional)")
        .fill(input.coordinatorName);
    }
    if (input.coordinatorPhone) {
      await section.getByPlaceholder("Telepon").fill(input.coordinatorPhone);
    }
    await section.getByRole("button", { name: "Tambah Wilayah" }).click();
  }

  async editWilayah(
    currentName: string,
    input: { code?: string; name?: string; coordinatorName?: string },
  ): Promise<void> {
    const section = this.wilayahSection();
    await section.getByRole("button", { name: `Ubah ${currentName}` }).click();
    if (input.code !== undefined) {
      await section.getByPlaceholder("Kode (mis. wil-01)").fill(input.code);
    }
    if (input.name !== undefined) {
      await section.getByPlaceholder("Nama wilayah / stasi").fill(input.name);
    }
    if (input.coordinatorName !== undefined) {
      await section
        .getByPlaceholder("Koordinator (opsional)")
        .fill(input.coordinatorName);
    }
    await section.getByRole("button", { name: "Simpan Perubahan" }).click();
  }

  async toggleWilayahActive(name: string): Promise<void> {
    const row = this.wilayahSection()
      .locator("div.divide-y > div")
      .filter({ hasText: name });
    await row.locator("button[title]").first().click();
  }

  async deleteWilayah(name: string): Promise<void> {
    await this.wilayahSection()
      .getByRole("button", { name: `Hapus ${name}` })
      .click();
  }

  async expectWilayahVisible(name: string, code: string): Promise<void> {
    const section = this.wilayahSection();
    await section
      .locator("div.divide-y > div")
      .filter({ hasText: name })
      .filter({ hasText: code })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectWilayahNotVisible(name: string): Promise<void> {
    await this.wilayahSection()
      .locator("div.divide-y > div")
      .filter({ hasText: name })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async createLingkungan(input: {
    wilayahName?: string;
    code: string;
    name: string;
    patronSaint?: string;
    headName?: string;
    headPhone?: string;
  }): Promise<void> {
    const section = this.lingkunganSection();
    if (input.wilayahName) {
      const select = section.locator("form select");
      await select
        .locator("option", { hasText: input.wilayahName })
        .waitFor({ state: "attached", timeout: 10_000 });
      await select.selectOption({ label: input.wilayahName });
    }
    await section.getByPlaceholder("Kode (mis. ling-01)").fill(input.code);
    await section.getByPlaceholder("Nama lingkungan").fill(input.name);
    if (input.patronSaint) {
      await section
        .getByPlaceholder("Santo/Santa Pelindung")
        .fill(input.patronSaint);
    }
    if (input.headName) {
      await section
        .getByPlaceholder("Ketua lingkungan (opsional)")
        .fill(input.headName);
    }
    if (input.headPhone) {
      await section.getByPlaceholder("Telepon").fill(input.headPhone);
    }
    await section.getByRole("button", { name: "Tambah Lingkungan" }).click();
  }

  async editLingkungan(
    currentName: string,
    input: { code?: string; name?: string; headName?: string },
  ): Promise<void> {
    const section = this.lingkunganSection();
    await section.getByRole("button", { name: `Ubah ${currentName}` }).click();
    if (input.code !== undefined) {
      await section.getByPlaceholder("Kode (mis. ling-01)").fill(input.code);
    }
    if (input.name !== undefined) {
      await section.getByPlaceholder("Nama lingkungan").fill(input.name);
    }
    if (input.headName !== undefined) {
      await section
        .getByPlaceholder("Ketua lingkungan (opsional)")
        .fill(input.headName);
    }
    await section.getByRole("button", { name: "Simpan Perubahan" }).click();
  }

  async toggleLingkunganActive(name: string): Promise<void> {
    const row = this.lingkunganSection()
      .locator("div.divide-y > div")
      .filter({ hasText: name });
    await row.locator("button[title]").first().click();
  }

  async deleteLingkungan(name: string): Promise<void> {
    await this.lingkunganSection()
      .getByRole("button", { name: `Hapus ${name}` })
      .click();
  }

  async expectLingkunganVisible(name: string, code: string): Promise<void> {
    await this.lingkunganSection()
      .locator("div.divide-y > div")
      .filter({ hasText: name })
      .filter({ hasText: code })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectLingkunganNotVisible(name: string): Promise<void> {
    await this.lingkunganSection()
      .locator("div.divide-y > div")
      .filter({ hasText: name })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async expectErrorAlert(message: string): Promise<void> {
    await this.page
      .getByRole("alert")
      .filter({ hasText: message })
      .waitFor({ state: "visible", timeout: 10_000 });
  }
}
