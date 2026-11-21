import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class MePage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Profil & Keamanan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async updateProfile(input: { name: string; phone?: string }): Promise<void> {
    await this.page.getByLabel("Nama lengkap").fill(input.name);
    if (input.phone !== undefined) {
      await this.page.getByLabel("Nomor telepon (WhatsApp)").fill(input.phone);
    }
    await this.page.getByRole("button", { name: "Simpan Profil" }).click();
  }

  async expectProfileSaved(): Promise<void> {
    await this.page
      .getByText("Profil berhasil diperbarui.", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  private sessionsCard() {
    return this.page
      .locator("div")
      .filter({
        has: this.page.getByRole("heading", { name: "Sesi Aktif" }),
      })
      .last();
  }

  async expectCurrentSessionVisible(): Promise<void> {
    await this.sessionsCard()
      .getByText("Perangkat ini", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async revokeOtherSession(): Promise<void> {
    const card = this.sessionsCard();
    const beforeCount = await card.locator("ul > li").count();
    assert.ok(beforeCount >= 2, `Expected >= 2 sessions, got ${beforeCount}`);
    const otherRow = card
      .locator("ul > li")
      .filter({ hasNot: this.page.getByText("Perangkat ini", { exact: true }) })
      .first();
    await otherRow.waitFor({ state: "visible", timeout: 10_000 });
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/me/sessions/search") &&
          r.request().method() === "POST" &&
          r.status() === 200,
      ),
      otherRow.getByRole("button", { name: "Cabut sesi" }).click(),
    ]);
    await this.page.waitForFunction(
      ([el, expected]) => el.querySelectorAll("ul > li").length === expected,
      [await card.elementHandle(), beforeCount - 1] as const,
      { timeout: 10_000 },
    );
  }

  async expectLastAccountProtected(): Promise<void> {
    await this.page
      .getByText("Minimal 1 metode login harus tetap tertaut.", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
    const unlinkBtn = this.page
      .getByRole("button", { name: "Putuskan" })
      .first();
    assert.equal(await unlinkBtn.isDisabled(), true);
  }

  async logout(): Promise<void> {
    await this.page.getByRole("button", { name: "Akun saya" }).click();
    await this.page.getByRole("menuitem", { name: "Keluar" }).click();
    await this.expectPath("/login");
  }
}
