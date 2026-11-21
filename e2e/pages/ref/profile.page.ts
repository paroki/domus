import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export interface ParishProfileInput {
  name?: string;
  patronSaint?: string;
  pastorName?: string;
  phone?: string;
  email?: string;
  website?: string;
  address?: string;
}

export class ParishProfilePage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Profil Paroki" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByRole("button", { name: "Simpan Profil" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async updateProfile(input: ParishProfileInput): Promise<void> {
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

  async expectProfileValues(expected: ParishProfileInput): Promise<void> {
    if (expected.name !== undefined) {
      assert.equal(
        await this.page.getByLabel("Nama", { exact: true }).inputValue(),
        expected.name,
      );
    }
    if (expected.patronSaint !== undefined) {
      assert.equal(
        await this.page.getByLabel("Santo/Santa Pelindung").inputValue(),
        expected.patronSaint,
      );
    }
    if (expected.pastorName !== undefined) {
      assert.equal(
        await this.page.getByLabel("Pastor Paroki").inputValue(),
        expected.pastorName,
      );
    }
    if (expected.phone !== undefined) {
      assert.equal(
        await this.page.getByLabel("Telepon").inputValue(),
        expected.phone,
      );
    }
    if (expected.email !== undefined) {
      assert.equal(
        await this.page.getByLabel("Email", { exact: true }).inputValue(),
        expected.email,
      );
    }
    if (expected.website !== undefined) {
      assert.equal(
        await this.page.getByLabel("Situs Web").inputValue(),
        expected.website,
      );
    }
    if (expected.address !== undefined) {
      assert.equal(
        await this.page.getByLabel("Alamat").inputValue(),
        expected.address,
      );
    }
  }

  expectFieldValues(expected: ParishProfileInput): Promise<void> {
    return this.expectProfileValues(expected);
  }
}
