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
  "I see users filter buttons {string}",
  async function (this: DomusWorld, labels: string) {
    await this.pages.user.users.expectFilterButtonsVisible(splitCsv(labels));
  },
);

Then(
  "I do not see users directory controls {string}",
  async function (this: DomusWorld, labels: string) {
    await this.pages.user.users.expectHiddenControls(splitCsv(labels));
  },
);

Given(
  "a paroki member with email {string} name {string} exists in the active paroki",
  async function (this: DomusWorld, email: string, name: string) {
    assert.ok(this.activeParokiId, "Expected activeParokiId");
    const res = await this.client.post(
      `/ref/paroki/${this.activeParokiId}/members`,
      {
        email,
        name,
        role: "member",
      },
    );
    assert.ok(
      res.status === 201 || res.status === 409,
      `Expected 201 or 409 creating paroki member, got ${res.status}`,
    );
  },
);

When(
  "I demote paroki membership in the active paroki",
  async function (this: DomusWorld) {
    assert.ok(this.activeParokiName, "Expected activeParokiName");
    await this.pages.user.users.changeParokiMembershipRole(
      this.activeParokiName,
      "Anggota",
    );
  },
);

When(
  "I promote paroki membership in the active paroki",
  async function (this: DomusWorld) {
    assert.ok(this.activeParokiName, "Expected activeParokiName");
    await this.pages.user.users.changeParokiMembershipRole(
      this.activeParokiName,
      "Admin Paroki",
    );
  },
);

Then(
  "I see active paroki membership with role {string}",
  async function (this: DomusWorld, roleLabel: string) {
    assert.ok(this.activeParokiName, "Expected activeParokiName");
    await this.pages.user.users.expectParokiMembershipVisible(
      this.activeParokiName,
      roleLabel as "Admin Paroki" | "Anggota",
    );
  },
);

When(
  "I remove paroki membership in the active paroki",
  async function (this: DomusWorld) {
    assert.ok(this.activeParokiName, "Expected activeParokiName");
    await this.pages.user.users.removeParokiMembership(this.activeParokiName);
  },
);

Then(
  "I see users error {string}",
  async function (this: DomusWorld, message: string) {
    await this.pages.user.users.expectErrorAlert(message);
  },
);

Given(
  "a wilayah with code {string} and name {string} exists in the active paroki",
  async function (this: DomusWorld, code: string, name: string) {
    const res = await this.client.post("/ref/wilayah", {
      jenis: "wilayah",
      kode: code,
      nama: name,
    });
    assert.ok(
      res.status === 201 || res.status === 409,
      `Expected 201 or 409 creating wilayah, got ${res.status}`,
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
  "I create organization with kind {string} name {string} code {string} wilayah {string}",
  async function (
    this: DomusWorld,
    kindLabel: string,
    name: string,
    code: string,
    wilayahLabel: string,
  ) {
    await this.pages.org.orgs.createOrg({
      kindLabel,
      name,
      code,
      wilayahLabel,
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
  "I add user to org {string} with role {string}",
  async function (this: DomusWorld, orgName: string, roleLabel: string) {
    await this.pages.user.users.addToOrg({
      orgName,
      roleLabel: roleLabel as "Admin Organisasi" | "Anggota",
    });
  },
);

Then(
  "I see user org membership {string} with role {string}",
  async function (this: DomusWorld, orgName: string, roleLabel: string) {
    await this.pages.user.users.expectOrgMembershipVisible(
      orgName,
      roleLabel as "Admin Organisasi" | "Anggota",
    );
  },
);
