import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class MePage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Profil & Keamanan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async updateProfile(input: {
    name: string;
    phone?: string;
    gender?: "pria" | "wanita";
    birthDate?: string;
    timezone?: string;
  }): Promise<void> {
    await this.page.getByLabel("Nama lengkap").fill(input.name);
    if (input.phone !== undefined) {
      await this.page.getByLabel("Nomor telepon (WhatsApp)").fill(input.phone);
    }
    await this.page
      .getByLabel("Jenis kelamin")
      .selectOption(input.gender ?? "pria");
    await this.page
      .getByLabel("Tanggal lahir")
      .fill(input.birthDate ?? "1995-05-10");
    if (input.timezone !== undefined) {
      await this.page.getByLabel("Zona waktu").selectOption(input.timezone);
    }
    await this.page.getByRole("button", { name: "Simpan Profil" }).click();
  }

  async expectProfileValues(input: {
    gender: string;
    birthDate: string;
    timezone?: string;
  }): Promise<void> {
    const genderVal = await this.page.getByLabel("Jenis kelamin").inputValue();
    const birthDateVal = await this.page
      .getByLabel("Tanggal lahir")
      .inputValue();
    assert.equal(genderVal, input.gender);
    assert.equal(birthDateVal, input.birthDate);
    if (input.timezone !== undefined) {
      const tzVal = await this.page.getByLabel("Zona waktu").inputValue();
      assert.equal(tzVal, input.timezone);
    }
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

  async createToken(name: string, expiryValue = "30"): Promise<string> {
    await this.page.getByLabel("Nama token").fill(name);
    await this.page.getByLabel("Masa berlaku").selectOption(expiryValue);
    await this.page.getByRole("button", { name: "Buat Token" }).click();
    const banner = this.page.getByTestId("pat-created-banner");
    await banner.waitFor({ state: "visible", timeout: 10_000 });
    const raw = (
      (await this.page.getByTestId("pat-plaintext-token").textContent()) ?? ""
    ).trim();
    assert.ok(raw.startsWith("dms_"), `Expected dms_ prefix, got ${raw}`);
    return raw;
  }

  async expectTokenVisible(name: string): Promise<void> {
    await this.page
      .getByTestId(`pat-row-${name}`)
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectDeviceNotificationCardVisible(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Aplikasi & Notifikasi Perangkat" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByRole("button", { name: "Kirim Tes Notifikasi" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async setNotificationChannel(label: string, enabled: boolean): Promise<void> {
    const checkbox = this.page.getByLabel(label, { exact: true });
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().endsWith("/api/me") &&
          r.request().method() === "PATCH" &&
          r.status() === 200,
      ),
      enabled ? checkbox.check() : checkbox.uncheck(),
    ]);
  }

  async expectNotificationChannelChecked(
    label: string,
    checked: boolean,
  ): Promise<void> {
    const actual = await this.page
      .getByLabel(label, { exact: true })
      .isChecked();
    assert.equal(actual, checked);
  }

  async sendTestNotification(): Promise<void> {
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().endsWith("/api/me/push/test") &&
          r.request().method() === "POST" &&
          r.status() === 204,
      ),
      this.page.getByRole("button", { name: "Kirim Tes Notifikasi" }).click(),
    ]);
  }

  async expectTestNotificationSent(): Promise<void> {
    await this.page
      .getByText("Tes notifikasi telah dikirim.", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async revokeToken(name: string): Promise<void> {
    const row = this.page.getByTestId(`pat-row-${name}`);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/me/tokens/search") &&
          r.request().method() === "POST" &&
          r.status() === 200,
      ),
      row.getByRole("button", { name: "Cabut token" }).click(),
    ]);
    await row.waitFor({ state: "detached", timeout: 10_000 });
  }
}
