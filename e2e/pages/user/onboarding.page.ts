import assert from "node:assert/strict";
import { BasePage } from "../base.page";

function isoYearsAgo(years: number): string {
  const d = new Date();
  d.setUTCFullYear(d.getUTCFullYear() - years);
  return d.toISOString().slice(0, 10);
}

export class OnboardingPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Lengkapi Profil Anda" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByRole("heading", { name: "Data Diri" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectBirthDateBounds(): Promise<void> {
    const input = this.page.getByLabel("Tanggal lahir");
    const min = await input.getAttribute("min");
    const max = await input.getAttribute("max");
    assert.equal(min, isoYearsAgo(120));
    assert.equal(max, isoYearsAgo(13));
  }

  async completeBiodataAndFinish(input: {
    name: string;
    phone?: string;
    gender?: "pria" | "wanita";
    birthDate?: string;
    timezone?: string;
  }): Promise<void> {
    await this.page.getByLabel("Nama lengkap").fill(input.name);
    if (input.phone !== undefined) {
      await this.page.getByLabel("Nomor telepon (opsional)").fill(input.phone);
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

    const finishBtn = this.page.getByRole("button", {
      name: "Simpan & Selesai",
    });
    if ((await finishBtn.count()) > 0) {
      await finishBtn.click();
    } else {
      await this.page.getByRole("button", { name: "Simpan & Lanjut" }).click();
      for (let i = 0; i < 3; i++) {
        if (!new URL(this.page.url()).pathname.startsWith("/onboarding")) {
          break;
        }
        const skipBtn = this.page.getByRole("button", {
          name: /Lewati|Selesai/,
        });
        if ((await skipBtn.count()) > 0) {
          await skipBtn.first().click();
        }
      }
    }

    await this.page.waitForURL(
      (url) => !url.pathname.startsWith("/onboarding"),
      { timeout: 10_000 },
    );
  }
}
