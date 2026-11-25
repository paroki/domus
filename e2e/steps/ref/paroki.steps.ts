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
  "I see the paroki admin restriction banner",
  async function (this: DomusWorld) {
    await this.pages.ref.settings.expectParokiAdminRestrictionBanner();
  },
);

Then("I see the paroki page", async function (this: DomusWorld) {
  await this.pages.ref.paroki.expectLoaded();
});

Then(
  "I see the super admin restriction banner on paroki page",
  async function (this: DomusWorld) {
    await this.pages.ref.paroki.expectForbiddenBanner();
  },
);

When(
  "I create a keuskupan with code {string} and name {string}",
  async function (this: DomusWorld, code: string, name: string) {
    await this.pages.ref.paroki.createKeuskupan({ code, name });
  },
);

Then(
  "I see keuskupan {string} with code {string}",
  async function (this: DomusWorld, name: string, code: string) {
    await this.pages.ref.paroki.expectKeuskupanVisible(name, code);
  },
);

When(
  "I create a paroki under keuskupan {string} with code {string} and name {string}",
  async function (
    this: DomusWorld,
    keuskupanName: string,
    code: string,
    name: string,
  ) {
    await this.pages.ref.paroki.createParoki({ keuskupanName, code, name });
  },
);

Then(
  "I see paroki {string} with code {string}",
  async function (this: DomusWorld, name: string, code: string) {
    await this.pages.ref.paroki.expectParokiVisible(name, code);
  },
);

When(
  "I delete paroki {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.paroki.deleteParoki(name);
  },
);

Then(
  "I do not see paroki {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.paroki.expectParokiNotVisible(name);
  },
);

When(
  "I delete keuskupan {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.paroki.deleteKeuskupan(name);
  },
);

Then(
  "I do not see keuskupan {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.paroki.expectKeuskupanNotVisible(name);
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

When(
  "I create a kevikepan under keuskupan {string} with code {string} and name {string}",
  async function (
    this: DomusWorld,
    keuskupanName: string,
    code: string,
    name: string,
  ) {
    await this.pages.ref.kevikepanSipil.createKevikepan({
      keuskupanName,
      code,
      name,
    });
  },
);

Then(
  "I see kevikepan {string} with code {string}",
  async function (this: DomusWorld, name: string, code: string) {
    await this.pages.ref.kevikepanSipil.expectKevikepanVisible(name, code);
  },
);

When(
  "I delete kevikepan {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.kevikepanSipil.deleteKevikepan(name);
  },
);

Then(
  "I do not see kevikepan {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.kevikepanSipil.expectKevikepanNotVisible(name);
  },
);

When(
  "I create a kota with code {string} and name {string}",
  async function (this: DomusWorld, code: string, name: string) {
    await this.pages.ref.kevikepanSipil.createKota({ code, name });
  },
);

When(
  "I create a kecamatan under kota {string} with code {string} and name {string}",
  async function (
    this: DomusWorld,
    kotaName: string,
    code: string,
    name: string,
  ) {
    await this.pages.ref.kevikepanSipil.createKecamatan({
      kotaName,
      code,
      name,
    });
  },
);

When(
  "I create a kelurahan under kecamatan {string} with code {string} name {string} and kodePos {string}",
  async function (
    this: DomusWorld,
    kecamatanName: string,
    code: string,
    name: string,
    kodePos: string,
  ) {
    await this.pages.ref.kevikepanSipil.createKelurahan({
      kecamatanName,
      code,
      name,
      kodePos,
    });
  },
);

When("I delete kota {string}", async function (this: DomusWorld, name: string) {
  await this.pages.ref.kevikepanSipil.deleteKota(name);
});

Then(
  "I do not see kota {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.kevikepanSipil.expectKotaNotVisible(name);
  },
);

When(
  "I delete kecamatan {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.kevikepanSipil.deleteKecamatan(name);
  },
);

Then(
  "I do not see kecamatan {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.kevikepanSipil.expectKecamatanNotVisible(name);
  },
);

When(
  "I delete kelurahan {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.kevikepanSipil.deleteKelurahan(name);
  },
);

Then(
  "I do not see kelurahan {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.ref.kevikepanSipil.expectKelurahanNotVisible(name);
  },
);

Then(
  "I see ref error {string}",
  async function (this: DomusWorld, message: string) {
    await this.pages.ref.kevikepanSipil.expectErrorAlert(message);
  },
);

When(
  "I edit paroki struktur kodeResmi to {string}",
  async function (this: DomusWorld, kodeResmi: string) {
    await this.pages.ref.paroki.editParokiStruktur({ kodeResmi });
  },
);
