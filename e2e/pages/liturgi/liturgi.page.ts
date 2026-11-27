import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class LiturgiPage extends BasePage {
  async expectSchedulesLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Perayaan Liturgi & Jadwal Misa" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectScheduleInDesktopMatrix(title: string): Promise<void> {
    await this.page.setViewportSize({ width: 1280, height: 800 });
    const table = this.page.locator("table");
    await table.waitFor({ state: "visible", timeout: 10_000 });
    const headers = await table.locator("thead th").allTextContents();
    assert.equal(
      headers.length,
      7,
      `Expected 7 columns in desktop matrix, got ${headers.length}: ${headers.join(", ")}`,
    );
    await table
      .getByRole("link", { name: title, exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectScheduleInMobileStack(title: string): Promise<void> {
    await this.page.setViewportSize({ width: 375, height: 812 });
    try {
      await this.page
        .locator("table")
        .waitFor({ state: "hidden", timeout: 10_000 });
      const mobileCard = this.page
        .locator("a")
        .filter({ has: this.page.getByRole("heading", { name: title }) });
      await mobileCard.first().waitFor({ state: "visible", timeout: 10_000 });
    } finally {
      await this.page.setViewportSize({ width: 1280, height: 720 });
    }
  }

  async filterSchedulesBy(
    label: "Semua Jadwal" | "Jadwal Tugas Saya" | "Perlu Pengganti",
  ): Promise<void> {
    await this.page.getByRole("button", { name: label, exact: true }).click();
  }

  async openScheduleFromList(title: string): Promise<void> {
    await this.page
      .locator("table")
      .getByRole("link", { name: title, exact: true })
      .click();
    await this.page.waitForURL(/\/liturgi\/[^/]+$/, { timeout: 10_000 });
  }

  async expectMinistriesLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Pengaturan 6 Bidang Petugas Liturgi" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async mapMinistry(
    roleLabel: string,
    opts: { orgName?: string; leaderQuery?: string; leaderName?: string },
  ): Promise<void> {
    const card = this.page.locator("div.flex.flex-col.gap-4.p-5").filter({
      has: this.page.getByRole("heading", { name: roleLabel, exact: true }),
    });
    await card.getByRole("button", { name: "Ubah Pemetaan" }).click();
    const form = card.locator("form");
    await form.waitFor({ state: "visible", timeout: 10_000 });

    if (opts.orgName) {
      const orgSelect = form.locator("select").nth(0);
      await orgSelect
        .locator("option", { hasText: opts.orgName })
        .waitFor({ state: "attached", timeout: 10_000 });
      await orgSelect.selectOption({ label: opts.orgName });
    }

    if (opts.leaderQuery) {
      await form
        .getByPlaceholder("Cari nama atau email…")
        .fill(opts.leaderQuery);
    }
    if (opts.leaderName) {
      const leaderSelect = form.locator("select").nth(1);
      const opt = leaderSelect
        .locator("option", { hasText: opts.leaderName })
        .first();
      await opt.waitFor({ state: "attached", timeout: 10_000 });
      const val = await opt.getAttribute("value");
      assert.ok(val, `Expected option value for leader ${opts.leaderName}`);
      await leaderSelect.selectOption(val);
    }

    await form.getByRole("button", { name: "Simpan Pemetaan" }).click();
    await form.waitFor({ state: "detached", timeout: 10_000 });
  }

  async createSchedule(input: {
    title: string;
    location?: string;
    startAt: string;
    endAt: string;
    description?: string;
  }): Promise<string> {
    await this.page
      .getByRole("heading", { name: "Buat Jadwal Misa", level: 1 })
      .waitFor({ state: "visible", timeout: 10_000 });

    await this.page.getByLabel("Judul Misa").fill(input.title);
    if (input.location) {
      await this.page.getByLabel("Lokasi").fill(input.location);
    }
    await this.page.getByLabel(/Waktu Mulai/).fill(input.startAt);
    await this.page.getByLabel(/Waktu Selesai/).fill(input.endAt);
    if (input.description) {
      await this.page.getByLabel("Catatan / Deskripsi").fill(input.description);
    }

    await this.page.getByRole("button", { name: "Simpan Jadwal Misa" }).click();
    await this.page.waitForURL(/\/liturgi\/[^/]+$/, { timeout: 10_000 });
    await this.expectDetailLoaded(input.title);
    const url = new URL(this.page.url());
    return url.pathname.split("/").pop() ?? "";
  }

  async expectDetailLoaded(title: string): Promise<void> {
    await this.page
      .getByRole("heading", { name: title, exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByRole("heading", { name: "Informasi & Waktu Perayaan Misa" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async submitRsvp(choice: "Hadir" | "Mungkin" | "Tidak Hadir"): Promise<void> {
    await this.page.getByRole("button", { name: choice, exact: true }).click();
    await this.page
      .getByText(`Status Anda: ${choice}`, { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async openAddPetugas(): Promise<void> {
    await this.page
      .getByRole("link", { name: /Tambah Petugas/ })
      .first()
      .click();
    await this.page
      .getByRole("heading", { name: "Tambah Petugas Liturgi", level: 1 })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectAllowedRoleOptions(expectedLabels: string[]): Promise<void> {
    const select = this.page.getByLabel("Bidang Pelayanan");
    await select.waitFor({ state: "visible", timeout: 10_000 });
    const options = await select.locator("option").allTextContents();
    assert.deepEqual(
      options.map((o) => o.trim()),
      expectedLabels,
      `Expected role options ${JSON.stringify(expectedLabels)}, got ${JSON.stringify(options)}`,
    );
  }

  async assignPetugas(input: {
    roleLabel?: string;
    slotLabel?: string;
    mode: "user" | "org" | "manual";
    memberQuery?: string;
    memberName?: string;
    orgName?: string;
    manualName?: string;
  }): Promise<void> {
    if (input.roleLabel) {
      await this.page
        .getByLabel("Bidang Pelayanan")
        .selectOption({ label: input.roleLabel });
    }
    if (input.slotLabel) {
      await this.page
        .getByLabel("Peran / Slot Tugas (opsional)")
        .fill(input.slotLabel);
    }

    const modeButtonLabel =
      input.mode === "user"
        ? "Anggota Terdaftar"
        : input.mode === "org"
          ? "Organisasi / Kelompok Kor"
          : "Nama Manual";
    await this.page
      .getByRole("button", { name: modeButtonLabel, exact: true })
      .click();

    if (input.mode === "user" && input.memberName) {
      if (input.memberQuery) {
        await this.page
          .getByPlaceholder("Cari nama anggota paroki…")
          .fill(input.memberQuery);
      }
      const memberSelect = this.page.locator("select").filter({
        has: this.page.locator("option", { hasText: "Pilih anggota paroki" }),
      });
      const opt = memberSelect
        .locator("option", { hasText: input.memberName })
        .first();
      await opt.waitFor({ state: "attached", timeout: 10_000 });
      const val = await opt.getAttribute("value");
      assert.ok(val, `Expected option value for member ${input.memberName}`);
      await memberSelect.selectOption(val);
    } else if (input.mode === "org" && input.orgName) {
      const orgSelect = this.page.locator("select").filter({
        has: this.page.locator("option", {
          hasText: "Pilih organisasi / kelompok",
        }),
      });
      await orgSelect.selectOption({ label: input.orgName });
    } else if (input.mode === "manual" && input.manualName) {
      await this.page.getByLabel("Nama Petugas").fill(input.manualName);
    }

    await this.page.getByRole("button", { name: "Simpan Petugas" }).click();
    await this.page.waitForURL(/\/liturgi\/[^/]+$/, { timeout: 10_000 });
  }

  private petugasRow(slotOrName: string) {
    return this.page.locator("li").filter({ hasText: slotOrName }).first();
  }

  async expectPetugasVisible(slotLabel: string, name: string): Promise<void> {
    const row = this.petugasRow(slotLabel);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row.getByText(name, { exact: true }).waitFor({
      state: "visible",
      timeout: 10_000,
    });
  }

  async openRespondForPetugas(slotOrName: string): Promise<void> {
    const row = this.petugasRow(slotOrName);
    await row
      .getByRole("link", { name: "Berhalangan / Ajukan Pengganti" })
      .click();
    await this.page
      .getByRole("heading", { name: "Konfirmasi Tugas & Usulan Pengganti" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async submitDeclineWithSubstitute(input: {
    note: string;
    subMode: "user" | "manual";
    memberQuery?: string;
    memberName?: string;
    substituteName?: string;
  }): Promise<void> {
    await this.page
      .getByRole("button", { name: "Berhalangan", exact: true })
      .click();
    await this.page.getByLabel("Catatan / Alasan (opsional)").fill(input.note);

    if (input.subMode === "user" && input.memberName) {
      await this.page
        .getByRole("button", { name: "Anggota Terdaftar", exact: true })
        .click();
      if (input.memberQuery) {
        await this.page
          .getByPlaceholder("Cari nama anggota paroki…")
          .fill(input.memberQuery);
      }
      const memberSelect = this.page.locator("select").filter({
        has: this.page.locator("option", { hasText: "Pilih anggota paroki" }),
      });
      const opt = memberSelect
        .locator("option", { hasText: input.memberName })
        .first();
      await opt.waitFor({ state: "attached", timeout: 10_000 });
      const val = await opt.getAttribute("value");
      assert.ok(
        val,
        `Expected option value for substitute ${input.memberName}`,
      );
      await memberSelect.selectOption(val);
    } else if (input.subMode === "manual" && input.substituteName) {
      await this.page
        .getByRole("button", { name: "Nama Manual", exact: true })
        .click();
      await this.page
        .getByPlaceholder("Masukkan nama petugas atau kelompok")
        .fill(input.substituteName);
    }

    await this.page.getByRole("button", { name: "Kirim Konfirmasi" }).click();
    await this.page.waitForURL(/\/liturgi\/[^/]+$/, { timeout: 10_000 });
  }

  async expectSubstituteCallout(
    slotOrName: string,
    substituteName: string,
  ): Promise<void> {
    const row = this.petugasRow(slotOrName);
    await row
      .getByText(new RegExp(`Usulan Pengganti.*${substituteName}`))
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async approveSubstitute(slotOrName: string): Promise<void> {
    const row = this.petugasRow(slotOrName);
    const btn = row.getByRole("button", { name: "Setujui Pengganti" });
    await btn.click();
    await btn.waitFor({ state: "detached", timeout: 10_000 });
  }

  async openAddLagu(): Promise<void> {
    await this.page.getByRole("link", { name: /Tambah Lagu/ }).click();
    await this.page
      .getByRole("heading", { name: "Tambah Lagu Perayaan", level: 1 })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async addLaguFromCantus(input: {
    bagian: string;
    searchQuery: string;
    cantusTitle: string;
    expectedSumber: string;
    expectedNomor: string;
    expectedNadaDasar: string;
  }): Promise<void> {
    await this.page
      .getByPlaceholder("Cari judul, nomor (PS-320), atau sumber…")
      .fill(input.searchQuery);
    const resultBtn = this.page
      .locator("button")
      .filter({ hasText: input.cantusTitle })
      .first();
    await resultBtn.waitFor({ state: "visible", timeout: 10_000 });
    await resultBtn.click();

    await this.page
      .getByText("Terhubung ke Katalog Cantus", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });

    assert.equal(
      await this.page.getByLabel("Sumber Buku (opsional)").inputValue(),
      input.expectedSumber,
    );
    assert.equal(
      await this.page.getByLabel("Nomor Lagu (opsional)").inputValue(),
      input.expectedNomor,
    );
    assert.equal(
      await this.page.getByLabel("Judul Lagu").inputValue(),
      input.cantusTitle,
    );
    assert.equal(
      await this.page.getByLabel("Nada Dasar (opsional)").inputValue(),
      input.expectedNadaDasar,
    );

    await this.page.getByLabel("Bagian Liturgi").selectOption(input.bagian);
    await this.page.getByRole("button", { name: "Simpan Lagu" }).click();
    await this.page.waitForURL(/\/liturgi\/[^/]+$/, { timeout: 10_000 });
  }

  async addLaguManual(input: {
    bagian: string;
    judul: string;
    sumber?: string;
    nomor?: string;
    nadaDasar?: string;
    note?: string;
  }): Promise<void> {
    await this.page.getByLabel("Bagian Liturgi").selectOption(input.bagian);
    if (input.sumber) {
      await this.page.getByLabel("Sumber Buku (opsional)").fill(input.sumber);
    }
    if (input.nomor) {
      await this.page.getByLabel("Nomor Lagu (opsional)").fill(input.nomor);
    }
    await this.page.getByLabel("Judul Lagu").fill(input.judul);
    if (input.nadaDasar) {
      await this.page.getByLabel("Nada Dasar (opsional)").fill(input.nadaDasar);
    }
    if (input.note) {
      await this.page.getByLabel("Catatan Lagu (opsional)").fill(input.note);
    }
    await this.page.getByRole("button", { name: "Simpan Lagu" }).click();
    await this.page.waitForURL(/\/liturgi\/[^/]+$/, { timeout: 10_000 });
  }

  async expectLaguVisible(
    judul: string,
    opts?: { catalogLinked?: boolean },
  ): Promise<void> {
    if (opts?.catalogLinked) {
      const link = this.page
        .locator('a[href^="/cantus/"]')
        .filter({ hasText: judul });
      await link.waitFor({ state: "visible", timeout: 10_000 });
    } else {
      await this.page
        .getByText(judul, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }
}
