import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class OrgJoinPage extends BasePage {
  async expectPreviewLoaded(
    orgName: string,
    options?: {
      roleBadge?: string;
      defaultPosition?: string;
      statusBadge?: string;
    },
  ): Promise<void> {
    await this.page
      .getByRole("heading", { name: orgName })
      .waitFor({ state: "visible", timeout: 10_000 });
    if (options?.roleBadge) {
      await this.page
        .getByText(options.roleBadge, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.defaultPosition) {
      await this.page
        .getByText(options.defaultPosition, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.statusBadge) {
      await this.page
        .getByText(options.statusBadge, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectInvalidOrRevoked(textSubstring: string): Promise<void> {
    const body = this.page.locator("main");
    await body.waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByText(textSubstring, { exact: false })
      .waitFor({ state: "visible", timeout: 10_000 });
    assert.equal(
      await this.page.getByRole("button", { name: /^Gabung ke / }).count(),
      0,
    );
  }

  async requestGuestOtp(input: { name: string; email: string }): Promise<void> {
    await this.page
      .getByLabel("Nama lengkap (untuk pengguna baru)")
      .fill(input.name);
    await this.page.getByLabel("Alamat email").fill(input.email);
    await this.page.getByRole("button", { name: "Kirim Kode OTP" }).click();
    await this.page
      .getByLabel("Kode OTP")
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async verifyGuestOtp(orgName: string, code: string): Promise<void> {
    await this.page.getByLabel("Kode OTP").fill(code);
    await this.page
      .getByRole("button", { name: `Gabung ke ${orgName}` })
      .click();
  }

  async redeemAsAuthenticatedUser(orgName: string): Promise<void> {
    await this.page
      .getByRole("button", { name: `Gabung ke ${orgName}` })
      .click();
  }
}
