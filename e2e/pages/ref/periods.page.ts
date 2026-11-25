import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class PeriodsPage extends BasePage {
  private rowFor(name: string) {
    return this.page
      .locator("section > div.divide-y > div")
      .filter({ hasText: name });
  }

  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Periode Fiskal & Liturgi" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async createPeriod(input: {
    type?: "fiscal" | "liturgical";
    code: string;
    name: string;
    startDate: string;
    endDate: string;
    isCurrent?: boolean;
  }): Promise<void> {
    const form = this.page.locator("form");
    if (input.type) {
      await form.locator("select").selectOption(input.type);
    }
    await form.getByPlaceholder("Kode (mis. kaj)").fill(input.code);
    await form.getByPlaceholder("Nama periode").fill(input.name);
    await form.getByLabel("Tanggal mulai").fill(input.startDate);
    await form.getByLabel("Tanggal selesai").fill(input.endDate);
    const checkbox = form.getByLabel("Jadikan sebagai periode aktif");
    if (input.isCurrent) {
      await checkbox.check();
    } else {
      await checkbox.uncheck();
    }
    await form.getByRole("button", { name: "Tambah Periode" }).click();
  }

  async editPeriod(
    currentName: string,
    input: {
      code?: string;
      name?: string;
      startDate?: string;
      endDate?: string;
      isCurrent?: boolean;
    },
  ): Promise<void> {
    await this.page
      .getByRole("button", { name: `Ubah ${currentName}` })
      .click();
    const form = this.page.locator("form");
    if (input.code !== undefined) {
      await form.getByPlaceholder("Kode (mis. kaj)").fill(input.code);
    }
    if (input.name !== undefined) {
      await form.getByPlaceholder("Nama periode").fill(input.name);
    }
    if (input.startDate !== undefined) {
      await form.getByLabel("Tanggal mulai").fill(input.startDate);
    }
    if (input.endDate !== undefined) {
      await form.getByLabel("Tanggal selesai").fill(input.endDate);
    }
    if (input.isCurrent !== undefined) {
      const checkbox = form.getByLabel("Jadikan sebagai periode aktif");
      if (input.isCurrent) await checkbox.check();
      else await checkbox.uncheck();
    }
    await form.getByRole("button", { name: "Simpan Perubahan" }).click();
  }

  async makeCurrent(name: string): Promise<void> {
    const row = this.rowFor(name);
    await row.getByRole("button", { name: "Jadikan Periode Aktif" }).click();
  }

  async closePeriod(name: string): Promise<void> {
    const row = this.rowFor(name);
    await row.getByRole("button", { name: "Tutup" }).click();
  }

  async reopenPeriod(name: string): Promise<void> {
    const row = this.rowFor(name);
    await row.getByRole("button", { name: "Buka Kembali" }).click();
  }

  async deletePeriod(name: string): Promise<void> {
    await this.page.getByRole("button", { name: `Hapus ${name}` }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Ya, Hapus" })
      .click();
  }

  async expectPeriodVisible(
    name: string,
    options?: { status?: "Terbuka" | "Ditutup"; isCurrent?: boolean },
  ): Promise<void> {
    const row = this.rowFor(name);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    if (options?.status) {
      await row
        .getByText(options.status, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.isCurrent === true) {
      await row
        .getByText("Periode Aktif", { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.isCurrent === false) {
      const count = await row
        .getByText("Periode Aktif", { exact: true })
        .count();
      assert.equal(
        count,
        0,
        `Expected ${name} not to have Periode Aktif badge`,
      );
    }
  }

  async expectPeriodNotVisible(name: string): Promise<void> {
    await this.rowFor(name).waitFor({ state: "detached", timeout: 10_000 });
  }

  async expectErrorAlert(message: string): Promise<void> {
    await this.page
      .getByRole("alert")
      .filter({ hasText: message })
      .waitFor({ state: "visible", timeout: 10_000 });
  }
}
