import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import { loginAsSuperAdmin, loginAsUser } from "../../support/auth";
import { TestClient } from "../../support/test-client";
import type { DomusWorld } from "../../support/world";

async function ensureParokiMember(
  parokiId: string,
  email: string,
  name: string,
  extra?: { liturgiAdmin?: boolean },
): Promise<string> {
  const sa = await loginAsSuperAdmin();
  const saClient = new TestClient();
  saClient.setToken(sa.rawToken);

  const res = await saClient.post(`/ref/paroki/${parokiId}/members`, {
    email,
    name,
    role: "member",
    ...extra,
  });
  assert.ok(
    res.status === 201 || res.status === 409,
    `Expected 201 or 409 adding member ${email}, got ${res.status}`,
  );

  const searchRes = await saClient.post("/users/search", { q: email });
  const body = searchRes.data as {
    data?: Array<{ id: string; email: string }>;
  };
  const found = body.data?.find((u) => u.email === email);
  assert.ok(found?.id, `Expected user ${email} to exist`);
  return found.id;
}

async function ensureParokiOrg(
  client: TestClient,
  name: string,
  code: string,
): Promise<string> {
  const res = await client.post<{ data?: { id?: string } }>("/orgs", {
    kind: "section",
    name,
    code,
  });
  assert.equal(
    res.status,
    201,
    `Expected 201 creating org ${name}, got ${res.status}`,
  );
  const orgId = res.data?.data?.id;
  assert.ok(orgId, `Expected orgId for ${name}`);
  return orgId;
}

async function signInOnActiveParoki(
  world: DomusWorld,
  email: string,
  name: string,
): Promise<void> {
  assert.ok(world.activeParokiId, "activeParokiId must be set");
  const session = await loginAsUser({
    email,
    name,
    onboarded: true,
    activeParokiId: world.activeParokiId,
  });
  await world.context.clearCookies();
  await world.context.addCookies(session.cookies);
  world.client.setToken(session.rawToken);
  world.client.setParokiId(world.activeParokiId);
}

Given(
  "a parish choir org {string} coded {string} exists in the active paroki",
  async function (this: DomusWorld, orgName: string, orgCode: string) {
    assert.ok(this.activeParokiId, "activeParokiId must be set");
    const orgId = await ensureParokiOrg(this.client, orgName, orgCode);
    this.activeOrgId = orgId;
    this.activeOrgName = orgName;
  },
);

Given(
  "a liturgi setup with lektor coordinator {string} named {string}, lectors {string} and {string}, and Mass {string}",
  async function (
    this: DomusWorld,
    coordEmail: string,
    coordName: string,
    lector1Name: string,
    lector2Name: string,
    massTitle: string,
  ) {
    assert.ok(this.activeParokiId, "activeParokiId must be set");
    const parokiId = this.activeParokiId;

    const coordId = await ensureParokiMember(parokiId, coordEmail, coordName);
    await ensureParokiMember(parokiId, "lektor1.e2e@domus.test", lector1Name);
    await ensureParokiMember(parokiId, "lektor2.e2e@domus.test", lector2Name);

    const litAdminEmail = "liturgi-setup.e2e@domus.test";
    await ensureParokiMember(parokiId, litAdminEmail, "Liturgi Setup E2E", {
      liturgiAdmin: true,
    });

    const session = await loginAsUser({
      email: litAdminEmail,
      name: "Liturgi Setup E2E",
      onboarded: true,
      activeParokiId: parokiId,
    });
    const adminClient = new TestClient();
    adminClient.setToken(session.rawToken);
    adminClient.setParokiId(parokiId);

    const mapRes = await adminClient.patch("/liturgi/ministries/lektor", {
      leaderUserId: coordId,
    });
    assert.equal(mapRes.status, 200, "Expected 200 mapping lektor ministry");

    const schedRes = await adminClient.post<{ data?: { id?: string } }>(
      "/liturgi/schedules",
      {
        title: massTitle,
        location: "Gereja Utama",
        startAt: "2099-11-10T09:00:00+08:00",
        endAt: "2099-11-10T10:30:00+08:00",
        rsvpEnabled: true,
      },
    );
    assert.equal(
      schedRes.status,
      201,
      `Expected 201 creating Mass schedule, got ${schedRes.status}`,
    );
    this.activeLiturgiEventId = schedRes.data?.data?.id;
    assert.ok(this.activeLiturgiEventId, "Expected activeLiturgiEventId");
  },
);

