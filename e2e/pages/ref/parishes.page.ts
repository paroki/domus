import { BasePage } from "../base.page";

export class ParishesPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Keuskupan & Paroki" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectForbiddenBanner(): Promise<void> {
    await this.page
      .getByText("Hanya Super Admin yang dapat mengelola keuskupan dan paroki.")
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async createDiocese(input: {
    code: string;
    name: string;
    bishopName?: string;
  }): Promise<void> {
    const section = this.page
      .locator("section")
      .filter({ has: this.page.getByRole("heading", { name: "Keuskupan" }) });

    await section.getByPlaceholder("Kode (mis. kaj)").fill(input.code);
    await section.getByPlaceholder("Nama keuskupan").fill(input.name);
    if (input.bishopName) {
      await section
        .getByPlaceholder("Nama uskup (opsional)")
        .fill(input.bishopName);
    }
    await section.getByRole("button", { name: "Tambah Keuskupan" }).click();
  }

  async expectDioceseVisible(name: string, code: string): Promise<void> {
    const section = this.page
      .locator("section")
      .filter({ has: this.page.getByRole("heading", { name: "Keuskupan" }) });
    await section
      .getByText(`${name} (${code})`)
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteDiocese(name: string): Promise<void> {
    await this.page.getByRole("button", { name: `Hapus ${name}` }).click();
  }

  async expectDioceseNotVisible(name: string): Promise<void> {
    const section = this.page
      .locator("section")
      .filter({ has: this.page.getByRole("heading", { name: "Keuskupan" }) });
    await section
      .getByText(name, { exact: false })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async createParish(input: {
    dioceseName?: string;
    code: string;
    name: string;
  }): Promise<void> {
    const section = this.page.locator("section").filter({
      has: this.page.getByRole("heading", { name: "Paroki", exact: true }),
    });

    if (input.dioceseName) {
      await section
        .locator("select")
        .selectOption({ label: input.dioceseName });
    }
    await section.getByPlaceholder("Kode (mis. kaj)").fill(input.code);
    await section.getByPlaceholder("Nama paroki").fill(input.name);
    const sessionPatchPromise = this.page
      .waitForResponse(
        (r) =>
          r.url().includes("/api/me/session") &&
          r.request().method() === "PATCH",
        { timeout: 5_000 },
      )
      .catch(() => null);
    await section.getByRole("button", { name: "Tambah Paroki" }).click();
    await this.expectParishVisible(input.name, input.code);
    await sessionPatchPromise;
  }

  async expectParishVisible(name: string, code: string): Promise<void> {
    const section = this.page.locator("section").filter({
      has: this.page.getByRole("heading", { name: "Paroki", exact: true }),
    });
    await section
      .getByText(`${name} (${code})`)
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteParish(name: string): Promise<void> {
    await this.page.getByRole("button", { name: `Hapus ${name}` }).click();
  }

  async expectParishNotVisible(name: string): Promise<void> {
    const section = this.page.locator("section").filter({
      has: this.page.getByRole("heading", { name: "Paroki", exact: true }),
    });
    await section
      .getByText(name, { exact: false })
      .waitFor({ state: "detached", timeout: 10_000 });
  }
}
