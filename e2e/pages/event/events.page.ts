import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class EventsPage extends BasePage {
  private cardFor(title: string) {
    return this.page.locator("div.space-y-4 > div").filter({
      has: this.page.getByRole("link", { name: title, exact: true }),
    });
  }

  private filterBar() {
    return this.page.locator("div.flex.flex-wrap.items-center.gap-3");
  }

  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Kegiatan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectCreateButtonVisible(visible: boolean): Promise<void> {
    const btn = this.page.getByRole("button", { name: "Buat Kegiatan" });
    if (visible) {
      await btn.waitFor({ state: "visible", timeout: 10_000 });
    } else {
      assert.equal(await btn.count(), 0);
    }
  }

  async createEvent(input: {
    orgLabel?: string;
    visibilityLabel?: "Publik Paroki" | "Internal Organisasi";
    title: string;
    startAt: string;
    endAt: string;
    location?: string;
    rsvpDeadline?: string;
    description?: string;
    rsvpEnabled?: boolean;
  }): Promise<void> {
    const form = this.page.locator("form").filter({
      has: this.page.getByLabel("Judul Kegiatan"),
    });
    if ((await form.count()) === 0) {
      await this.page.getByRole("button", { name: "Buat Kegiatan" }).click();
      await form.waitFor({ state: "visible", timeout: 10_000 });
    }

    if (input.orgLabel) {
      await form
        .getByLabel("Penyelenggara")
        .selectOption({ label: input.orgLabel });
    }
    if (input.visibilityLabel) {
      await form
        .getByLabel("Visibilitas")
        .selectOption({ label: input.visibilityLabel });
    }
    await form.getByLabel("Judul Kegiatan").fill(input.title);
    await form.getByLabel("Waktu Mulai").fill(input.startAt);
    await form.getByLabel("Waktu Selesai").fill(input.endAt);
    if (input.location !== undefined) {
      await form.getByLabel("Lokasi (opsional)").fill(input.location);
    }
    if (input.rsvpDeadline !== undefined) {
      await form
        .getByLabel("Batas Waktu RSVP (opsional)")
        .fill(input.rsvpDeadline);
    }
    if (input.description !== undefined) {
      await form.getByLabel("Deskripsi (opsional)").fill(input.description);
    }
    if (input.rsvpEnabled !== undefined) {
      const checkbox = form.getByLabel(
        "Aktifkan konfirmasi kehadiran (RSVP) anggota organisasi",
      );
      if (input.rsvpEnabled) await checkbox.check();
      else await checkbox.uncheck();
    }

    await form.getByRole("button", { name: "Simpan Kegiatan" }).click();
  }

  async expectFormClosed(): Promise<void> {
    const form = this.page.locator("form").filter({
      has: this.page.getByLabel("Judul Kegiatan"),
    });
    await form.waitFor({ state: "detached", timeout: 10_000 });
  }

  async expectFormError(message: string): Promise<void> {
    await this.page
      .getByText(message, { exact: false })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async filterByOrganizer(label: string): Promise<void> {
    await this.filterBar().locator("select").nth(0).selectOption({ label });
  }

  async filterByVisibility(label: string): Promise<void> {
    await this.filterBar().locator("select").nth(1).selectOption({ label });
  }

  async filterByStatus(label: string): Promise<void> {
    await this.filterBar().locator("select").nth(2).selectOption({ label });
  }

  async toggleMyOrgsOnly(myOrgsOnly: boolean): Promise<void> {
    await this.filterBar()
      .getByRole("button", {
        name: myOrgsOnly ? "Organisasi Saya" : "Semua Kegiatan",
      })
      .click();
  }

  async search(query: string): Promise<void> {
    await this.page
      .getByPlaceholder("Cari judul, lokasi, atau deskripsi kegiatan…")
      .fill(query);
  }

  async expectEventCardVisible(
    title: string,
    options?: {
      visibilityBadge?: string;
      statusBadge?: string;
      location?: string;
    },
  ): Promise<void> {
    const card = this.cardFor(title);
    await card.waitFor({ state: "visible", timeout: 10_000 });
    if (options?.visibilityBadge) {
      await card
        .getByText(options.visibilityBadge, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.statusBadge) {
      await card
        .getByText(options.statusBadge, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.location) {
      await card
        .getByText(options.location, { exact: false })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectTimeRangeTimezone(title: string, tzLabel: string): Promise<void> {
    const card = this.cardFor(title);
    await card.waitFor({ state: "visible", timeout: 10_000 });
    await card
      .getByText(tzLabel, { exact: false })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectEventCardNotVisible(title: string): Promise<void> {
    await this.cardFor(title).waitFor({ state: "detached", timeout: 10_000 });
  }

  async quickRsvp(
    title: string,
    label: "Hadir" | "Mungkin" | "Tidak Hadir",
  ): Promise<void> {
    const card = this.cardFor(title);
    await card.getByRole("button", { name: label, exact: true }).click();
  }

  async openEventDetail(title: string): Promise<void> {
    await this.page.getByRole("link", { name: title, exact: true }).click();
  }
}
