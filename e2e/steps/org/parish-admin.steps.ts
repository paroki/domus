import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import type { DomusWorld } from "../../support/world";

function splitCsv(raw: string): string[] {
  return raw
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

When(
  "I open home launcher tile {string}",
  async function (this: DomusWorld, label: string) {
    await this.pages.shell.home.openTile(label);
  },
);

Then(
  "I see users directory scopes {string}",
  async function (this: DomusWorld, labels: string) {
    await this.pages.user.users.expectVisibleScopes(splitCsv(labels));
  },
);

Then(
  "I do not see users directory scopes {string}",
  async function (this: DomusWorld, labels: string) {
    await this.pages.user.users.expectHiddenScopes(splitCsv(labels));
  },
);

When(
  "I change parish member {string} role to {string}",
  async function (this: DomusWorld, name: string, roleLabel: string) {
    await this.pages.user.users.changeParishMemberRole(
      name,
      roleLabel as "Admin Paroki" | "Anggota",
    );
  },
);

Then(
  "I see users error {string}",
  async function (this: DomusWorld, message: string) {
    await this.pages.user.users.expectErrorAlert(message);
  },
);

When(
  "I remove parish member {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.user.users.removeParishMember(name);
  },
);

Then(
  "I do not see parish member {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.user.users.expectParishMemberNotVisible(name);
  },
);

Given(
  "a zone with code {string} and name {string} exists in the active parish",
  async function (this: DomusWorld, code: string, name: string) {
    const res = await this.client.post("/ref/zones", {
      type: "zone",
      code,
      name,
    });
    assert.ok(
      res.status === 201 || res.status === 409,
      `Expected 201 or 409 creating zone, got ${res.status}`,
    );
  },
);

Then("I see the orgs directory page", async function (this: DomusWorld) {
  await this.pages.org.orgs.expectLoaded();
});

When(
  "I create organization with kind {string} name {string} code {string}",
  async function (
    this: DomusWorld,
    kindLabel: string,
    name: string,
    code: string,
  ) {
    await this.pages.org.orgs.createOrg({ kindLabel, name, code });
  },
);

When(
  "I create organization with kind {string} name {string} code {string} parent {string}",
  async function (
    this: DomusWorld,
    kindLabel: string,
    name: string,
    code: string,
    parentLabel: string,
  ) {
    await this.pages.org.orgs.createOrg({
      kindLabel,
      name,
      code,
      parentLabel,
    });
  },
);

When(
  "I create organization with kind {string} name {string} code {string} zone {string}",
  async function (
    this: DomusWorld,
    kindLabel: string,
    name: string,
    code: string,
    zoneLabel: string,
  ) {
    await this.pages.org.orgs.createOrg({
      kindLabel,
      name,
      code,
      zoneLabel,
    });
  },
);

When(
  "I filter organizations by kind {string}",
  async function (this: DomusWorld, kindLabel: string) {
    await this.pages.org.orgs.filterByKind(kindLabel);
  },
);

When(
  "I search organizations for {string}",
  async function (this: DomusWorld, query: string) {
    await this.pages.org.orgs.search(query);
  },
);

Then(
  "I see organization card {string} with code {string} and kind {string}",
  async function (
    this: DomusWorld,
    name: string,
    code: string,
    kindBadge: string,
  ) {
    await this.pages.org.orgs.expectOrgCardVisible(name, { code, kindBadge });
  },
);

When(
  "I select organization {string}",
  async function (this: DomusWorld, orgName: string) {
    await this.pages.user.users.selectOrganization(orgName);
  },
);

When("I self-assign as Organization Admin", async function (this: DomusWorld) {
  await this.pages.user.users.selfAssignOrgAdmin();
});

When(
  "I add org member with email {string} name {string} and role {string}",
  async function (
    this: DomusWorld,
    email: string,
    name: string,
    roleLabel: string,
  ) {
    await this.pages.user.users.addOrgMember({
      email,
      name,
      roleLabel: roleLabel as "Admin Organisasi" | "Anggota",
    });
  },
);

Then(
  "I see org member {string} with role {string}",
  async function (this: DomusWorld, name: string, roleLabel: string) {
    await this.pages.user.users.expectOrgMemberVisible(
      name,
      roleLabel as "Admin Organisasi" | "Anggota",
    );
  },
);
