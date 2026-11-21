import assert from "node:assert/strict";
import { type DataTable, Given, Then, When } from "@cucumber/cucumber";
import type { ParishProfileInput } from "../../pages/ref/profile.page";
import { loginAsSuperAdmin } from "../../support/auth";
import { TestClient } from "../../support/test-client";
import type { DomusWorld } from "../../support/world";

Then("I see the parish profile page", async function (this: DomusWorld) {
  await this.pages.ref.profile.expectLoaded();
});

When(
  "I update the parish profile with:",
  async function (this: DomusWorld, table: DataTable) {
    const raw = table.rowsHash() as ParishProfileInput;
    await this.pages.ref.profile.updateProfile(raw);
  },
);

Then("I see the profile saved banner", async function (this: DomusWorld) {
  await this.pages.ref.profile.expectSavedBanner();
});

Then(
  "I see the parish profile fields match:",
  async function (this: DomusWorld, table: DataTable) {
    const raw = table.rowsHash() as ParishProfileInput;
    await this.pages.ref.profile.expectProfileValues(raw);
  },
);

Then("I see the territory page", async function (this: DomusWorld) {
  await this.pages.ref.territory.expectLoaded();
});

When(
  "I create a zone with code {string} and name {string}",
  async function (this: DomusWorld, code: string, name: string) {
    await this.pages.ref.territory.createZone({ code, name });
  },
);

Then(
  "I see zone {string} with code {string}",
  async function (this: DomusWorld, name: string, code: string) {
    await this.pages.ref.territory.expectZoneVisible(name, code);
  },
);

When(
  "I create a ward under zone {string} with code {string} and name {string}",
  async function (
    this: DomusWorld,
    zoneName: string,
    code: string,
    name: string,
  ) {
    await this.pages.ref.territory.createWard({ zoneName, code, name });
  },
);

Then(
  "I see ward {string} with code {string}",
  async function (this: DomusWorld, name: string, code: string) {
    await this.pages.ref.territory.expectWardVisible(name, code);
  },
);

When("I delete zone {string}", async function (this: DomusWorld, name: string) {
  await this.pages.ref.territory.deleteZone(name);
});

Then(
  "I do not see zone {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.territory.expectZoneNotVisible(name);
  },
);

When("I delete ward {string}", async function (this: DomusWorld, name: string) {
  await this.pages.ref.territory.deleteWard(name);
});

Then(
  "I do not see ward {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.territory.expectWardNotVisible(name);
  },
);

Then(
  "I see territory error {string}",
  async function (this: DomusWorld, message: string) {
    await this.pages.ref.territory.expectErrorAlert(message);
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
    const superAdmin = loginAsSuperAdmin();
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
  "I create a parish lookup in category {string} with code {string} and label {string}",
  async function (
    this: DomusWorld,
    category: string,
    code: string,
    label: string,
  ) {
    await this.pages.ref.lookups.createLookup({ category, code, label });
  },
);

Then("I see the periods page", async function (this: DomusWorld) {
  await this.pages.ref.periods.expectLoaded();
});

When(
  "I create a fiscal period with code {string} name {string} from {string} to {string} as current",
  async function (
    this: DomusWorld,
    code: string,
    name: string,
    startDate: string,
    endDate: string,
  ) {
    await this.pages.ref.periods.createPeriod({
      type: "fiscal",
      code,
      name,
      startDate,
      endDate,
      isCurrent: true,
    });
  },
);

When(
  "I create a fiscal period with code {string} name {string} from {string} to {string}",
  async function (
    this: DomusWorld,
    code: string,
    name: string,
    startDate: string,
    endDate: string,
  ) {
    await this.pages.ref.periods.createPeriod({
      type: "fiscal",
      code,
      name,
      startDate,
      endDate,
      isCurrent: false,
    });
  },
);

When(
  "I make period {string} current",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.periods.makeCurrent(name);
  },
);

When(
  "I close period {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.periods.closePeriod(name);
  },
);

When(
  "I delete period {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.periods.deletePeriod(name);
  },
);

Then(
  "I see period {string} with status {string} and current badge",
  async function (this: DomusWorld, name: string, status: string) {
    await this.pages.ref.periods.expectPeriodVisible(name, {
      status: status as "Terbuka" | "Ditutup",
      isCurrent: true,
    });
  },
);

Then(
  "I see period {string} with status {string}",
  async function (this: DomusWorld, name: string, status: string) {
    await this.pages.ref.periods.expectPeriodVisible(name, {
      status: status as "Terbuka" | "Ditutup",
    });
  },
);

Then(
  "I see period {string} without current badge",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.periods.expectPeriodVisible(name, {
      isCurrent: false,
    });
  },
);

Then(
  "I do not see period {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.periods.expectPeriodNotVisible(name);
  },
);

Then(
  "I see period error {string}",
  async function (this: DomusWorld, message: string) {
    await this.pages.ref.periods.expectErrorAlert(message);
  },
);
