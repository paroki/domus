import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
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