Given(
  "a liturgi song setup with choir {string} led by {string}, Mass {string}, and cantus song {string}",
  async function (
    this: DomusWorld,
    choirName: string,
    choirAdminEmail: string,
    massTitle: string,
    cantusTitle: string,
  ) {
    assert.ok(this.activeParokiId, "activeParokiId must be set");
    const parokiId = this.activeParokiId;

    const orgId = await ensureParokiOrg(this.client, choirName, "kor-e2e");
    await ensureParokiMember(parokiId, choirAdminEmail, "Kor Admin E2E");

    const sa = await loginAsSuperAdmin();
    const saClient = new TestClient();
    saClient.setToken(sa.rawToken);
    saClient.setParokiId(parokiId);

    const joinRes = await this.client.post(`/orgs/${orgId}/members`, {
      email: choirAdminEmail,
      role: "admin",
    });
    assert.ok(
      joinRes.status === 201 || joinRes.status === 409,
      `Expected 201 or 409 adding choir admin, got ${joinRes.status}`,
    );

    const litAdminEmail = "liturgi-song-setup.e2e@domus.test";
    await ensureParokiMember(
      parokiId,
      litAdminEmail,
      "Liturgi Song Setup E2E",
      {
        liturgiAdmin: true,
      },
    );
    const litSession = await loginAsUser({
      email: litAdminEmail,
      name: "Liturgi Song Setup E2E",
      onboarded: true,
      activeParokiId: parokiId,
    });
    const litClient = new TestClient();
    litClient.setToken(litSession.rawToken);
    litClient.setParokiId(parokiId);

    const mapRes = await litClient.patch("/liturgi/ministries/kor", { orgId });
    assert.equal(mapRes.status, 200, "Expected 200 mapping kor ministry");

    const schedRes = await litClient.post<{ data?: { id?: string } }>(
      "/liturgi/schedules",
      {
        title: massTitle,
        location: "Gereja Utama",
        startAt: "2099-11-11T17:00:00+08:00",
        endAt: "2099-11-11T18:30:00+08:00",
        rsvpEnabled: true,
      },
    );
    assert.equal(schedRes.status, 201, "Expected 201 creating Mass schedule");
    this.activeLiturgiEventId = schedRes.data?.data?.id;
    assert.ok(this.activeLiturgiEventId, "Expected activeLiturgiEventId");

    // Seed global cantus song as cantus:admin
    const cantusAdminId = await ensureParokiMember(
      parokiId,
      "cantus-for-liturgi@domus.test",
      "Cantus Seed E2E",
    );
    await saClient.patch(`/users/${cantusAdminId}`, { isCantusAdmin: true });
    const cantusSession = await loginAsUser({
      email: "cantus-for-liturgi@domus.test",
      name: "Cantus Seed E2E",
    });
    const cantusClient = new TestClient();
    cantusClient.setToken(cantusSession.rawToken);

    const cantusRes = await cantusClient.post<{ data?: { id?: string } }>(
      "/cantus",
      {
        slug: `tuhan-kasihanilah-e2e-${Date.now().toString(36)}`,
        jenis: "lagu",
        judul: cantusTitle,
        sumber: "PS",
        nomor: "PS-351",
        nadaDasar: "F",
        lirik: "Tuhan kasihanilah kami.",
      },
    );
    assert.equal(
      cantusRes.status,
      201,
      `Expected 201 creating cantus song, got ${cantusRes.status}`,
    );
    this.activeCantusId = cantusRes.data?.data?.id;
    assert.ok(this.activeCantusId, "Expected activeCantusId");
  },
);

