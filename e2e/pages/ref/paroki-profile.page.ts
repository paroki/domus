import type { Locator } from "playwright";
import { BasePage } from "../base.page";

export interface ParokiProfileInput {
  name?: string;
  patronSaint?: string;
  pastorName?: string;
  phone?: string;
  email?: string;
  website?: string;
  address?: string;
  timezone?: string;
  latitude?: string;
  longitude?: string;
}

export class ParokiProfilePage extends BasePage {
  private sectionCard(heading: string | RegExp): Locator {
    return this.page
      .locator("div.rounded-xl")
      .filter({ has: this.page.getByRole("heading", { name: heading }) })
      .first();
  }

  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Profil Paroki" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByRole("heading", { name: "Struktur Gerejawi" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async updateProfile(input: ParokiProfileInput): Promise<void> {
    if (
      input.name !== undefined ||
      input.patronSaint !== undefined ||
      input.pastorName !== undefined
    ) {
      const card = this.sectionCard("Identitas Gereja & Kop Surat");
      await card.getByRole("button", { name: "Ubah" }).click();
      if (input.name !== undefined) {
        await this.page.locator("#sec-nama").fill(input.name);
      }
      if (input.patronSaint !== undefined) {
        await this.page.locator("#sec-pelindung").fill(input.patronSaint);
      }
      if (input.pastorName !== undefined) {
        await this.page.locator("#sec-pastor").fill(input.pastorName);
      }
      await this.page.getByRole("button", { name: "Simpan Perubahan" }).click();
      await this.expectSavedBanner();
    }

    if (
      input.phone !== undefined ||
      input.email !== undefined ||
      input.website !== undefined
    ) {
      const card = this.sectionCard("Kontak Resmi");
      await card.getByRole("button", { name: "Ubah" }).click();
      if (input.phone !== undefined) {
        await this.page.locator("#sec-telepon").fill(input.phone);
      }
      if (input.email !== undefined) {
        await this.page.locator("#sec-email").fill(input.email);
      }
      if (input.website !== undefined) {
        await this.page.locator("#sec-website").fill(input.website);
      }
      await this.page.getByRole("button", { name: "Simpan Perubahan" }).click();
      await this.expectSavedBanner();
    }

    if (
      input.address !== undefined ||
      input.timezone !== undefined ||
      input.latitude !== undefined ||
      input.longitude !== undefined
    ) {
      const card = this.sectionCard(
        "Domisili Wilayah Sipil, Zona Waktu & Koordinat",
      );
      await card.getByRole("button", { name: "Ubah" }).click();
      if (input.timezone !== undefined) {
        await this.page.locator("#sec-timezone").selectOption(input.timezone);
      }
      if (input.address !== undefined) {
        await this.page.locator("#sec-alamat").fill(input.address);
      }
      if (input.latitude !== undefined) {
        await this.page.locator("#sec-lat").fill(input.latitude);
      }
      if (input.longitude !== undefined) {
        await this.page.locator("#sec-lng").fill(input.longitude);
      }
      await this.page.getByRole("button", { name: "Simpan Perubahan" }).click();
      await this.expectSavedBanner();
    }
  }

  async expectSavedBanner(): Promise<void> {
    await this.page
      .getByRole("status")
      .filter({ hasText: "Tersimpan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectProfileValues(expected: ParokiProfileInput): Promise<void> {
    const vals = [
      expected.name,
      expected.patronSaint,
      expected.pastorName,
      expected.phone,
      expected.email,
      expected.website,
      expected.address,
      expected.timezone,
    ].filter((v): v is string => v !== undefined);
    for (const v of vals) {
      await this.page
        .getByText(v, { exact: false })
        .first()
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  expectFieldValues(expected: ParokiProfileInput): Promise<void> {
    return this.expectProfileValues(expected);
  }
}
