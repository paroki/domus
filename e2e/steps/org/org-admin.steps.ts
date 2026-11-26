import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import { loginAsSuperAdmin, loginAsUser } from "../../support/auth";
import { TestClient } from "../../support/test-client";
import type { DomusWorld } from "../../support/world";

function splitCsv(raw: string): string[] {
  return raw
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

Then(
  "I see launcher tiles {string}",
  async function (this: DomusWorld, labels: string) {
    await this.pages.shell.home.expectVisibleTiles(splitCsv(labels));
  },
);

Then(
  "I do not see launcher tiles {string}",
  async function (this: DomusWorld, labels: string) {
    await this.pages.shell.home.expectHiddenTiles(splitCsv(labels));
  },
);

Then("I see empty settings message", async function (this: DomusWorld) {
  await this.pages.ref.settings.expectEmptySettings();
});

Then(
  "I do not see the create organization button",
  async function (this: DomusWorld) {
    await this.pages.org.orgs.expectCreateButtonVisible(false);
  },
);

Given(
  "an org member with email {string} name {string} exists in the active org",
  async function (this: DomusWorld, email: string, name: string) {
    assert.ok(this.activeOrgId, "Expected activeOrgId");
    const res = await this.client.post(`/orgs/${this.activeOrgId}/members`, {
      email,
      name,
      role: "member",
    });
    assert.ok(
      res.status === 201 || res.status === 409,
      `Expected 201 or 409 creating org member, got ${res.status}`,
    );
  },
);

When(
  "I toggle active org membership role via {string}",
  async function (this: DomusWorld, buttonLabel: string) {
    assert.ok(this.activeOrgName, "Expected activeOrgName");
    await this.pages.user.users.toggleOrgMembershipRole(
      this.activeOrgName,
      buttonLabel as "Jadikan Admin" | "Jadikan Anggota",
    );
  },
);

When(
  "I toggle active org membership status via {string}",
  async function (this: DomusWorld, buttonLabel: string) {
    assert.ok(this.activeOrgName, "Expected activeOrgName");
    await this.pages.user.users.toggleOrgMembershipStatus(
      this.activeOrgName,
      buttonLabel as "Nonaktifkan" | "Aktifkan",
    );
  },
);

Then(
  "I see active org membership with role {string}",
  async function (this: DomusWorld, roleLabel: string) {
    assert.ok(this.activeOrgName, "Expected activeOrgName");
    await this.pages.user.users.expectOrgMembershipVisible(
      this.activeOrgName,
      roleLabel as "Admin Organisasi" | "Anggota",
    );
  },
);

Then(
  "I see active org membership metadata {string}",
  async function (this: DomusWorld, text: string) {
    assert.ok(this.activeOrgName, "Expected activeOrgName");
    await this.pages.user.users.expectOrgMembershipMetadata(
      this.activeOrgName,
      text,
    );
  },
);

Then(
  "I see active org membership status {string}",
  async function (this: DomusWorld, statusLabel: string) {
    assert.ok(this.activeOrgName, "Expected activeOrgName");
    await this.pages.user.users.expectOrgMembershipStatus(
      this.activeOrgName,
      statusLabel as "Aktif" | "Nonaktif",
    );
  },
);

When("I remove active org membership", async function (this: DomusWorld) {
  assert.ok(this.activeOrgName, "Expected activeOrgName");
  await this.pages.user.users.removeOrgMembership(this.activeOrgName);
});

When(
  "I open organization detail {string}",
  async function (this: DomusWorld, orgName: string) {
    await this.pages.org.orgs.openOrgDetail(orgName);
  },
);

Then(
  "I see organization detail {string}",
  async function (this: DomusWorld, orgName: string) {
    await this.pages.org.detail.expectLoaded(orgName);
  },
);

Then(
  "I see org edit button and no delete button",
  async function (this: DomusWorld) {
    await this.pages.org.detail.expectEditAndDeleteVisibility({
      canEdit: true,
      canDelete: false,
    });
  },
);

Then(
  "I see read-only org roster with member {string}",
  async function (this: DomusWorld, memberName: string) {
    await this.pages.org.detail.expectRosterReadOnly(memberName);
  },
);

When(
  "I update organization description to {string}",
  async function (this: DomusWorld, description: string) {
    await this.pages.org.detail.updateProfile({ description });
  },
);

Then(
  "I see organization description {string}",
  async function (this: DomusWorld, description: string) {
    await this.pages.org.detail.expectDescriptionVisible(description);
  },
);

When(
  "I create org invite with role {string} max uses {int}",
  async function (this: DomusWorld, roleLabel: string, maxUses: number) {
    await this.pages.org.detail.createInvite({
      roleLabel: roleLabel as "Anggota (member)" | "Admin Organisasi (admin)",
      maxUses,
    });
  },
);

Then(
  "I see org invite with status {string} usage {string}",
  async function (this: DomusWorld, statusLabel: string, maxUsesText: string) {
    await this.pages.org.detail.expectInviteVisible({
      statusLabel,
      maxUsesText,
    });
  },
);

When("I revoke the org invite", async function (this: DomusWorld) {
  await this.pages.org.detail.revokeInvite();
});

Then(
  "I see org invite with status {string}",
  async function (this: DomusWorld, statusLabel: string) {
    await this.pages.org.detail.expectInviteVisible({ statusLabel });
  },
);

Then(
  "I see org error {string}",
  async function (this: DomusWorld, message: string) {
    await this.pages.org.detail.expectErrorAlert(message);
  },
);

Then(
  "I see org period {string} with status {string} and current badge",
  async function (this: DomusWorld, name: string, status: string) {
    await this.pages.org.detail.expectPeriodeVisible(name, {
      status: status as "Terbuka" | "Ditutup",
      isCurrent: true,
    });
  },
);

Then(
  "I see org period {string} with status {string}",
  async function (this: DomusWorld, name: string, status: string) {
    await this.pages.org.detail.expectPeriodeVisible(name, {
      status: status as "Terbuka" | "Ditutup",
    });
  },
);

Then(
  "I see org period {string} without current badge",
  async function (this: DomusWorld, name: string) {
    await this.pages.org.detail.expectPeriodeVisible(name, {
      isCurrent: false,
    });
  },
);

Then(
  "I do not see org period {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.org.detail.expectPeriodeNotVisible(name);
  },
);

