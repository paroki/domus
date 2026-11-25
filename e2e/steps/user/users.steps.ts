import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import type { UserDetailTabLabel } from "../../pages/user/users.page";
import type { DomusWorld } from "../../support/world";

Then("I see the users forbidden banner", async function (this: DomusWorld) {
  await this.pages.user.users.expectForbidden();
});

Then("I see the users directory page", async function (this: DomusWorld) {
  await this.pages.user.users.expectLoaded();
});

Given(
  "a global user with email {string} name {string} phone {string} exists",
  async function (
    this: DomusWorld,
    email: string,
    name: string,
    phone: string,
  ) {
    const res = await this.client.post("/users", { email, name, phone });
    assert.ok(
      res.status === 201 || res.status === 409,
      `Expected 201 or 409 creating global user, got ${res.status}`,
    );
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
  "I open user detail {string}",
  async function (this: DomusWorld, nameOrEmail: string) {
    await this.pages.user.users.openUserDetail(nameOrEmail);
  },
);

When("I go back to the users directory", async function (this: DomusWorld) {
  await this.pages.user.users.backToDirectory();
});

When(
  "I switch user detail tab to {string}",
  async function (this: DomusWorld, tabLabel: string) {
    await this.pages.user.users.selectDetailTab(tabLabel as UserDetailTabLabel);
  },
);

When("I promote the user to Super Admin", async function (this: DomusWorld) {
  await this.pages.user.users.promoteToSuperAdmin();
});

When("I demote the user from Super Admin", async function (this: DomusWorld) {
  await this.pages.user.users.demoteFromSuperAdmin();
});

When("I disable the user", async function (this: DomusWorld) {
  await this.pages.user.users.disableUser();
});

When("I enable the user", async function (this: DomusWorld) {
  await this.pages.user.users.enableUser();
});

When(
  "I filter users by status {string}",
  async function (this: DomusWorld, statusLabel: string) {
    await this.pages.user.users.filterByStatus(statusLabel);
  },
);

When("I delete the user", async function (this: DomusWorld) {
  await this.pages.user.users.deleteUser();
});

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
      "/ref/keuskupan",
      {
        kode: `e2e-d-${suffix}`,
        nama: dioceseName,
      },
    );
    assert.equal(dioceseRes.status, 201);
    const dioceseId = dioceseRes.data?.data?.id;
    assert.ok(dioceseId, "Expected created diocese ID");

    const parishRes = await this.client.post<{ data?: { id?: string } }>(
      "/ref/paroki",
      {
        keuskupanId: dioceseId,
        kode: `e2e-p-${suffix}`,
        nama: parishName,
      },
    );
    assert.equal(parishRes.status, 201);
  },
);

When(
  "I add user to parish {string} with role {string}",
  async function (this: DomusWorld, parishName: string, roleLabel: string) {
    await this.pages.user.users.addToParish(
      parishName,
      roleLabel as "Admin Paroki" | "Anggota",
    );
  },
);

Then(
  "I see user parish membership {string} with role {string}",
  async function (this: DomusWorld, parishName: string, roleLabel: string) {
    await this.pages.user.users.expectParishMembershipVisible(
      parishName,
      roleLabel as "Admin Paroki" | "Anggota",
    );
  },
);
