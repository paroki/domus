import { BasePage } from "../base.page";

export class WilayahPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Wilayah & Stasi" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async createWilayah(input: {
    jenis?: "wilayah" | "stasi";
    code: string;
    name: string;
    coordinatorName?: string;
    coordinatorPhone?: string;
  }): Promise<void> {
    await this.goto("/settings/paroki/wilayah/new");
    if (input.jenis) {
      await this.page.locator("#wilayah-jenis").selectOption(input.jenis);
    }
    await this.page.locator("#wilayah-kode").fill(input.code);
    await this.page.locator("#wilayah-nama").fill(input.name);
    if (input.coordinatorName) {
      await this.page
        .locator("#wilayah-koordinator")
        .fill(input.coordinatorName);
    }
    if (input.coordinatorPhone) {
      await this.page.locator("#wilayah-telp").fill(input.coordinatorPhone);
    }
    await this.page.getByRole("button", { name: "Simpan Wilayah" }).click();
    await this.page.waitForURL(/\/settings\/paroki\/wilayah$/, {
      timeout: 10_000,
    });
  }

  async deleteWilayah(name: string): Promise<void> {
    await this.goto("/settings/paroki/wilayah");
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.getByRole("button", { name: "Hapus Wilayah" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Ya, Hapus" })
      .click();
  }

  async expectWilayahVisible(name: string, code: string): Promise<void> {
    await this.goto("/settings/paroki/wilayah");
    await this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .filter({ hasText: code })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectWilayahNotVisible(name: string): Promise<void> {
    await this.page.waitForURL(/\/settings\/paroki\/wilayah$/, {
      timeout: 10_000,
    });
    await this.page
      .getByText(name, { exact: true })
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
    await this.goto("/settings/paroki/lingkungan/new");
    if (input.wilayahName) {
      const select = this.page.locator("#lingkungan-wilayah");
      await select
        .locator("option", { hasText: input.wilayahName })
        .waitFor({ state: "attached", timeout: 10_000 });
      await select.selectOption({ label: input.wilayahName });
    }
    await this.page.locator("#lingkungan-kode").fill(input.code);
    await this.page.locator("#lingkungan-nama").fill(input.name);
    if (input.patronSaint) {
      await this.page.locator("#lingkungan-pelindung").fill(input.patronSaint);
    }
    if (input.headName) {
      await this.page.locator("#lingkungan-ketua").fill(input.headName);
    }
    if (input.headPhone) {
      await this.page.locator("#lingkungan-telp").fill(input.headPhone);
    }
    await this.page.getByRole("button", { name: "Simpan Lingkungan" }).click();
    await this.page.waitForURL(/\/settings\/paroki\/lingkungan/, {
      timeout: 10_000,
    });
  }

  async deleteLingkungan(name: string): Promise<void> {
    await this.goto("/settings/paroki/lingkungan");
    const card = this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .first();
    await card.getByRole("link", { name: "Ubah" }).click();
    await this.page.getByRole("button", { name: "Hapus Lingkungan" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Ya, Hapus" })
      .click();
    await this.page.waitForURL(/\/settings\/paroki\/lingkungan$/, {
      timeout: 10_000,
    });
  }

  async expectLingkunganVisible(name: string, code: string): Promise<void> {
    await this.goto("/settings/paroki/lingkungan");
    await this.page
      .locator("div.rounded-xl")
      .filter({ hasText: name })
      .filter({ hasText: code })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectLingkunganNotVisible(name: string): Promise<void> {
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
