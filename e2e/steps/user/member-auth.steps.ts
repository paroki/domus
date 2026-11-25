import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import {
  bootstrapOrgAdminSession,
  loginAsUser,
  seedOtpCode,
} from "../../support/auth";
import { TestClient } from "../../support/test-client";
import type { DomusWorld } from "../../support/world";

let currentInviteCode = "";
let lastCreatedPat = "";

When(
  "I request email OTP on login for {string}",
  async function (this: DomusWorld, email: string) {
    await this.pages.user.login.requestEmailOtp(email);
  },
);

Then(
  "I see login error containing {string}",
  async function (this: DomusWorld, text: string) {
    await this.pages.user.login.expectErrorAlert(text);
  },
);

Given(
  "an organization {string} has an active member invite with position {string}",
  async function (this: DomusWorld, orgName: string, defaultPosition: string) {
    const setupClient = new TestClient();
    const { paroki, org } = await bootstrapOrgAdminSession(setupClient, {
      orgName,
    });
    setupClient.setParokiId(paroki.id);
    const invRes = await setupClient.post<{
      data?: { id?: string; code?: string };
    }>(`/orgs/${org.id}/invites`, {
      role: "member",
      defaultPosition,
      ttlSeconds: 86400,
    });
    assert.equal(invRes.status, 201);
    const code = invRes.data?.data?.code;
    assert.ok(code, "Expected invite code");
    currentInviteCode = code;
    this.activeOrgId = org.id;
    this.activeOrgName = org.name;
  },
);

Given(
  "an organization {string} has a revoked invite",
  async function (this: DomusWorld, orgName: string) {
    const setupClient = new TestClient();
    const { paroki, org } = await bootstrapOrgAdminSession(setupClient, {
      orgName,
    });
    setupClient.setParokiId(paroki.id);
    const invRes = await setupClient.post<{
      data?: { id?: string; code?: string };
    }>(`/orgs/${org.id}/invites`, {
      role: "member",
      ttlSeconds: 86400,
    });
    assert.equal(invRes.status, 201);
    const inviteId = invRes.data?.data?.id;
    const code = invRes.data?.data?.code;
    assert.ok(inviteId && code, "Expected invite id and code");

    const delRes = await setupClient.delete(
      `/orgs/${org.id}/invites/${inviteId}`,
    );
    assert.equal(delRes.status, 204);
    currentInviteCode = code;
    this.activeOrgId = org.id;
    this.activeOrgName = org.name;
  },
);

When("I open the organization invite link", async function (this: DomusWorld) {
  await this.pages.base.goto(`/org/join/${currentInviteCode}`);
});

Then(
  "I see invite preview for {string} with role {string} and position {string}",
  async function (
    this: DomusWorld,
    orgName: string,
    roleBadge: string,
    defaultPosition: string,
  ) {
    await this.pages.org.join.expectPreviewLoaded(orgName, {
      roleBadge,
      defaultPosition,
    });
  },
);

When(
  "I request guest OTP with name {string} and email {string}",
  async function (this: DomusWorld, name: string, email: string) {
    await this.pages.org.join.requestGuestOtp({ name, email });
  },
);

When(
  "a deterministic OTP {string} is seeded for {string}",
  async function (this: DomusWorld, code: string, email: string) {
    await seedOtpCode(email, code, "email");
  },
);

When(
  "I verify guest OTP {string} for {string}",
  async function (this: DomusWorld, code: string, orgName: string) {
    await this.pages.org.join.verifyGuestOtp(orgName, code);
  },
);

Then("I am redirected to onboarding", async function (this: DomusWorld) {
  await this.pages.base.expectPath("/onboarding");
  await this.pages.user.onboarding.expectLoaded();
});

Then("I see onboarding birth date bounds", async function (this: DomusWorld) {
  await this.pages.user.onboarding.expectBirthDateBounds();
});

When(
  "I complete onboarding with name {string}, phone {string}, gender {string}, and birth date {string}",
  async function (
    this: DomusWorld,
    name: string,
    phone: string,
    gender: string,
    birthDate: string,
  ) {
    await this.pages.user.onboarding.completeBiodataAndFinish({
      name,
      phone,
      gender: gender as "pria" | "wanita",
      birthDate,
    });
  },
);

Then(
  "I am on the invited organization detail page {string}",
  async function (this: DomusWorld, orgName: string) {
    assert.ok(this.activeOrgId, "Expected activeOrgId");
    await this.pages.base.expectPath(`/org/${this.activeOrgId}`);
    await this.pages.org.detail.expectLoaded(orgName);
  },
);

Then(
  "I see revoked invite message {string}",
  async function (this: DomusWorld, message: string) {
    await this.pages.org.join.expectInvalidOrRevoked(message);
  },
);

Given(
  "a secondary session exists for {string}",
  async function (this: DomusWorld, email: string) {
    await loginAsUser({ email, name: "Regular User", onboarded: true });
  },
);

Then("I see the profile and security page", async function (this: DomusWorld) {
  await this.pages.user.me.expectLoaded();
});

When(
  "I update my profile to name {string}, phone {string}, gender {string}, and birth date {string}",
  async function (
    this: DomusWorld,
    name: string,
    phone: string,
    gender: string,
    birthDate: string,
  ) {
    await this.pages.user.me.updateProfile({
      name,
      phone,
      gender: gender as "pria" | "wanita",
      birthDate,
    });
  },
);

Then(
  "I see my profile shows gender {string} and birth date {string}",
  async function (this: DomusWorld, gender: string, birthDate: string) {
    await this.pages.user.me.expectProfileValues({ gender, birthDate });
  },
);

Then("I see profile saved message", async function (this: DomusWorld) {
  await this.pages.user.me.expectProfileSaved();
});

Then(
  "I see current device session in active sessions",
  async function (this: DomusWorld) {
    await this.pages.user.me.expectCurrentSessionVisible();
  },
);

When("I revoke the secondary session", async function (this: DomusWorld) {
  await this.pages.user.me.revokeOtherSession();
});

Then("I see last linked account protection", async function (this: DomusWorld) {
  await this.pages.user.me.expectLastAccountProtected();
});

When("I log out from header menu", async function (this: DomusWorld) {
  await this.pages.user.me.logout();
});

When(
  "I create a personal access token named {string} with expiry {string}",
  async function (this: DomusWorld, name: string, expiry: string) {
    lastCreatedPat = await this.pages.user.me.createToken(name, expiry);
  },
);

Then(
  "I see the one-time personal access token banner and token {string} in the list",
  async function (this: DomusWorld, name: string) {
    assert.ok(lastCreatedPat.startsWith("dms_"));
    await this.pages.user.me.expectTokenVisible(name);
  },
);

Then(
  "I can authenticate GET {string} with the personal access token",
  async function (this: DomusWorld, apiPath: string) {
    const url = new URL(apiPath, this.page.url()).toString();
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${lastCreatedPat}` },
    });
    assert.equal(
      res.status,
      200,
      `Expected 200 from ${url}, got ${res.status}`,
    );
  },
);

When(
  "I revoke the personal access token {string}",
  async function (this: DomusWorld, name: string) {
    await this.pages.user.me.revokeToken(name);
  },
);

Then(
  "the personal access token {string} is rejected by {string}",
  async function (this: DomusWorld, _name: string, apiPath: string) {
    const url = new URL(apiPath, this.page.url()).toString();
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${lastCreatedPat}` },
    });
    assert.equal(
      res.status,
      401,
      `Expected 401 from ${url}, got ${res.status}`,
    );
  },
);
