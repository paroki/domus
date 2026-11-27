import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import { loginAsSuperAdmin, loginAsUser } from "../../support/auth";
import { TestClient } from "../../support/test-client";
import type { DomusWorld } from "../../support/world";

async function ensureUserExists(email: string, name: string): Promise<string> {
  const sa = await loginAsSuperAdmin();
  const saClient = new TestClient();
  saClient.setToken(sa.rawToken);
  const res = await saClient.post("/users", {
    email,
    name,
    isSuperAdmin: false,
  });
  assert.ok(
    res.status === 201 || res.status === 409,
    `Expected 201 or 409 creating user ${email}, got ${res.status}`,
  );
  const searchRes = await saClient.post("/users/search", { q: email });
  const body = searchRes.data as {
    data?: Array<{ id: string; email: string }>;
  };
  const found = body.data?.find((u) => u.email === email);
  assert.ok(found?.id, `Expected user ${email} to exist`);
  return found.id;
}

Given(
  "a global user {string} named {string} exists",
  async function (this: DomusWorld, email: string, name: string) {
    await ensureUserExists(email, name);
  },
);

async function signInAsCantusAdmin(world: DomusWorld, email: string) {
  const userId = await ensureUserExists(email, "Cantus Admin E2E");
  const sa = await loginAsSuperAdmin();
  const saClient = new TestClient();
  saClient.setToken(sa.rawToken);
  await saClient.patch(`/users/${userId}`, { isCantusAdmin: true });

  const session = await loginAsUser({ email, name: "Cantus Admin E2E" });
  await world.context.clearCookies();
  await world.context.addCookies(session.cookies);
  world.client.setToken(session.rawToken);
}

Given(
  "I am signed in as cantus admin {string}",
  async function (this: DomusWorld, email: string) {
    await signInAsCantusAdmin(this, email);
  },
);

When(
  "I sign in as cantus admin {string}",
  async function (this: DomusWorld, email: string) {
    await signInAsCantusAdmin(this, email);
  },
);

When(
  "I sign in as regular global user {string}",
  async function (this: DomusWorld, email: string) {
    await ensureUserExists(email, "Parish Member E2E");
    const session = await loginAsUser({ email, name: "Parish Member E2E" });
    await this.context.clearCookies();
    await this.context.addCookies(session.cookies);
    this.client.setToken(session.rawToken);
  },
);

When("I sign in as super admin", async function (this: DomusWorld) {
  const sa = await loginAsSuperAdmin();
  await this.context.clearCookies();
  await this.context.addCookies(sa.cookies);
  this.client.setToken(sa.rawToken);
});

Then(
  "I see the cantus directory without create CTA",
  async function (this: DomusWorld) {
    await this.pages.cantus.expectDirectoryLoaded();
    await this.pages.cantus.expectCreateButtonVisible(false);
  },
);

Then("I see the cantus forbidden card", async function (this: DomusWorld) {
  await this.pages.cantus.expectForbidden();
});

When("I grant cantus admin on user detail", async function (this: DomusWorld) {
  await this.pages.user.users.grantCantusAdmin();
});

When("I revoke cantus admin on user detail", async function (this: DomusWorld) {
  await this.pages.user.users.revokeCantusAdmin();
});

Then(
  "I see the cantus admin badge on user detail",
  async function (this: DomusWorld) {
    await this.pages.user.users.expectDetailCantusAdminBadge(true);
  },
);

Then(
  "I do not see the cantus admin badge on user detail",
  async function (this: DomusWorld) {
    await this.pages.user.users.expectDetailCantusAdminBadge(false);
  },
);

Then(
  "I see user {string} with cantus admin badge in users directory",
  async function (this: DomusWorld, email: string) {
    await this.pages.user.users.search(email);
    await this.pages.user.users.expectUserVisible("Cantus Admin E2E", email, {
      hasCantusAdminBadge: true,
    });
  },
);

