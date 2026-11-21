import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export type UserScopeLabel =
  | "Pengguna Global"
  | "Admin & Anggota Paroki"
  | "Pengurus & Anggota Organisasi";

export class UsersPage extends BasePage {
  private rowFor(nameOrEmail: string) {
    return this.page
      .locator("div.divide-y > div")
      .filter({ hasText: nameOrEmail });
  }

  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Direktori Pengguna" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectForbidden(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Akses Ditolak" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByText(
        "Hanya Super Admin, Admin Paroki, atau Admin Organisasi yang dapat mengelola direktori pengguna.",
      )
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async selectScope(scopeLabel: UserScopeLabel): Promise<void> {
    await this.page.getByRole("tab", { name: scopeLabel }).click();
  }

  async createGlobalUser(input: {
    email: string;
    name: string;
    phone?: string;
    isSuperAdmin?: boolean;
  }): Promise<void> {
    await this.page.getByRole("button", { name: "Tambah Pengguna" }).click();
    await this.page.getByLabel("Alamat email").fill(input.email);
    await this.page.getByLabel("Nama lengkap").fill(input.name);
    if (input.phone) {
      await this.page.getByLabel("Nomor telepon (opsional)").fill(input.phone);
    }
    if (input.isSuperAdmin) {
      await this.page.getByLabel("Jadikan Super Admin").check();
    }
    await this.page.getByRole("button", { name: "Buat Pengguna" }).click();
    await this.page
      .getByRole("button", { name: "Buat Pengguna" })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async search(query: string): Promise<void> {
    await this.page
      .getByRole("searchbox", { name: "Cari nama atau email…" })
      .fill(query);
  }

  async filterByStatus(statusLabel: string): Promise<void> {
    await this.page
      .getByLabel("Semua status")
      .selectOption({ label: statusLabel });
  }

  async expectUserVisible(
    name: string,
    email: string,
    options?: {
      status?: "Aktif" | "Nonaktif";
      hasSuperAdminBadge?: boolean;
    },
  ): Promise<void> {
    const row = this.rowFor(email);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row.getByText(name, { exact: true }).waitFor({ state: "visible" });
    if (options?.status) {
      await row
        .getByText(options.status, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.hasSuperAdminBadge === true) {
      await row
        .getByText("Super Admin", { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.hasSuperAdminBadge === false) {
      const count = await row.getByText("Super Admin", { exact: true }).count();
      assert.equal(count, 0, `Expected no Super Admin badge for ${email}`);
    }
  }

  async promoteToSuperAdmin(nameOrEmail: string): Promise<void> {
    const row = this.rowFor(nameOrEmail);
    await row.getByRole("button", { name: "Jadikan Super Admin" }).click();
    await row
      .getByText("Super Admin", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async demoteFromSuperAdmin(nameOrEmail: string): Promise<void> {
    const row = this.rowFor(nameOrEmail);
    await row.getByRole("button", { name: "Cabut Super Admin" }).click();
    await row
      .getByText("Super Admin", { exact: true })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async disableUser(nameOrEmail: string): Promise<void> {
    const row = this.rowFor(nameOrEmail);
    await row.getByRole("button", { name: "Nonaktifkan" }).click();
    await row
      .getByText("Nonaktif", { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async enableUser(nameOrEmail: string): Promise<void> {
    const row = this.rowFor(nameOrEmail);
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/users/") &&
          r.request().method() === "PATCH" &&
          r.status() === 200,
      ),
      row.getByRole("button", { name: "Aktifkan" }).click(),
    ]);
  }

  async deleteUser(name: string): Promise<void> {
    this.page.once("dialog", (dialog) => dialog.accept());
    await this.page.getByRole("button", { name: `Hapus ${name}` }).click();
  }

  async expectUserNotVisible(nameOrEmail: string): Promise<void> {
    await this.rowFor(nameOrEmail).waitFor({
      state: "detached",
      timeout: 10_000,
    });
  }

  async selectParish(parishName: string): Promise<void> {
    const select = this.page.getByLabel("Pilih paroki");
    await select
      .locator("option", { hasText: parishName })
      .waitFor({ state: "attached", timeout: 10_000 });
    await select.selectOption({ label: parishName });
  }

  async addParishMember(input: {
    email: string;
    name: string;
    phone?: string;
    roleLabel: "Admin Paroki" | "Anggota";
  }): Promise<void> {
    await this.page.getByRole("button", { name: "Tambah Pengguna" }).click();
    await this.page.getByLabel("Alamat email").fill(input.email);
    await this.page.getByLabel("Nama lengkap").fill(input.name);
    if (input.phone) {
      await this.page.getByLabel("Nomor telepon (opsional)").fill(input.phone);
    }
    await this.page
      .locator("form")
      .getByLabel("Peran")
      .selectOption({ label: input.roleLabel });
    await this.page
      .getByRole("button", { name: "Simpan Anggota Paroki" })
      .click();
    await this.page
      .getByRole("button", { name: "Simpan Anggota Paroki" })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async expectParishMemberVisible(
    name: string,
    roleLabel: "Admin Paroki" | "Anggota",
  ): Promise<void> {
    const row = this.rowFor(name);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row
      .locator("span", { hasText: new RegExp(`^${roleLabel}$`) })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async changeParishMemberRole(
    name: string,
    roleLabel: "Admin Paroki" | "Anggota",
  ): Promise<void> {
    await this.page
      .getByLabel(`Peran ${name}`)
      .selectOption({ label: roleLabel });
  }

  async removeParishMember(name: string): Promise<void> {
    this.page.once("dialog", (dialog) => dialog.accept());
    await this.page.getByRole("button", { name: `Hapus ${name}` }).click();
  }

  async selectOrganization(orgName: string): Promise<void> {
    const select = this.page.getByLabel("Pilih organisasi");
    await select
      .locator("option", { hasText: orgName })
      .waitFor({ state: "attached", timeout: 10_000 });
    await select.selectOption({ label: orgName });
  }

  async selfAssignOrgAdmin(): Promise<void> {
    await this.page
      .getByRole("button", { name: "Jadikan Saya Admin Organisasi" })
      .click();
  }

  async addOrgMember(input: {
    email: string;
    name?: string;
    phone?: string;
    roleLabel: "Admin Organisasi" | "Anggota";
    position?: string;
    periodName?: string;
  }): Promise<void> {
    await this.page.getByRole("button", { name: "Tambah Pengguna" }).click();
    await this.page.getByLabel("Alamat email").fill(input.email);
    if (input.name) {
      await this.page.getByLabel("Nama lengkap").fill(input.name);
    }
    if (input.phone) {
      await this.page.getByLabel("Nomor telepon (opsional)").fill(input.phone);
    }
    const roleSelect = this.page.locator("form").getByLabel("Peran");
    if (await roleSelect.isEnabled()) {
      await roleSelect.selectOption({ label: input.roleLabel });
    }
    if (input.position) {
      await this.page
        .locator("form")
        .getByLabel("Jabatan")
        .fill(input.position);
    }
    if (input.periodName) {
      await this.page
        .locator("form")
        .getByLabel("Periode")
        .selectOption({ label: input.periodName });
    }
    await this.page
      .getByRole("button", { name: "Simpan Anggota Organisasi" })
      .click();
    await this.page
      .getByRole("button", { name: "Simpan Anggota Organisasi" })
      .waitFor({ state: "detached", timeout: 10_000 });
  }

  async editOrgMemberPosition(
    name: string,
    input: { position: string; periodName?: string },
  ): Promise<void> {
    const row = this.rowFor(name);
    await row.getByRole("button", { name: "Ubah Jabatan" }).click();
    await row.getByLabel("Jabatan").fill(input.position);
    if (input.periodName) {
      await row.getByLabel("Periode").selectOption({ label: input.periodName });
    }
    await row.getByRole("button", { name: "Simpan" }).click();
  }

  async toggleOrgMemberRole(
    name: string,
    targetButtonLabel: "Jadikan Admin" | "Jadikan Anggota",
  ): Promise<void> {
    const row = this.rowFor(name);
    await row.getByRole("button", { name: targetButtonLabel }).click();
  }

  async toggleOrgMemberStatus(
    name: string,
    targetButtonLabel: "Nonaktifkan" | "Aktifkan",
  ): Promise<void> {
    const row = this.rowFor(name);
    await row.getByRole("button", { name: targetButtonLabel }).click();
  }

  async removeOrgMember(name: string): Promise<void> {
    this.page.once("dialog", (dialog) => dialog.accept());
    await this.page.getByRole("button", { name: `Hapus ${name}` }).click();
  }
}
