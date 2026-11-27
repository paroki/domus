import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class EventDetailPage extends BasePage {
  async expectLoaded(title: string): Promise<void> {
    await this.page
      .getByRole("heading", { name: title })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectBadges(badges: string[]): Promise<void> {
    for (const badge of badges) {
      await this.page
        .locator("span")
        .filter({ hasText: new RegExp(`^${badge}$`) })
        .first()
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectLocationVisible(location: string): Promise<void> {
    await this.page
      .getByText(location, { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectTimezoneVisible(tzLabel: string): Promise<void> {
    await this.page
      .getByText(tzLabel, { exact: false })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectSummaryDenominatorText(text: string): Promise<void> {
    await this.page
      .getByText(text, { exact: false })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectAdminActionsVisible(visible: boolean): Promise<void> {
    const editBtn = this.page.getByRole("button", { name: "Ubah Kegiatan" });
    const deleteBtn = this.page.getByRole("button", { name: "Hapus Kegiatan" });
    if (visible) {
      await editBtn.waitFor({ state: "visible", timeout: 10_000 });
      await deleteBtn.waitFor({ state: "visible", timeout: 10_000 });
    } else {
      assert.equal(await editBtn.count(), 0);
      assert.equal(await deleteBtn.count(), 0);
    }
  }

  async editEvent(input: {
    title?: string;
    jenisLabel?: string;
    location?: string;
    visibilityLabel?: "Publik Paroki" | "Internal Organisasi";
    description?: string;
    rsvpEnabled?: boolean;
  }): Promise<void> {
    await this.page.getByRole("button", { name: "Ubah Kegiatan" }).click();
    const form = this.page.locator("form").filter({
      has: this.page.getByRole("button", { name: "Simpan Perubahan" }),
    });
    await form.waitFor({ state: "visible", timeout: 10_000 });

    if (input.title !== undefined) {
      await form.getByLabel("Judul Kegiatan").fill(input.title);
    }
    if (input.jenisLabel !== undefined) {
      await form
        .getByLabel("Jenis Kegiatan")
        .selectOption({ label: input.jenisLabel });
    }
    if (input.location !== undefined) {
      await form.getByLabel("Lokasi").fill(input.location);
    }
    if (input.visibilityLabel) {
      const visSelect = form.getByLabel("Visibilitas");
      if (await visSelect.isEnabled()) {
        await visSelect.selectOption({ label: input.visibilityLabel });
      }
    }
    if (input.description !== undefined) {
      await form.getByLabel("Deskripsi").fill(input.description);
    }
    if (input.rsvpEnabled !== undefined) {
      const checkbox = form.getByLabel("Aktifkan konfirmasi kehadiran (RSVP)");
      if (input.rsvpEnabled) await checkbox.check();
      else await checkbox.uncheck();
    }

    await form.getByRole("button", { name: "Simpan Perubahan" }).click();
    await form.waitFor({ state: "detached", timeout: 10_000 });
  }

  async cancelEvent(): Promise<void> {
    this.page.once("dialog", (dialog) => dialog.accept());
    await this.page.getByRole("button", { name: "Batalkan Kegiatan" }).click();
  }

  async rescheduleEvent(): Promise<void> {
    this.page.once("dialog", (dialog) => dialog.accept());
    await this.page.getByRole("button", { name: "Jadwalkan Kembali" }).click();
  }

  async deleteEvent(): Promise<void> {
    this.page.once("dialog", (dialog) => dialog.accept());
    await this.page.getByRole("button", { name: "Hapus Kegiatan" }).click();
  }

  async submitRsvp(
    label: "Hadir" | "Mungkin" | "Tidak Hadir",
    note?: string,
  ): Promise<void> {
    if (note !== undefined) {
      await this.page
        .getByPlaceholder("Catatan kehadiran (opsional)")
        .fill(note);
    }
    await this.page.getByRole("button", { name: label, exact: true }).click();
  }

  async removeRsvp(): Promise<void> {
    this.page.once("dialog", (dialog) => dialog.accept());
    await this.page.getByRole("button", { name: "Hapus RSVP" }).click();
  }

  async expectRsvpClosed(): Promise<void> {
    await this.page
      .getByText("Konfirmasi kehadiran (RSVP) telah ditutup")
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectRsvpOpen(): Promise<void> {
    await this.page
      .getByRole("button", { name: "Hadir", exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectRsvpRestrictedBanner(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Khusus Anggota Organisasi" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectSummaryCounts(counts: {
    going: number;
    maybe: number;
    notGoing: number;
    unresponded?: number;
  }): Promise<void> {
    const grid = this.page.locator("div.grid.sm\\:grid-cols-4");
    await grid.waitFor({ state: "visible", timeout: 10_000 });
    const cards = grid.locator("> div");
    assert.equal(
      (await cards.nth(0).locator("p").nth(1).textContent())?.trim(),
      String(counts.going),
    );
    assert.equal(
      (await cards.nth(1).locator("p").nth(1).textContent())?.trim(),
      String(counts.maybe),
    );
    assert.equal(
      (await cards.nth(2).locator("p").nth(1).textContent())?.trim(),
      String(counts.notGoing),
    );
    if (counts.unresponded !== undefined) {
      assert.equal(
        (await cards.nth(3).locator("p").nth(1).textContent())?.trim(),
        String(counts.unresponded),
      );
    }
  }

  async expectAttendeeVisible(
    name: string,
    statusLabel: string,
    note?: string,
  ): Promise<void> {
    const row = this.page
      .locator("div.divide-y > div")
      .filter({ hasText: name });
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row
      .getByText(statusLabel, { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
    if (note) {
      await row
        .getByText(note, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectRsvpStatus(statusLabel: string, note?: string): Promise<void> {
    await this.page
      .getByText(`Status: ${statusLabel}`, { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
    if (note) {
      await this.page
        .getByText(note, { exact: true })
        .first()
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectRsvpRemoved(_memberName: string): Promise<void> {
    await this.page
      .getByText(/^Status: /)
      .waitFor({ state: "detached", timeout: 10_000 });
    await this.page
      .getByRole("button", { name: "Hapus RSVP" })
      .waitFor({ state: "detached", timeout: 10_000 });
    await this.expectSummaryCounts({
      going: 0,
      maybe: 0,
      notGoing: 0,
    });
  }
}
