import { BasePage } from "../base.page";

export class HomePage extends BasePage {
  async expectRendered(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Selamat datang di Domus" })
      .waitFor({ state: "visible", timeout: 10_000 });

    const menus = ["Organisasi", "Kegiatan", "Pengaturan Paroki", "Pengguna"];
    for (const label of menus) {
      await this.page
        .getByRole("link", { name: new RegExp(label) })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async openTile(label: string): Promise<void> {
    await this.page.getByRole("link", { name: new RegExp(label) }).click();
  }
}