When(
  "I open the new org period form and submit {string} from {string} to {string}",
  async function (
    this: DomusWorld,
    nama: string,
    tanggalMulai: string,
    tanggalSelesai: string,
  ) {
    await this.pages.org.detail.openNewPeriodeForm();
    await this.pages.org.detail.fillAndSubmitPeriodeForm({
      nama,
      tanggalMulai,
      tanggalSelesai,
    });
  },
);

When(
  "I resubmit the org period form with name {string} from {string} to {string}",
  async function (
    this: DomusWorld,
    nama: string,
    tanggalMulai: string,
    tanggalSelesai: string,
  ) {
    await this.pages.org.detail.fillAndSubmitPeriodeForm({
      nama,
      kode: nama.toLowerCase().replace(/[^a-z0-9]+/g, "-"),
      tanggalMulai,
      tanggalSelesai,
    });
  },
);

When(
  "I make org period {string} active",
  async function (this: DomusWorld, name: string) {
    await this.pages.org.detail.makePeriodeCurrent(name);
  },
);

When(
  "I toggle org period {string} via {string}",
  async function (this: DomusWorld, name: string, buttonLabel: string) {
    await this.pages.org.detail.togglePeriodeStatus(
      name,
      buttonLabel as "Tutup" | "Buka",
    );
  },
);

When(
  "I delete org period {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.org.detail.deletePeriode(name);
  },
);

Given(
  "a parishioner named {string} exists in the active parish",
  async function (this: DomusWorld, nama: string) {
    assert.ok(this.activeParokiId, "Expected activeParokiId");
    const sa = await loginAsSuperAdmin();
    const saClient = new TestClient();
    saClient.setToken(sa.rawToken);
    const umatAdminEmail = "umatadmin.e2e@domus.test";
    const addRes = await saClient.post(
      `/ref/paroki/${this.activeParokiId}/members`,
      {
        email: umatAdminEmail,
        name: "Umat Admin E2E",
        role: "member",
        umatAdmin: true,
      },
    );
    assert.ok(
      addRes.status === 201 || addRes.status === 409,
      `Expected 201 or 409 adding umat admin, got ${addRes.status}`,
    );

    const uaSession = await loginAsUser({
      email: umatAdminEmail,
      name: "Umat Admin E2E",
      onboarded: true,
      activeParokiId: this.activeParokiId,
    });
    const uaClient = new TestClient();
    uaClient.setToken(uaSession.rawToken);
    uaClient.setParokiId(this.activeParokiId);
    const res = await uaClient.post("/umat", {
      nama,
      jenisKelamin: "pria",
    });
    assert.equal(
      res.status,
      201,
      `Expected 201 creating umat, got ${res.status}: ${JSON.stringify(res.data)}`,
    );
  },
);

When(
  "I create root org jabatan {string} assigned to org member {string}",
  async function (this: DomusWorld, jabatan: string, memberName: string) {
    await this.pages.org.detail.createJabatan({
      jabatan,
      source: "member",
      memberName,
    });
  },
);

When(
  "I create root org jabatan {string} assigned to umat {string}",
  async function (this: DomusWorld, jabatan: string, umatName: string) {
    await this.pages.org.detail.createJabatan({
      jabatan,
      source: "umat",
      umatName,
    });
  },
);

When(
  "I create child org jabatan {string} under {string} assigned to umat {string}",
  async function (
    this: DomusWorld,
    jabatan: string,
    parentJabatan: string,
    umatName: string,
  ) {
    await this.pages.org.detail.createJabatan({
      parentJabatan,
      jabatan,
      source: "umat",
      umatName,
    });
  },
);

When(
  "I create child org jabatan {string} under {string} with manual name {string}",
  async function (
    this: DomusWorld,
    jabatan: string,
    parentJabatan: string,
    manualNama: string,
  ) {
    await this.pages.org.detail.createJabatan({
      parentJabatan,
      jabatan,
      source: "manual",
      manualNama,
    });
  },
);

Then(
  "I see org jabatan {string} held by {string}",
  async function (this: DomusWorld, jabatan: string, holderName: string) {
    await this.pages.org.detail.expectJabatanVisible(
      jabatan,
      holderName,
      false,
    );
  },
);

Then(
  "I see child org jabatan {string} held by {string}",
  async function (this: DomusWorld, jabatan: string, holderName: string) {
    await this.pages.org.detail.expectJabatanVisible(jabatan, holderName, true);
  },
);

When(
  "I delete org jabatan {string}",
  async function (this: DomusWorld, jabatan: string) {
    await this.pages.org.detail.deleteJabatan(jabatan);
  },
);

Then(
  "I do not see org jabatan {string}",
  async function (this: DomusWorld, jabatan: string) {
    await this.pages.org.detail.expectJabatanNotVisible(jabatan);
  },
);
