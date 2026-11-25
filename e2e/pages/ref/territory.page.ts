import { BasePage } from "../base.page";

export class TerritoryPage extends BasePage {
  private zonesSection() {
    return this.page.locator("section").filter({
      has: this.page.getByRole("heading", { name: "Wilayah / Stasi" }),
    });
  }

  private wardsSection() {
    return this.page.locator("section").filter({
      has: this.page.getByRole("heading", { name: "Lingkungan" }),
    });
  }

  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Wilayah & Lingkungan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async createZone(input: {
    type?: "wilayah" | "stasi";
    code: string;
    name: string;
    coordinatorName?: string;
    coordinatorPhone?: string;
  }): Promise<void> {
    const section = this.zonesSection();
    if (input.type) {
      await section.locator("form select").selectOption(input.type);
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

  async editZone(
    currentName: string,
    input: { code?: string; name?: string; coordinatorName?: string },
  ): Promise<void> {
    const section = this.zonesSection();
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

  async toggleZoneActive(name: string): Promise<void> {
    const row = this.zonesSection()
      .locator("div.divide-y > div")
      .filter({ hasText: name });
    await row.locator("button[title]").first().click();
  }

  async deleteZone(name: string): Promise<void> {
    await this.zonesSection()
      .getByRole("button", { name: `Hapus ${name}` })
      .click();
  }

  async expectZoneVisible(name: string, code: string): Promise<void> {
    const section = this.zonesSection();
    await section
      .locator("div.divide-y > div")
      .filter({ hasText: name })
      .filter({ hasText: code })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectZoneNotVisible(name: string): Promise<void> {
    await this.zonesSection()
      .locator("div.divide-y > div")
      .filter({ hasText: name })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async createWard(input: {
    zoneName?: string;
    code: string;
    name: string;
    patronSaint?: string;
    headName?: string;
    headPhone?: string;
  }): Promise<void> {
    const section = this.wardsSection();
    if (input.zoneName) {
      const select = section.locator("form select");
      await select
        .locator("option", { hasText: input.zoneName })
        .waitFor({ state: "attached", timeout: 10_000 });
      await select.selectOption({ label: input.zoneName });
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

  async editWard(
    currentName: string,
    input: { code?: string; name?: string; headName?: string },
  ): Promise<void> {
    const section = this.wardsSection();
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

  async toggleWardActive(name: string): Promise<void> {
    const row = this.wardsSection()
      .locator("div.divide-y > div")
      .filter({ hasText: name });
    await row.locator("button[title]").first().click();
  }

  async deleteWard(name: string): Promise<void> {
    await this.wardsSection()
      .getByRole("button", { name: `Hapus ${name}` })
      .click();
  }

  async expectWardVisible(name: string, code: string): Promise<void> {
    await this.wardsSection()
      .locator("div.divide-y > div")
      .filter({ hasText: name })
      .filter({ hasText: code })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectWardNotVisible(name: string): Promise<void> {
    await this.wardsSection()
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