When(
  "I grant liturgi admin on the active paroki",
  async function (this: DomusWorld) {
    assert.ok(this.activeParokiName, "activeParokiName must be set");
    await this.pages.user.users.grantLiturgiAdmin(this.activeParokiName);
  },
);

Then(
  "I see the liturgi admin badge on the active paroki",
  async function (this: DomusWorld) {
    assert.ok(this.activeParokiName, "activeParokiName must be set");
    await this.pages.user.users.expectLiturgiAdminBadge(
      this.activeParokiName,
      true,
    );
  },
);

When(
  "I sign in as {string} named {string} on the active paroki",
  async function (this: DomusWorld, email: string, name: string) {
    await signInOnActiveParoki(this, email, name);
  },
);

When(
  "I map liturgi ministry {string} to leader {string} with query {string}",
  async function (
    this: DomusWorld,
    roleLabel: string,
    leaderName: string,
    leaderQuery: string,
  ) {
    await this.pages.liturgi.expectMinistriesLoaded();
    await this.pages.liturgi.mapMinistry(roleLabel, {
      leaderName,
      leaderQuery,
    });
  },
);

When(
  "I map liturgi ministry {string} to org {string}",
  async function (this: DomusWorld, roleLabel: string, orgName: string) {
    await this.pages.liturgi.expectMinistriesLoaded();
    await this.pages.liturgi.mapMinistry(roleLabel, { orgName });
  },
);

When(
  "I create a liturgi schedule titled {string} at {string} start {string} end {string}",
  async function (
    this: DomusWorld,
    title: string,
    location: string,
    startAt: string,
    endAt: string,
  ) {
    this.activeLiturgiEventId = await this.pages.liturgi.createSchedule({
      title,
      location,
      startAt,
      endAt,
    });
  },
);

Then(
  "I see liturgi schedule {string} in desktop 7-column matrix and mobile card stack",
  async function (this: DomusWorld, title: string) {
    await this.pages.liturgi.expectSchedulesLoaded();
    await this.pages.liturgi.expectScheduleInDesktopMatrix(title);
    await this.pages.liturgi.expectScheduleInMobileStack(title);
  },
);

Then(
  "I see liturgi detail {string}",
  async function (this: DomusWorld, title: string) {
    await this.pages.liturgi.expectDetailLoaded(title);
  },
);

When(
  "I submit liturgi RSVP {string}",
  async function (this: DomusWorld, choice: string) {
    await this.pages.liturgi.submitRsvp(
      choice as "Hadir" | "Mungkin" | "Tidak Hadir",
    );
  },
);

Then(
  "I see liturgi RSVP status {string}",
  async function (this: DomusWorld, choice: string) {
    await this.page
      .getByText(`Status Anda: ${choice}`, { exact: true })
      .waitFor({ state: "visible", timeout: 10_000 });
  },
);

When("I open the active liturgi schedule", async function (this: DomusWorld) {
  assert.ok(this.activeLiturgiEventId, "activeLiturgiEventId must be set");
  await this.pages.base.goto(`/liturgi/${this.activeLiturgiEventId}`);
});

When("I open add liturgi petugas form", async function (this: DomusWorld) {
  await this.pages.liturgi.openAddPetugas();
});

Then(
  "I see only role {string} in the liturgi role select",
  async function (this: DomusWorld, roleLabel: string) {
    await this.pages.liturgi.expectAllowedRoleOptions([roleLabel]);
  },
);

When(
  "I assign liturgi petugas slot {string} to member {string} with query {string}",
  async function (
    this: DomusWorld,
    slotLabel: string,
    memberName: string,
    memberQuery: string,
  ) {
    await this.pages.liturgi.assignPetugas({
      slotLabel,
      mode: "user",
      memberName,
      memberQuery,
    });
  },
);

