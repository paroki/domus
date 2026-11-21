import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import type { UserScopeLabel } from "../../pages/user/users.page";
import type { DomusWorld } from "../../support/world";

Then("I see the users forbidden banner", async function (this: DomusWorld) {
  await this.pages.user.users.expectForbidden();
});

Then("I see the users directory page", async function (this: DomusWorld) {
  await this.pages.user.users.expectLoaded();
});

When(
  "I create a global user with email {string} name {string} phone {string}",
  async function (
    this: DomusWorld,
    email: string,
    name: string,
    phone: string,
  ) {
    await this.pages.user.users.createGlobalUser({ email, name, phone });
  },
);

When(
  "I search users for {string}",
  async function (this: DomusWorld, query: string) {
    await this.pages.user.users.search(query);
  },
);

Then(
  "I see global user {string} with email {string} and status {string}",
  async function (
    this: DomusWorld,
    name: string,
    email: string,
    status: string,
  ) {
    await this.pages.user.users.expectUserVisible(name, email, {
      status: status as "Aktif" | "Nonaktif",
    });
  },
);

When(
  "I promote user {string} to Super Admin",
  async function (this: DomusWorld, email: string) {
    await this.pages.user.users.promoteToSuperAdmin(email);
  },
);

Then(
  "I see global user {string} with Super Admin badge",
  async function (this: DomusWorld, name: string) {
    await this.pages.user.users.expectUserVisible(name, name, {
      hasSuperAdminBadge: true,
    });
  },
);

When(
  "I demote user {string} from Super Admin",
  async function (this: DomusWorld, email: string) {
    await this.pages.user.users.demoteFromSuperAdmin(email);
  },
);

Then(
  "I see global user {string} without Super Admin badge",
  async function (this: DomusWorld, name: string) {
    await this.pages.user.users.expectUserVisible(name, name, {
      hasSuperAdminBadge: false,
    });
  },
);

When(
  "I disable user {string}",
  async function (this: DomusWorld, email: string) {
    await this.pages.user.users.disableUser(email);
  },
);

When(
  "I enable user {string}",
  async function (this: DomusWorld, email: string) {
    await this.pages.user.users.enableUser(email);
  },
);

When(
  "I filter users by status {string}",
  async function (this: DomusWorld, statusLabel: string) {
    await this.pages.user.users.filterByStatus(statusLabel);
  },
);

When(
  "I delete global user {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.user.users.deleteUser(name);
  },
);

Then(
  "I do not see user {string}",
  async function (this: DomusWorld, email: string) {
    await this.pages.user.users.expectUserNotVisible(email);
  },
);

Given(
  "a diocese {string} and parish {string} exist",
  async function (this: DomusWorld, dioceseName: string, parishName: string) {
    const suffix = Date.now().toString(36);
    const dioceseRes = await this.client.post<{ data?: { id?: string } }>(
      "/ref/dioceses",
      {
        code: `e2e-d-${suffix}`,
        name: dioceseName,
      },
    );
    assert.equal(dioceseRes.status, 201);
    const dioceseId = dioceseRes.data?.data?.id;
    assert.ok(dioceseId, "Expected created diocese ID");

    const parishRes = await this.client.post<{ data?: { id?: string } }>(
      "/ref/parishes",
      {
        dioceseId,
        code: `e2e-p-${suffix}`,
        name: parishName,
      },
    );
    assert.equal(parishRes.status, 201);
  },
);

When(
  "I switch users scope to {string}",
  async function (this: DomusWorld, scopeLabel: string) {
    await this.pages.user.users.selectScope(scopeLabel as UserScopeLabel);
  },
);

When(
  "I select parish {string}",
  async function (this: DomusWorld, parishName: string) {
    await this.pages.user.users.selectParish(parishName);
  },
);

When(
  "I appoint parish member with email {string} name {string} and role {string}",
  async function (
    this: DomusWorld,
    email: string,
    name: string,
    roleLabel: string,
  ) {
    await this.pages.user.users.addParishMember({
      email,
      name,
      roleLabel: roleLabel as "Admin Paroki" | "Anggota",
    });
  },
);

Then(
  "I see parish member {string} with role {string}",
  async function (this: DomusWorld, name: string, roleLabel: string) {
    await this.pages.user.users.expectParishMemberVisible(
      name,
      roleLabel as "Admin Paroki" | "Anggota",
    );
  },
);
