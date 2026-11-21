import { BasePage } from "../base.page";

export class OnboardingPage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Lengkapi Profil Anda" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByRole("heading", { name: "Data Diri" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async completeBiodataAndFinish(input: {
    name: string;
    phone?: string;
  }): Promise<void> {
    await this.page.getByLabel("Nama lengkap").fill(input.name);
    if (input.phone !== undefined) {
      await this.page.getByLabel("Nomor telepon (opsional)").fill(input.phone);
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
