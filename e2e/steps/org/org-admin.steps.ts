import { Then, When } from "@cucumber/cucumber";
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

When(
  "I toggle org member {string} role via {string}",
  async function (this: DomusWorld, name: string, buttonLabel: string) {
    await this.pages.user.users.toggleOrgMemberRole(
      name,
      buttonLabel as "Jadikan Admin" | "Jadikan Anggota",
    );
  },
);

When(
  "I toggle org member {string} status via {string}",
  async function (this: DomusWorld, name: string, buttonLabel: string) {
    await this.pages.user.users.toggleOrgMemberStatus(
      name,
      buttonLabel as "Nonaktifkan" | "Aktifkan",
    );
  },
);

When(
  "I add org member with email {string} name {string} phone {string} role {string} position {string} period {string}",
  async function (
    this: DomusWorld,
    email: string,
    name: string,
    phone: string,
    roleLabel: string,
    position: string,
    periodName: string,
  ) {
    await this.pages.user.users.addOrgMember({
      email,
      name,
      phone,
      roleLabel: roleLabel as "Admin Organisasi" | "Anggota",
      position,
      periodName,
    });
  },
);

Then(
  "I see org member {string} metadata {string}",
  async function (this: DomusWorld, name: string, text: string) {
    await this.pages.user.users.expectOrgMemberMetadata(name, text);
  },
);

When(
  "I edit org member {string} position to {string}",
  async function (this: DomusWorld, name: string, position: string) {
    await this.pages.user.users.editOrgMemberPosition(name, { position });
  },
);

Then(
  "I see org member {string} status {string}",
  async function (this: DomusWorld, name: string, statusLabel: string) {
    await this.pages.user.users.expectOrgMemberStatus(
      name,
      statusLabel as "Aktif" | "Nonaktif",
    );
  },
);

When(
  "I remove org member {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.user.users.removeOrgMember(name);
  },
);

Then(
  "I do not see org member {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.user.users.expectOrgMemberNotVisible(name);
  },
);

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
  "I create org invite with role {string} position {string} max uses {int}",
  async function (
    this: DomusWorld,
    roleLabel: string,
    defaultPosition: string,
    maxUses: number,
  ) {
    await this.pages.org.detail.createInvite({
      roleLabel: roleLabel as "Anggota (member)" | "Admin Organisasi (admin)",
      defaultPosition,
      maxUses,
    });
  },
);

Then(
  "I see org invite with status {string} position {string} usage {string}",
  async function (
    this: DomusWorld,
    statusLabel: string,
    defaultPosition: string,
    maxUsesText: string,
  ) {
    await this.pages.org.detail.expectInviteVisible({
      statusLabel,
      defaultPosition,
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
