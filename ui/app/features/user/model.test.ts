import assert from "node:assert/strict";
import test from "node:test";
import type { User } from "./api/types";
import {
  formatUserAgent,
  isSuperAdmin,
  providerLabel,
  ROLE_SUPER_ADMIN,
  sanitizeRedirect,
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
  const admin = { roles: [ROLE_SUPER_ADMIN, "user:user"] } as User;
  const member = { roles: ["user:user"] } as User;
  assert.equal(isSuperAdmin(admin), true);
  assert.equal(isSuperAdmin(member), false);
  assert.equal(isSuperAdmin(null), false);
});

test("sanitizeRedirect rejects protocol-relative and external URLs", () => {
  assert.equal(sanitizeRedirect("/me"), "/me");
  assert.equal(sanitizeRedirect("/users?q=a"), "/users?q=a");
  assert.equal(sanitizeRedirect("//evil.example.com"), "/");
  assert.equal(sanitizeRedirect("https://evil.example.com"), "/");
  assert.equal(sanitizeRedirect(null), "/");
});