When(
  "I filter liturgi schedules by {string}",
  async function (this: DomusWorld, label: string) {
    await this.pages.liturgi.expectSchedulesLoaded();
    await this.pages.liturgi.filterSchedulesBy(
      label as "Semua Jadwal" | "Jadwal Tugas Saya" | "Perlu Pengganti",
    );
  },
);

When(
  "I open liturgi schedule {string} from the matrix",
  async function (this: DomusWorld, title: string) {
    await this.pages.liturgi.openScheduleFromList(title);
  },
);

When(
  "I decline liturgi slot {string} with note {string} and propose substitute {string} with query {string}",
  async function (
    this: DomusWorld,
    slotLabel: string,
    note: string,
    memberName: string,
    memberQuery: string,
  ) {
    await this.pages.liturgi.openRespondForPetugas(slotLabel);
    await this.pages.liturgi.submitDeclineWithSubstitute({
      note,
      subMode: "user",
      memberName,
      memberQuery,
    });
  },
);

Then(
  "I see substitute proposal {string} on liturgi slot {string}",
  async function (this: DomusWorld, substituteName: string, slotLabel: string) {
    await this.pages.liturgi.expectSubstituteCallout(slotLabel, substituteName);
  },
);

When(
  "I approve substitute on liturgi slot {string}",
  async function (this: DomusWorld, slotLabel: string) {
    await this.pages.liturgi.approveSubstitute(slotLabel);
  },
);

Then(
  "I see {string} as the main officer for liturgi slot {string}",
  async function (this: DomusWorld, name: string, slotLabel: string) {
    await this.pages.liturgi.expectPetugasVisible(slotLabel, name);
  },
);

When("I open add liturgi lagu form", async function (this: DomusWorld) {
  await this.pages.liturgi.openAddLagu();
});

When(
  "I add liturgi song from cantus {string} for bagian {string} expecting sumber {string} nomor {string} nadaDasar {string}",
  async function (
    this: DomusWorld,
    cantusTitle: string,
    bagian: string,
    expectedSumber: string,
    expectedNomor: string,
    expectedNadaDasar: string,
  ) {
    await this.pages.liturgi.addLaguFromCantus({
      bagian,
      searchQuery: cantusTitle,
      cantusTitle,
      expectedSumber,
      expectedNomor,
      expectedNadaDasar,
    });
  },
);

When(
  "I add manual liturgi song titled {string} for bagian {string}",
  async function (this: DomusWorld, judul: string, bagian: string) {
    await this.pages.liturgi.addLaguManual({ bagian, judul });
  },
);

Then(
  "I see catalog-linked liturgi song {string} and manual song {string}",
  async function (this: DomusWorld, catalogTitle: string, manualTitle: string) {
    await this.pages.liturgi.expectLaguVisible(catalogTitle, {
      catalogLinked: true,
    });
    await this.pages.liturgi.expectLaguVisible(manualTitle, {
      catalogLinked: false,
    });
  },
);

Then(
  "deleting the referenced cantus song is rejected with 409 CANTUS_IN_USE",
  async function (this: DomusWorld) {
    assert.ok(this.activeCantusId, "activeCantusId must be set");
    const session = await loginAsUser({
      email: "cantus-for-liturgi@domus.test",
      name: "Cantus Seed E2E",
    });
    const cantusClient = new TestClient();
    cantusClient.setToken(session.rawToken);

    const res = await cantusClient.delete<{ error?: { code?: string } }>(
      `/cantus/${this.activeCantusId}`,
    );
    assert.equal(
      res.status,
      409,
      `Expected 409 CANTUS_IN_USE deleting referenced cantus, got ${res.status}`,
    );
    assert.equal(res.data?.error?.code, "CANTUS_IN_USE");
  },
);
