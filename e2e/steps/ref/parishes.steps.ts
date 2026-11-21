import { Then, When } from "@cucumber/cucumber";
import type { DomusWorld } from "../../support/world";

function splitCsv(raw: string): string[] {
  return raw
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

Then("I see the settings page", async function (this: DomusWorld) {
  await this.pages.ref.settings.expectLoaded();
});

Then(
  "I see settings tiles {string}",
  async function (this: DomusWorld, labels: string) {
    await this.pages.ref.settings.expectVisibleTiles(splitCsv(labels));
  },
);

Then(
  "I do not see settings tiles {string}",
  async function (this: DomusWorld, labels: string) {
    await this.pages.ref.settings.expectHiddenTiles(splitCsv(labels));
  },
);

Then(
  "I see the parish admin restriction banner",
  async function (this: DomusWorld) {
    await this.pages.ref.settings.expectParishAdminRestrictionBanner();
  },
);

Then("I see the parishes page", async function (this: DomusWorld) {
  await this.pages.ref.parishes.expectLoaded();
});

Then(
  "I see the super admin restriction banner on parishes page",
  async function (this: DomusWorld) {
    await this.pages.ref.parishes.expectForbiddenBanner();
  },
);

When(
  "I create a diocese with code {string} and name {string}",
  async function (this: DomusWorld, code: string, name: string) {
    await this.pages.ref.parishes.createDiocese({ code, name });
  },
);

Then(
  "I see diocese {string} with code {string}",
  async function (this: DomusWorld, name: string, code: string) {
    await this.pages.ref.parishes.expectDioceseVisible(name, code);
  },
);

When(
  "I create a parish under diocese {string} with code {string} and name {string}",
  async function (
    this: DomusWorld,
    dioceseName: string,
    code: string,
    name: string,
  ) {
    await this.pages.ref.parishes.createParish({ dioceseName, code, name });
  },
);

Then(
  "I see parish {string} with code {string}",
  async function (this: DomusWorld, name: string, code: string) {
    await this.pages.ref.parishes.expectParishVisible(name, code);
  },
);

When(
  "I delete parish {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.parishes.deleteParish(name);
  },
);

Then(
  "I do not see parish {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.parishes.expectParishNotVisible(name);
  },
);

When(
  "I delete diocese {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.parishes.deleteDiocese(name);
  },
);

Then(
  "I do not see diocese {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.parishes.expectDioceseNotVisible(name);
  },
);

Then("I see the lookups page", async function (this: DomusWorld) {
  await this.pages.ref.lookups.expectLoaded();
});

When(
  "I create a global lookup in category {string} with code {string} and label {string}",
  async function (
    this: DomusWorld,
    category: string,
    code: string,
    label: string,
  ) {
    await this.pages.ref.lookups.createLookup({ category, code, label });
  },
);

Then(
  "I see lookup {string} with scope {string}",
  async function (this: DomusWorld, label: string, scopeLabel: string) {
    await this.pages.ref.lookups.expectLookupVisible(label, scopeLabel);
  },
);

When(
  "I edit lookup {string} to label {string}",
  async function (this: DomusWorld, currentLabel: string, newLabel: string) {
    await this.pages.ref.lookups.editLookup(currentLabel, { label: newLabel });
  },
);

When(
  "I toggle active state for lookup {string}",
  async function (this: DomusWorld, label: string) {
    await this.pages.ref.lookups.toggleLookupActive(label);
  },
);

When(
  "I delete lookup {string}",
  async function (this: DomusWorld, label: string) {
    await this.pages.ref.lookups.deleteLookup(label);
  },
);

Then(
  "I do not see lookup {string}",
  async function (this: DomusWorld, label: string) {
    await this.pages.ref.lookups.expectLookupNotVisible(label);
  },
);
