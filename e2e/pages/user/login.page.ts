import { BasePage } from "../base.page";

export class LoginPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Masuk ke Domus" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByRole("button", { name: "Kirim Kode OTP" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }
}
