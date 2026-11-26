import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import type { UserDetailTabLabel } from "../../pages/user/users.page";
import { loginAsSuperAdmin, loginAsUser } from "../../support/auth";
import { TestClient } from "../../support/test-client";
import type { DomusWorld } from "../../support/world";

async function ensureMemberAndUmatAdmin(
  world: DomusWorld,
  email: string,
  name: string,
  grantUmatAdmin: boolean,
) {
  const sa = await loginAsSuperAdmin();
  const saClient = new TestClient();
  saClient.setToken(sa.rawToken);
  const res = await saClient.post(
    `/ref/paroki/${world.activeParokiId}/members`,
    {
      email,
      name,
      role: "member",
      ...(grantUmatAdmin ? { umatAdmin: true } : {}),
    },
  );
  assert.ok(
    res.status === 201 || res.status === 409,
    `Expected 201 or 409 adding member ${email}, got ${res.status}`,
  );
}

Given(
  "a paroki member {string} named {string} exists in the active paroki",
  async function (this: DomusWorld, email: string, name: string) {
    await ensureMemberAndUmatAdmin(this, email, name, false);
  },
);

Given(
  "I am signed in as umat admin on the active paroki",
  async function (this: DomusWorld) {
    await ensureMemberAndUmatAdmin(
      this,
      "umat-e2e@domus.test",
      "Umat E2E",
      true,
    );
    const session = await loginAsUser({
      email: "umat-e2e@domus.test",
      name: "Umat E2E",
      activeParokiId: this.activeParokiId,
    });
    await this.context.addCookies(session.cookies);
    this.client.setToken(session.rawToken);
    if (this.activeParokiId) {
      this.client.setParokiId(this.activeParokiId);
    }
  },
);

When(
  "I sign in as {string} on the active paroki",
  async function (this: DomusWorld, email: string) {
    const session = await loginAsUser({
      email,
      name: "Admin Umat E2E",
      activeParokiId: this.activeParokiId,
    });
    await this.context.addCookies(session.cookies);
    this.client.setToken(session.rawToken);
    if (this.activeParokiId) {
      this.client.setParokiId(this.activeParokiId);
    }
  },
);

Then("I see the umat forbidden card", async function (this: DomusWorld) {
  await this.pages.umat.expectForbidden();
});

Then("I see the umat directory page", async function (this: DomusWorld) {
  await this.pages.umat.expectDirectoryLoaded();
});

When(
  "I open user detail for {string}",
  async function (this: DomusWorld, nameOrEmail: string) {
    await this.pages.user.users.openUserDetail(nameOrEmail);
  },
);

When(
  "I switch to user detail tab {string}",
  async function (this: DomusWorld, tab: string) {
    await this.pages.user.users.selectDetailTab(tab as UserDetailTabLabel);
  },
);

When(
  "I grant umat admin for the active paroki",
  async function (this: DomusWorld) {
    await this.pages.user.users.grantUmatAdmin(this.activeParokiName ?? "");
  },
);

Then(
  "I see the umat admin badge for the active paroki",
  async function (this: DomusWorld) {
    await this.pages.user.users.expectUmatAdminBadge(
      this.activeParokiName ?? "",
      true,
    );
  },
);

When(
  "I create a Kartu Keluarga named {string} with nomorKk {string} and alamat {string}",
  async function (
    this: DomusWorld,
    namaKeluarga: string,
    nomorKk: string,
    alamat: string,
  ) {
    await this.pages.umat.createKeluarga({ namaKeluarga, nomorKk, alamat });
  },
);

Then(
  "I see decrypted nomorKk {string} on keluarga detail",
  async function (this: DomusWorld, rawNomorKk: string) {
    await this.pages.umat.expectKeluargaDetailNomorKk(rawNomorKk);
  },
);

When(
  "I search keluarga by nomorKk {string}",
  async function (this: DomusWorld, nomorKk: string) {
    await this.pages.umat.searchKeluargaByNomorKk(nomorKk);
  },
);

Then(
  "I see masked keluarga {string} with {string} and no raw {string}",
  async function (
    this: DomusWorld,
    namaKeluarga: string,
    masked: string,
    raw: string,
  ) {
    await this.pages.umat.expectKeluargaRowMasked(namaKeluarga, masked, raw);
  },
);

When(
  "I create an umat via wizard named {string} with nik {string} linked to {string}",
  async function (
    this: DomusWorld,
    nama: string,
    nik: string,
    keluargaName: string,
  ) {
    await this.pages.umat.createUmatViaWizard({
      nama,
      nik,
      jenisKelamin: "Pria",
      keluargaName,
      telepon: "081234567890",
      email: "yohanes-e2e@domus.test",
      jurusan: "Teologi",
    });
  },
);

When(
  "I edit umat nama baptis to {string}",
  async function (this: DomusWorld, namaBaptis: string) {
    await this.pages.umat.editUmatNamaBaptis(namaBaptis);
  },
);

When(
  "I link umat to paroki member {string}",
  async function (this: DomusWorld, email: string) {
    await this.pages.umat.linkUserAccount(email);
  },
);

When(
  "I add sakramen {string} on {string} buku {string} nomor {string}",
  async function (
    this: DomusWorld,
    jenisLabel: string,
    tanggal: string,
    buku: string,
    nomor: string,
  ) {
    await this.pages.umat.addOrUpdateSakramen({
      jenisLabel,
      tanggal,
      buku,
      nomor,
    });
  },
);

When(
  "I edit sakramen {string} nomor to {string}",
  async function (this: DomusWorld, jenisLabel: string, nomor: string) {
    await this.pages.umat.editSakramenNomor(jenisLabel, nomor);
  },
);

When(
  "I delete sakramen {string}",
  async function (this: DomusWorld, jenisLabel: string) {
    await this.pages.umat.deleteSakramen(jenisLabel);
  },
);

When(
  "I search umat by nik {string}",
  async function (this: DomusWorld, nik: string) {
    await this.pages.umat.searchUmatByNik(nik);
  },
);

Then(
  "I see masked umat {string} with {string} and no raw {string}",
  async function (this: DomusWorld, nama: string, masked: string, raw: string) {
    await this.pages.umat.expectUmatRowMasked(nama, masked, raw);
  },
);

When(
  "I open keluarga detail for {string}",
  async function (this: DomusWorld, namaKeluarga: string) {
    await this.pages.umat.openKeluargaDetail(namaKeluarga);
  },
);

When("I try to delete the Kartu Keluarga", async function (this: DomusWorld) {
  await this.pages.umat.deleteKeluarga();
});

When("I delete the Kartu Keluarga", async function (this: DomusWorld) {
  await this.pages.umat.deleteKeluarga();
});

Then(
  "I see umat error {string}",
  async function (this: DomusWorld, message: string) {
    await this.pages.umat.expectErrorAlert(message);
  },
);

When(
  "I open umat detail for {string}",
  async function (this: DomusWorld, nama: string) {
    await this.pages.umat.openUmatDetail(nama);
  },
);

When("I delete the umat", async function (this: DomusWorld) {
  await this.pages.umat.deleteUmat();
});
