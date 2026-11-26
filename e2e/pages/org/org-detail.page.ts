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
}
