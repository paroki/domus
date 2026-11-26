import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export type UserDetailTabLabel = "Profil" | "Paroki" | "Organisasi";

export class UsersPage extends BasePage {
  private rowFor(nameOrEmail: string) {
    return this.page
      .locator("div.divide-y > a")
      .filter({ hasText: nameOrEmail });
  }

  private parokiMembershipRow(parokiName: string) {
    return this.page
      .locator("#user-tabpanel-paroki div.divide-y > div")
      .filter({ hasText: parokiName });
  }

  private orgMembershipRow(orgName: string) {
    return this.page
      .locator("#user-tabpanel-org div.divide-y > div")
      .filter({ hasText: orgName });
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

  async expectFilterButtonsVisible(labels: string[]): Promise<void> {
    for (const label of labels) {
      await this.page
        .getByRole("button", { name: label, exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async expectHiddenControls(labels: string[]): Promise<void> {
    for (const label of labels) {
      const tabCount = await this.page
        .getByRole("tab", { name: label, exact: true })
        .count();
      const btnCount = await this.page
        .getByRole("button", { name: label, exact: true })
        .count();
      assert.equal(
        tabCount + btnCount,
        0,
        `Expected control "${label}" to be hidden`,
      );
    }
  }

  async expectErrorAlert(message: string): Promise<void> {
    await this.page
      .getByRole("alert")
      .filter({ hasText: message })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async search(query: string): Promise<void> {
    await this.page
      .getByRole("searchbox", { name: "Cari nama atau email…" })
      .fill(query);
  }

  async filterByStatus(statusLabel: string): Promise<void> {
    await this.page
      .getByRole("button", { name: statusLabel, exact: true })
      .click();
  }

  async openUserDetail(nameOrEmail: string): Promise<void> {
    await this.rowFor(nameOrEmail).click();
    await this.page
      .getByRole("link", { name: "Kembali ke Direktori Pengguna" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async backToDirectory(): Promise<void> {
    await this.page
      .getByRole("link", { name: "Kembali ke Direktori Pengguna" })
      .click();
    await this.expectLoaded();
  }

  async selectDetailTab(tabLabel: UserDetailTabLabel): Promise<void> {
    await this.page
      .getByRole("tab", { name: new RegExp(`^${tabLabel}`) })
      .click();
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

  async expectDetailSuperAdminBadge(visible: boolean): Promise<void> {
    const badge = this.page
      .locator("div.flex.flex-wrap.items-center.gap-2")
      .first()
      .getByText("Super Admin", { exact: true });
    if (visible) {
      await badge.waitFor({ state: "visible", timeout: 10_000 });
    } else {
      await badge.waitFor({ state: "detached", timeout: 10_000 });
    }
  }

  async promoteToSuperAdmin(): Promise<void> {
    await this.page
      .getByRole("button", { name: "Jadikan Super Admin" })
      .click();
    await this.expectDetailSuperAdminBadge(true);
  }

  async demoteFromSuperAdmin(): Promise<void> {
    await this.page.getByRole("button", { name: "Cabut Super Admin" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Cabut Super Admin" })
      .click();
    await this.expectDetailSuperAdminBadge(false);
  }

  async disableUser(): Promise<void> {
    await this.page.getByRole("button", { name: "Nonaktifkan" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Nonaktifkan" })
      .click();
    await this.page
      .getByRole("button", { name: "Aktifkan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async enableUser(): Promise<void> {
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/users/") &&
          r.request().method() === "PATCH" &&
          r.status() === 200,
      ),
      this.page.getByRole("button", { name: "Aktifkan" }).click(),
    ]);
    await this.page
      .getByRole("button", { name: "Nonaktifkan" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteUser(): Promise<void> {
    await this.page.getByRole("button", { name: "Hapus", exact: true }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Hapus", exact: true })
      .click();
    await this.expectLoaded();
  }

  async expectUserNotVisible(nameOrEmail: string): Promise<void> {
    await this.rowFor(nameOrEmail).waitFor({
      state: "detached",
      timeout: 10_000,
    });
  }

  async addToParoki(
    parokiName: string,
    roleLabel: "Admin Paroki" | "Anggota",
  ): Promise<void> {
    const panel = this.page.locator("#user-tabpanel-paroki");
    const select = panel.getByLabel("Pilih paroki");
    await select
      .locator("option", { hasText: parokiName })
      .waitFor({ state: "attached", timeout: 10_000 });
    await select.selectOption({ label: parokiName });
    await panel.getByLabel("Peran").selectOption({ label: roleLabel });
    await panel.getByRole("button", { name: "Tambahkan ke Paroki" }).click();
  }

  async expectParokiMembershipVisible(
    parokiName: string,
    roleLabel: "Admin Paroki" | "Anggota",
  ): Promise<void> {
    const row = this.parokiMembershipRow(parokiName);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row
      .locator("span", { hasText: new RegExp(`^${roleLabel}$`) })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async changeParokiMembershipRole(
    parokiName: string,
    roleLabel: "Admin Paroki" | "Anggota",
  ): Promise<void> {
    const row = this.parokiMembershipRow(parokiName);
    if (roleLabel === "Anggota") {
      await row.getByRole("button", { name: "Turunkan ke Anggota" }).click();
      await this.page
        .getByRole("dialog")
        .getByRole("button", { name: "Turunkan ke Anggota" })
        .click();
      await this.page
        .getByRole("dialog")
        .waitFor({ state: "detached", timeout: 10_000 });
    } else {
      await row.getByRole("button", { name: "Jadikan Admin Paroki" }).click();
    }
  }

  async removeParokiMembership(parokiName: string): Promise<void> {
    await this.page
      .getByRole("button", { name: `Hapus ${parokiName}` })
      .click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Hapus", exact: true })
      .click();
    await this.parokiMembershipRow(parokiName).waitFor({
      state: "detached",
      timeout: 10_000,
    });
  }

  async expectParokiMembershipNotVisible(parokiName: string): Promise<void> {
    await this.parokiMembershipRow(parokiName).waitFor({
      state: "detached",
      timeout: 10_000,
    });
  }

  async grantUmatAdmin(parokiName: string): Promise<void> {
    const row = this.parokiMembershipRow(parokiName);
    await row.getByRole("button", { name: "Jadikan Admin Data Umat" }).click();
    await this.expectUmatAdminBadge(parokiName, true);
  }

  async revokeUmatAdmin(parokiName: string): Promise<void> {
    const row = this.parokiMembershipRow(parokiName);
    await row.getByRole("button", { name: "Cabut Admin Data Umat" }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Cabut Admin Data Umat" })
      .click();
    await this.expectUmatAdminBadge(parokiName, false);
  }

  async expectUmatAdminBadge(
    parokiName: string,
    visible: boolean,
  ): Promise<void> {
    const badge = this.parokiMembershipRow(parokiName).locator("span", {
      hasText: /^Admin Data Umat$/,
    });
    if (visible) {
      await badge.waitFor({ state: "visible", timeout: 10_000 });
    } else {
      await badge.waitFor({ state: "detached", timeout: 10_000 });
    }
  }

  async addToOrg(input: {
    orgName: string;
    roleLabel: "Admin Organisasi" | "Anggota";
    position?: string;
    periodName?: string;
  }): Promise<void> {
    const panel = this.page.locator("#user-tabpanel-org");
    const select = panel.getByLabel("Pilih organisasi");
    await select
      .locator("option", { hasText: input.orgName })
      .waitFor({ state: "attached", timeout: 10_000 });
    await select.selectOption({ label: input.orgName });
    const roleSelect = panel.getByLabel("Peran");
    if (await roleSelect.isEnabled()) {
      await roleSelect.selectOption({ label: input.roleLabel });
    }
    if (input.position) {
      await panel.getByLabel("Jabatan").fill(input.position);
    }
    if (input.periodName) {
      await panel
        .getByLabel("Periode")
        .selectOption({ label: input.periodName });
    }
    await panel
      .getByRole("button", { name: "Tambahkan ke Organisasi" })
      .click();
  }

  async expectOrgMembershipVisible(
    orgName: string,
    roleLabel: "Admin Organisasi" | "Anggota",
  ): Promise<void> {
    const row = this.orgMembershipRow(orgName);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row
      .locator("span", { hasText: new RegExp(`^${roleLabel}$`) })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectOrgMembershipMetadata(
    orgName: string,
    text: string,
  ): Promise<void> {
    const row = this.orgMembershipRow(orgName);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row
      .getByText(text, { exact: false })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectOrgMembershipStatus(
    orgName: string,
    statusLabel: "Aktif" | "Nonaktif",
  ): Promise<void> {
    const row = this.orgMembershipRow(orgName);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    await row
      .locator("span", { hasText: new RegExp(`^${statusLabel}$`) })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectOrgMembershipNotVisible(orgName: string): Promise<void> {
    await this.orgMembershipRow(orgName).waitFor({
      state: "detached",
      timeout: 10_000,
    });
  }

  async editOrgMembershipPosition(
    orgName: string,
    position: string,
  ): Promise<void> {
    const row = this.orgMembershipRow(orgName);
    await row.getByRole("button", { name: "Ubah Jabatan" }).click();
    await row.getByLabel("Jabatan").fill(position);
    await row.getByRole("button", { name: "Simpan" }).click();
  }

  async toggleOrgMembershipRole(
    orgName: string,
    targetButtonLabel: "Jadikan Admin" | "Jadikan Anggota",
  ): Promise<void> {
    const row = this.orgMembershipRow(orgName);
    await row.getByRole("button", { name: targetButtonLabel }).click();
  }

  async toggleOrgMembershipStatus(
    orgName: string,
    targetButtonLabel: "Nonaktifkan" | "Aktifkan",
  ): Promise<void> {
    const row = this.orgMembershipRow(orgName);
    await row.getByRole("button", { name: targetButtonLabel }).click();
  }

  async removeOrgMembership(orgName: string): Promise<void> {
    await this.page.getByRole("button", { name: `Hapus ${orgName}` }).click();
    await this.page
      .getByRole("dialog")
      .getByRole("button", { name: "Hapus", exact: true })
      .click();
    await this.expectOrgMembershipNotVisible(orgName);
  }
}
