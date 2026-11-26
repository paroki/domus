import assert from "node:assert/strict";
import { type DataTable, Given, Then, When } from "@cucumber/cucumber";
import type { ParokiProfileInput } from "../../pages/ref/paroki-profile.page";
import { loginAsSuperAdmin } from "../../support/auth";
import { TestClient } from "../../support/test-client";
import type { DomusWorld } from "../../support/world";

Then("I see the paroki profile page", async function (this: DomusWorld) {
  await this.pages.ref.profile.expectLoaded();
});

When(
  "I update the paroki profile with:",
  async function (this: DomusWorld, table: DataTable) {
    const raw = table.rowsHash() as ParokiProfileInput;
    await this.pages.ref.profile.updateProfile(raw);
  },
);

Then("I see the profile saved banner", async function (this: DomusWorld) {
  await this.pages.ref.profile.expectSavedBanner();
});

Then(
  "I see the paroki profile fields match:",
  async function (this: DomusWorld, table: DataTable) {
    const raw = table.rowsHash() as ParokiProfileInput;
    await this.pages.ref.profile.expectProfileValues(raw);
  },
);

Then("I see the wilayah page", async function (this: DomusWorld) {
  await this.pages.ref.wilayah.expectLoaded();
});

When(
  "I create a wilayah with code {string} and name {string}",
  async function (this: DomusWorld, code: string, name: string) {
    await this.pages.ref.wilayah.createWilayah({ code, name });
  },
);

Then(
  "I see wilayah {string} with code {string}",
  async function (this: DomusWorld, name: string, code: string) {
    await this.pages.ref.wilayah.expectWilayahVisible(name, code);
  },
);

When(
  "I create a lingkungan under wilayah {string} with code {string} and name {string}",
  async function (
    this: DomusWorld,
    wilayahName: string,
    code: string,
    name: string,
  ) {
    await this.pages.ref.wilayah.createLingkungan({ wilayahName, code, name });
  },
);

Then(
  "I see lingkungan {string} with code {string}",
  async function (this: DomusWorld, name: string, code: string) {
    await this.pages.ref.wilayah.expectLingkunganVisible(name, code);
  },
);

When(
  "I delete wilayah {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.wilayah.deleteWilayah(name);
  },
);

Then(
  "I do not see wilayah {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.wilayah.expectWilayahNotVisible(name);
  },
);

When(
  "I delete lingkungan {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.wilayah.deleteLingkungan(name);
  },
);

Then(
  "I do not see lingkungan {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.wilayah.expectLingkunganNotVisible(name);
  },
);

Then(
  "I see wilayah error {string}",
  async function (this: DomusWorld, message: string) {
    await this.pages.ref.wilayah.expectErrorAlert(message);
  },
);

Given(
  "a global lookup exists in category {string} with code {string} and label {string}",
  async function (
    this: DomusWorld,
    category: string,
    code: string,
    label: string,
  ) {
    const superAdmin = await loginAsSuperAdmin();
    const adminClient = new TestClient();
    adminClient.setToken(superAdmin.rawToken);
    const res = await adminClient.post("/ref/lookups", {
      category,
      code: `${code}-${Date.now().toString(36)}`,
      label,
      sortOrder: 0,
      global: true,
    });
    assert.ok(
      res.status === 201 || res.status === 409,
      `Expected 201 or 409 creating global lookup, got ${res.status}`,
    );
  },
);

Then(
  "lookup {string} is read-only",
  async function (this: DomusWorld, label: string) {
    await this.pages.ref.lookups.expectLookupReadOnly(label);
  },
);

When(
  "I create a paroki lookup in category {string} with code {string} and label {string}",
  async function (
    this: DomusWorld,
    category: string,
    code: string,
    label: string,
  ) {
    await this.pages.ref.lookups.createLookup({ category, code, label });
  },
);
