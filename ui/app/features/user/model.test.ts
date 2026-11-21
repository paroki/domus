import assert from "node:assert/strict";
import test from "node:test";
import type { User } from "./api/types";
import {
  buildOnboardingSteps,
  canManageUsers,
  formatUserAgent,
  getAvailableUserScopes,
  getBirthDateBounds,
  isOrgAdmin,
  isParishAdmin,
  isSuperAdmin,
  providerLabel,
  ROLE_ORG_ADMIN,
  ROLE_PARISH_ADMIN,
  ROLE_SUPER_ADMIN,
  ROLE_USER,
  sanitizeRedirect,
  shouldRedirectToOnboarding,
  USER_GENDERS,
} from "./model";

test("formatUserAgent detects common browsers and OS", () => {
  assert.equal(
    formatUserAgent(
      "Mozilla/5.0 (X11; Linux x86_64; rv:130.0) Gecko/20100101 Firefox/130.0",
    ),
    "Firefox · Linux",
  );
  assert.equal(
    formatUserAgent(
      "Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 Chrome/128.0.0.0 Mobile Safari/537.36",
    ),
    "Chrome · Android",
  );
  assert.equal(formatUserAgent(""), "Perangkat tidak dikenal");
});

test("providerLabel maps known providers", () => {
  assert.equal(providerLabel("email"), "Email OTP");
  assert.equal(providerLabel("google"), "Google");
  assert.equal(providerLabel("github"), "GitHub");
  assert.equal(providerLabel("telegram"), "Telegram");
  assert.equal(providerLabel("whatsapp"), "WhatsApp");
  assert.equal(providerLabel("custom"), "custom");
});

test("isSuperAdmin checks user:super-admin role", () => {
  const admin = { roles: [ROLE_SUPER_ADMIN, ROLE_USER] } as User;
  const member = { roles: [ROLE_USER] } as User;
  assert.equal(isSuperAdmin(admin), true);
  assert.equal(isSuperAdmin(member), false);
  assert.equal(isSuperAdmin(null), false);
});

test("role helpers and getAvailableUserScopes resolve directory scopes", () => {
  const superUser = { roles: [ROLE_SUPER_ADMIN, ROLE_USER] } as User;
  const parishUser = { roles: [ROLE_PARISH_ADMIN, ROLE_USER] } as User;
  const orgUser = { roles: [ROLE_ORG_ADMIN, ROLE_USER] } as User;
  const plainUser = { roles: [ROLE_USER] } as User;

  assert.equal(isParishAdmin(parishUser), true);
  assert.equal(isOrgAdmin(orgUser), true);
  assert.equal(canManageUsers(superUser), true);
  assert.equal(canManageUsers(parishUser), true);
  assert.equal(canManageUsers(orgUser), true);
  assert.equal(canManageUsers(plainUser), false);

  assert.deepEqual(getAvailableUserScopes(superUser.roles), [
    "global",
    "parish",
  ]);
  assert.deepEqual(getAvailableUserScopes(parishUser.roles), ["parish", "org"]);
  assert.deepEqual(getAvailableUserScopes(orgUser.roles), ["org"]);
  assert.deepEqual(getAvailableUserScopes([ROLE_SUPER_ADMIN, ROLE_ORG_ADMIN]), [
    "global",
    "parish",
    "org",
  ]);
  assert.deepEqual(getAvailableUserScopes(plainUser.roles), []);
  assert.deepEqual(getAvailableUserScopes(null), []);
});

test("sanitizeRedirect rejects protocol-relative and external URLs", () => {
  assert.equal(sanitizeRedirect("/me"), "/me");
  assert.equal(sanitizeRedirect("/users?q=a"), "/users?q=a");
  assert.equal(sanitizeRedirect("//evil.example.com"), "/");
  assert.equal(sanitizeRedirect("https://evil.example.com"), "/");
  assert.equal(sanitizeRedirect(null), "/");
});

test("buildOnboardingSteps builds dynamic steps from otpChannels", () => {
  assert.deepEqual(buildOnboardingSteps(["email"]), ["biodata"]);
  assert.deepEqual(buildOnboardingSteps(["email", "whatsapp"]), [
    "biodata",
    "whatsapp",
  ]);
  assert.deepEqual(buildOnboardingSteps(["email", "telegram"]), [
    "biodata",
    "telegram",
  ]);
  assert.deepEqual(buildOnboardingSteps(["email", "whatsapp", "telegram"]), [
    "biodata",
    "whatsapp",
    "telegram",
  ]);
});

test("shouldRedirectToOnboarding avoids /onboarding loop and handles states", () => {
  assert.equal(shouldRedirectToOnboarding(false, "/onboarding"), false);
  assert.equal(shouldRedirectToOnboarding(false, "/"), true);
  assert.equal(shouldRedirectToOnboarding(false, "/me"), true);
  assert.equal(shouldRedirectToOnboarding(true, "/"), false);
  assert.equal(shouldRedirectToOnboarding(true, "/me"), false);
  assert.equal(shouldRedirectToOnboarding(undefined, "/"), false);
  assert.equal(shouldRedirectToOnboarding(undefined, "/me"), false);
});

test("USER_GENDERS and getBirthDateBounds define biodata constraints", () => {
  assert.deepEqual(USER_GENDERS, ["pria", "wanita"]);
  assert.deepEqual(getBirthDateBounds(new Date("2026-11-21T15:30:00Z")), {
    min: "1906-11-21",
    max: "2013-11-21",
  });
});
