import { BasePage } from "../base.page";

export class ParokiPage extends BasePage {
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

  async createKeuskupan(input: {
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

  async expectKeuskupanVisible(name: string, code: string): Promise<void> {
    const section = this.page
      .locator("section")
      .filter({ has: this.page.getByRole("heading", { name: "Keuskupan" }) });
    await section
      .getByText(`${name} (${code})`)
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteKeuskupan(name: string): Promise<void> {
    await this.page.getByRole("button", { name: `Hapus ${name}` }).click();
  }

  async expectKeuskupanNotVisible(name: string): Promise<void> {
    const section = this.page
      .locator("section")
      .filter({ has: this.page.getByRole("heading", { name: "Keuskupan" }) });
    await section
      .getByText(name, { exact: false })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async createParoki(input: {
    keuskupanName?: string;
    code: string;
    name: string;
  }): Promise<void> {
    const section = this.page.locator("section").filter({
      has: this.page.getByRole("heading", { name: "Paroki", exact: true }),
    });

    if (input.keuskupanName) {
      await section
        .locator("select")
        .selectOption({ label: input.keuskupanName });
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
    await this.expectParokiVisible(input.name, input.code);
    await sessionPatchPromise;
  }

  async expectParokiVisible(name: string, code: string): Promise<void> {
    const section = this.page.locator("section").filter({
      has: this.page.getByRole("heading", { name: "Paroki", exact: true }),
    });
    await section
      .getByText(`${name} (${code})`)
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteParoki(name: string): Promise<void> {
    await this.page.getByRole("button", { name: `Hapus ${name}` }).click();
  }

  async expectParokiNotVisible(name: string): Promise<void> {
    const section = this.page.locator("section").filter({
      has: this.page.getByRole("heading", { name: "Paroki", exact: true }),
    });
    await section
      .getByText(name, { exact: false })
      .waitFor({ state: "detached", timeout: 10_000 });
  }
}
