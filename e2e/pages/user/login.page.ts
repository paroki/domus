import assert from "node:assert/strict";
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

  async requestEmailOtp(email: string): Promise<void> {
    await this.page.getByLabel("Alamat email").fill(email);
    await this.page.getByRole("button", { name: "Kirim Kode OTP" }).click();
  }

  async submitOtpCode(code: string): Promise<void> {
    await this.page.getByLabel("Kode OTP").fill(code);
    await this.page.getByRole("button", { name: "Masuk" }).click();
  }

  async expectErrorAlert(textSubstring: string): Promise<void> {
    const alert = this.page.getByRole("alert");
    await alert.waitFor({ state: "visible", timeout: 10_000 });
    const text = (await alert.textContent()) ?? "";
    assert.ok(
      text.includes(textSubstring),
      `Expected alert "${text}" to include "${textSubstring}"`,
    );
    assert.equal(await this.page.getByLabel("Kode OTP").count(), 0);
  }
}
