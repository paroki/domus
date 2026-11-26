import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class OrgDetailPage extends BasePage {
  private inviteForm() {
    return this.page.locator("form").filter({
      has: this.page.getByRole("button", { name: "Buat Tautan Undangan" }),
    });
  }

  private inviteRows() {
    return this.page.locator("div.divide-y > div").filter({
      has: this.page.getByRole("button", { name: "Salin Tautan" }),
    });
  }

  async expectLoaded(orgName: string): Promise<void> {
    await this.page
      .getByRole("heading", { name: orgName })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectEditAndDeleteVisibility(options: {
    canEdit: boolean;
    canDelete: boolean;
  }): Promise<void> {
    const editBtn = this.page.getByRole("button", {
      name: "Ubah Profil Organisasi",
    });
    const deleteBtn = this.page.getByRole("button", {
      name: "Hapus Organisasi",
    });

    if (options.canEdit) {
      await editBtn.waitFor({ state: "visible", timeout: 10_000 });
    } else {
      assert.equal(await editBtn.count(), 0);
    }

    if (options.canDelete) {
      await deleteBtn.waitFor({ state: "visible", timeout: 10_000 });
    } else {
      assert.equal(await deleteBtn.count(), 0);
    }
  }

  async updateProfile(input: {
    name?: string;
    description?: string;
    active?: boolean;
  }): Promise<void> {
    await this.page
      .getByRole("button", { name: "Ubah Profil Organisasi" })
      .click();
    if (input.name !== undefined) {
      await this.page.getByLabel("Nama Organisasi").fill(input.name);
    }
    if (input.description !== undefined) {
      await this.page.getByLabel("Deskripsi").fill(input.description);
    }
    if (input.active !== undefined) {
      const checkbox = this.page.getByLabel("Organisasi Aktif");
      if (input.active) await checkbox.check();
      else await checkbox.uncheck();
    }
    await this.page.getByRole("button", { name: "Simpan Perubahan" }).click();
    await this.page
      .getByRole("button", { name: "Simpan Perubahan" })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async expectDescriptionVisible(text: string): Promise<void> {
    await this.page
      .getByText(text, { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectRosterReadOnly(memberName: string): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Daftar Anggota" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByText(memberName, { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByRole("button", { name: "Keluar dari Organisasi" })
      .waitFor({ state: "visible", timeout: 10_000 });

    for (const forbidden of [
      "Tambah Anggota",
      "Ubah Jabatan",
      "Jadikan Admin",
      "Jadikan Anggota",
    ]) {
      assert.equal(
        await this.page.getByRole("button", { name: forbidden }).count(),
        0,
        `Expected "${forbidden}" button not to be rendered on /org/:orgId`,
      );
    }
  }

  async createInvite(input: {
    roleLabel?: "Anggota (member)" | "Admin Organisasi (admin)";
    ttlLabel?: string;
    maxUses?: number;
  }): Promise<void> {
    const form = this.inviteForm();
    if (input.roleLabel) {
      await form
        .locator("select")
        .first()
        .selectOption({ label: input.roleLabel });
    }
    if (input.ttlLabel) {
      await form
        .locator("select")
        .nth(1)
        .selectOption({ label: input.ttlLabel });
    }
    if (input.maxUses !== undefined) {
      await form
        .getByPlaceholder("Maks. pakai (opsional)")
        .fill(String(input.maxUses));
    }
    await form.getByRole("button", { name: "Buat Tautan Undangan" }).click();
  }

  async expectInviteVisible(options: {
    statusLabel: string;
    maxUsesText?: string;
  }): Promise<string> {
    let row = this.inviteRows().filter({
      has: this.page.getByText(options.statusLabel, { exact: true }),
    });
    if (options.maxUsesText) {
      row = row.filter({ hasText: options.maxUsesText });
    }
    const target = row.first();
    await target.waitFor({ state: "visible", timeout: 10_000 });
    const codeText = await target.locator("span.font-mono").textContent();
    return (codeText ?? "").trim();
  }

  async revokeInvite(code?: string): Promise<void> {
    const rows = this.inviteRows();
    const scope = code ? rows.filter({ hasText: code }) : rows;
    await scope.first().getByRole("button", { name: "Cabut Tautan" }).click();
  }

  async leaveOrganization(): Promise<void> {
    this.page.once("dialog", (dialog) => dialog.accept());
    await this.page
      .getByRole("button", { name: "Keluar dari Organisasi" })
      .click();
  }

  async expectErrorAlert(message: string): Promise<void> {
    await this.page
      .getByRole("alert")
      .filter({ hasText: message })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectInvitesSectionVisible(visible: boolean): Promise<void> {
    const heading = this.page.getByRole("heading", {
      name: "Tautan Undangan",
    });
    if (visible) {
      await heading.waitFor({ state: "visible", timeout: 10_000 });
    } else {
      assert.equal(await heading.count(), 0);
    }
  }

  async expectRosterRestrictedBanner(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Keanggotaan Diperlukan" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByText(
        "Hanya anggota atau pengurus organisasi ini yang dapat melihat daftar anggota internal.",
      )
      .waitFor({ state: "visible", timeout: 10_000 });
    assert.equal(
      await this.page.getByRole("heading", { name: "Daftar Anggota" }).count(),
      0,
    );
  }

  private periodeCard() {
    return this.page.locator("div").filter({
      has: this.page.getByRole("heading", { name: "Periode Kepengurusan" }),
    });
  }

  private periodeRow(name: string) {
    return this.periodeCard()
      .locator("div.divide-y > div")
      .filter({ hasText: name })
      .first();
  }

  private jabatanCard() {
    return this.page.locator("div").filter({
      has: this.page.getByRole("heading", {
        name: "Struktur & Jabatan Organisasi",
      }),
    });
  }

  private jabatanRow(jabatan: string) {
    return this.jabatanCard()
      .locator("div.divide-y > div")
      .filter({ hasText: jabatan })
      .first();
  }

  async openNewPeriodeForm(): Promise<void> {
    await this.page.getByRole("button", { name: "Tambah Periode" }).click();
    await this.page
      .getByRole("heading", { name: "Tambah Periode Kepengurusan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async fillAndSubmitPeriodeForm(input: {
    nama: string;
    kode?: string;
    tanggalMulai: string;
    tanggalSelesai: string;
    isCurrent?: boolean;
  }): Promise<void> {
    await this.page.getByLabel("Nama Periode").fill(input.nama);
    if (input.kode !== undefined) {
      await this.page.getByLabel("Kode Periode").fill(input.kode);
    }
    await this.page.getByLabel("Tanggal Mulai").fill(input.tanggalMulai);
    await this.page.getByLabel("Tanggal Selesai").fill(input.tanggalSelesai);
    if (input.isCurrent !== undefined) {
      const cb = this.page.getByLabel("Jadikan sebagai Periode Aktif saat ini");
      if (input.isCurrent) await cb.check();
      else await cb.uncheck();
    }
    await this.page.getByRole("button", { name: "Simpan Periode" }).click();
  }

  async expectPeriodeVisible(
    name: string,
    options?: { status?: "Terbuka" | "Ditutup"; isCurrent?: boolean },
  ): Promise<void> {
    const row = this.periodeRow(name);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    if (options?.status) {
      await row
        .getByText(options.status, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.isCurrent === true) {
      await row
        .getByText("Periode Aktif", { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    } else if (options?.isCurrent === false) {
      assert.equal(
        await row.getByText("Periode Aktif", { exact: true }).count(),
        0,
        `Expected "${name}" not to have Periode Aktif badge`,
      );
    }
  }

  async expectPeriodeNotVisible(name: string): Promise<void> {
    await this.periodeRow(name).waitFor({
      state: "detached",
      timeout: 10_000,
    });
  }

  async makePeriodeCurrent(name: string): Promise<void> {
    await this.periodeRow(name)
      .getByRole("button", { name: "Jadikan Aktif" })
      .click();
  }

  async togglePeriodeStatus(
    name: string,
    buttonLabel: "Tutup" | "Buka",
  ): Promise<void> {
    await this.periodeRow(name)
      .getByRole("button", { name: buttonLabel, exact: true })
      .click();
  }

  async deletePeriode(name: string): Promise<void> {
    await this.periodeRow(name).getByRole("button", { name: "Hapus" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Hapus" })
      .click();
  }

  async createJabatan(input: {
    parentJabatan?: string;
    jabatan: string;
    source: "member" | "umat" | "manual";
    memberName?: string;
    umatName?: string;
    manualNama?: string;
    urutan?: number;
  }): Promise<void> {
    if (input.parentJabatan) {
      await this.jabatanRow(input.parentJabatan)
        .getByRole("button", { name: "Tambah Sub-Jabatan" })
        .click();
    } else {
      await this.page.getByRole("button", { name: "Tambah Jabatan" }).click();
    }
    await this.page
      .getByRole("heading", { name: "Tambah Jabatan Organisasi" })
      .waitFor({ state: "visible", timeout: 10_000 });

    await this.page.getByLabel("Nama Jabatan").fill(input.jabatan);

    if (input.source === "member") {
      await this.page
        .getByRole("radio", { name: "Pilih dari Anggota Organisasi" })
        .check();
      const select = this.page.getByRole("combobox", {
        name: /Anggota Organisasi/,
      });
      await select.waitFor({ state: "visible", timeout: 10_000 });
      const opt = select
        .locator("option", { hasText: input.memberName ?? "" })
        .first();
      await opt.waitFor({ state: "attached", timeout: 10_000 });
      const val = await opt.getAttribute("value");
      assert.ok(val, `Expected option value for member ${input.memberName}`);
      await select.selectOption(val);
    } else if (input.source === "umat") {
      await this.page
        .getByRole("radio", { name: "Pilih dari Data Umat Paroki" })
        .check();
      if (input.umatName) {
        await this.page.getByLabel("Cari Nama Umat").fill(input.umatName);
      }
      const select = this.page.getByRole("combobox", {
        name: /Pilih Data Umat/,
      });
      const opt = select
        .locator("option", { hasText: input.umatName ?? "" })
        .first();
      await opt.waitFor({ state: "attached", timeout: 10_000 });
      const val = await opt.getAttribute("value");
      assert.ok(val, `Expected option value for umat ${input.umatName}`);
      await select.selectOption(val);
    } else {
      await this.page.getByLabel("Ketik Nama Manual / Kosongkan").check();
      if (input.manualNama !== undefined) {
        await this.page
          .getByLabel(
            "Nama Pengurus (opsional — kosongkan jika jabatan lowong)",
          )
          .fill(input.manualNama);
      }
    }

    if (input.urutan !== undefined) {
      await this.page.getByLabel("Urutan Tampil").fill(String(input.urutan));
    }
    await this.page.getByRole("button", { name: "Simpan Jabatan" }).click();
    await this.page
      .getByRole("heading", { name: "Struktur & Jabatan Organisasi" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectJabatanVisible(
    jabatan: string,
    holderName?: string,
    isChild = false,
  ): Promise<void> {
    const row = this.jabatanRow(jabatan);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    if (holderName) {
      await row
        .getByText(holderName, { exact: false })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (isChild) {
      await row
        .getByText("↳", { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectJabatanNotVisible(jabatan: string): Promise<void> {
    await this.jabatanRow(jabatan).waitFor({
      state: "detached",
      timeout: 10_000,
    });
  }

  async deleteJabatan(jabatan: string): Promise<void> {
    await this.jabatanRow(jabatan)
      .getByRole("button", { name: "Hapus" })
      .click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Hapus" })
      .click();
  }
}
