import assert from "node:assert/strict";
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
}

export class ParokiProfilePage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Profil Paroki" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByRole("button", { name: "Simpan Profil" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async updateProfile(input: ParokiProfileInput): Promise<void> {
    if (input.name !== undefined) {
      await this.page.getByLabel("Nama", { exact: true }).fill(input.name);
    }
    if (input.patronSaint !== undefined) {
      await this.page
        .getByLabel("Santo/Santa Pelindung")
        .fill(input.patronSaint);
    }
    if (input.pastorName !== undefined) {
      await this.page.getByLabel("Pastor Paroki").fill(input.pastorName);
    }
    if (input.phone !== undefined) {
      await this.page.getByLabel("Telepon").fill(input.phone);
    }
    if (input.email !== undefined) {
      await this.page.getByLabel("Email", { exact: true }).fill(input.email);
    }
    if (input.website !== undefined) {
      await this.page.getByLabel("Situs Web").fill(input.website);
    }
    if (input.address !== undefined) {
      await this.page.getByLabel("Alamat").fill(input.address);
    }
    await this.page.getByRole("button", { name: "Simpan Profil" }).click();
  }

  async expectSavedBanner(): Promise<void> {
    await this.page
      .getByText("Tersimpan", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  private async assertInputValue(
    locator: Locator,
    expected: string,
  ): Promise<void> {
    const deadline = Date.now() + 10_000;
    while (Date.now() < deadline) {
      if ((await locator.inputValue()) === expected) return;
      await this.page.waitForTimeout(100);
    }
    assert.equal(await locator.inputValue(), expected);
  }

  async expectProfileValues(expected: ParokiProfileInput): Promise<void> {
    if (expected.name !== undefined) {
      await this.assertInputValue(
        this.page.getByLabel("Nama", { exact: true }),
        expected.name,
      );
    }
    if (expected.patronSaint !== undefined) {
      await this.assertInputValue(
        this.page.getByLabel("Santo/Santa Pelindung"),
        expected.patronSaint,
      );
    }
    if (expected.pastorName !== undefined) {
      await this.assertInputValue(
        this.page.getByLabel("Pastor Paroki"),
        expected.pastorName,
      );
    }
    if (expected.phone !== undefined) {
      await this.assertInputValue(
        this.page.getByLabel("Telepon"),
        expected.phone,
      );
    }
    if (expected.email !== undefined) {
      await this.assertInputValue(
        this.page.getByLabel("Email", { exact: true }),
        expected.email,
      );
    }
    if (expected.website !== undefined) {
      await this.assertInputValue(
        this.page.getByLabel("Situs Web"),
        expected.website,
      );
    }
    if (expected.address !== undefined) {
      await this.assertInputValue(
        this.page.getByLabel("Alamat"),
        expected.address,
      );
    }
  }

  expectFieldValues(expected: ParokiProfileInput): Promise<void> {
    return this.expectProfileValues(expected);
  }
}
