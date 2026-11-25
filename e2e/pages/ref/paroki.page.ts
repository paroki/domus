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
  }): Promise<void> {
    await this.goto("/settings/gerejawi/keuskupan/new");
    await this.page.locator("#keuskupan-kode").fill(input.code);
    await this.page.locator("#keuskupan-nama").fill(input.name);
    if (input.bishopName) {
      await this.page.locator("#keuskupan-uskup").fill(input.bishopName);
    }
    await this.page.getByRole("button", { name: "Simpan Keuskupan" }).click();
    await this.page.waitForURL(/\/settings\/gerejawi\/keuskupan$/, {
      timeout: 10_000,
    });
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
    code: string;
    name: string;
  }): Promise<void> {
    await this.goto("/settings/gerejawi/paroki/new");
    if (input.keuskupanName) {
      await this.page
        .locator("#paroki-keuskupan")
        .selectOption({ label: input.keuskupanName });
    }
    await this.page.locator("#paroki-kode").fill(input.code);
    await this.page.locator("#paroki-nama").fill(input.name);
    await this.page.getByRole("button", { name: "Simpan Paroki" }).click();
    await this.page.waitForURL(/\/settings\/gerejawi\/paroki\/[^/]+$/, {
      timeout: 10_000,
    });
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
      .getByText(name, { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
  }
}