When(
  "I create a lagu entry via cantus wizard titled {string}",
  async function (this: DomusWorld, judul: string) {
    await this.pages.cantus.createLaguViaWizard({
      judul,
      sumber: "PS",
      nomor: "PS-320",
      kategori: "Pembuka",
      nadaDasar: "D",
      tags: ["Adven", "SATB"],
      lirik: "Hai makhluk semua, pujilah Tuhan!",
      birama: "4/4",
      tempo: "Moderato",
      kreditLagu: "Antonius E2E",
      media: [
        {
          type: "IMAGE",
          title: "Partitur Gambar",
          viewUrl: "https://cdn.example.com/ps320.png",
        },
        {
          type: "AUDIO",
          title: "Audio SATB",
          viewUrl: "https://cdn.example.com/ps320-satb.mp3",
        },
        {
          type: "PDF",
          title: "Partitur PDF",
          viewUrl: "https://cdn.example.com/ps320.pdf",
        },
      ],
    });
  },
);

Then(
  "I see the cantus detail page for {string} with inline media",
  async function (this: DomusWorld, judul: string) {
    await this.pages.cantus.expectDetailLoaded(judul);
    await this.pages.cantus.expectInlineMedia({
      imageAlt: "Partitur Gambar",
      audioTitle: "Audio SATB",
      pdfTitle: "Partitur PDF",
    });
  },
);

When(
  "I update cantus nada dasar to {string} and lirik to {string}",
  async function (this: DomusWorld, nadaDasar: string, lirik: string) {
    await this.pages.cantus.editIdentityNadaDasar(nadaDasar);
    await this.pages.cantus.editLirik(lirik);
  },
);

Then(
  "I see {string} and {string} on cantus detail",
  async function (this: DomusWorld, text1: string, text2: string) {
    await this.page.getByText(text1).first().waitFor({ state: "visible" });
    await this.page.getByText(text2).first().waitFor({ state: "visible" });
  },
);

When(
  "I create a mazmur entry via cantus wizard titled {string}",
  async function (this: DomusWorld, judul: string) {
    await this.pages.cantus.createMazmurViaWizard({
      judul,
      sumber: "Mazmur",
      nomor: "MT-001",
      tags: ["Tahun A", "Adven"],
      perayaan: "Minggu Adven I",
      siklusTahun: "A",
      masa: "Adven",
      refAlkitab: "Mzm 25:4-5",
      ulangan: "Kepada-Mu ya Tuhan, kuangkat jiwaku.",
      bait: "Beritahukanlah jalan-Mu kepadaku, ya Tuhan.",
      bpi: "Alleluya, Alleluya. Perlihatkanlah kepada kami kasih setia-Mu, ya Tuhan.",
    });
  },
);

Then(
  "I see the mazmur refrain, bait, and BPI blocks on cantus detail",
  async function (this: DomusWorld) {
    await this.pages.cantus.expectMazmurSections({
      ulangan: "Kepada-Mu ya Tuhan, kuangkat jiwaku.",
      bait: "Beritahukanlah jalan-Mu kepadaku, ya Tuhan.",
      bpi: "Alleluya, Alleluya. Perlihatkanlah kepada kami kasih setia-Mu, ya Tuhan.",
    });
  },
);

When(
  "I search cantus directory for {string} with sumber {string} and tag {string}",
  async function (
    this: DomusWorld,
    query: string,
    sumber: string,
    tag: string,
  ) {
    await this.pages.cantus.expectDirectoryLoaded();
    await this.pages.cantus.searchDirectory(query);
    await this.pages.cantus.filterJenis("Lagu Liturgi");
    await this.pages.cantus.filterSumber(sumber);
    await this.pages.cantus.filterTag(tag);
  },
);

Then(
  "I see cantus song {string} in the directory",
  async function (this: DomusWorld, judul: string) {
    await this.pages.cantus.expectSongVisible(judul);
  },
);

When(
  "I open cantus detail for {string}",
  async function (this: DomusWorld, judul: string) {
    await this.pages.cantus.openSongDetail(judul);
  },
);

Then(
  "I see cantus detail in read-only mode",
  async function (this: DomusWorld) {
    await this.pages.cantus.expectReadOnlyDetail();
  },
);

When(
  "I deactivate and soft-delete the cantus entry",
  async function (this: DomusWorld) {
    await this.pages.cantus.deactivateSong();
    await this.pages.cantus.deleteSong();
  },
);

Then(
  "I do not see cantus song {string} in the directory",
  async function (this: DomusWorld, judul: string) {
    await this.pages.cantus.expectSongNotVisible(judul);
  },
);
