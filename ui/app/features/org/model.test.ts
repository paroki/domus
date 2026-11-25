import assert from "node:assert/strict";
import test from "node:test";
import {
  buildInviteJoinUrl,
  computeInviteStatus,
  inviteStatusMeta,
  isParokiAdmin,
  kindLabel,
  ROLE_PAROKI_ADMIN,
  slugifyCode,
} from "./model";

test("isParokiAdmin checks ref:paroki-admin role", () => {
  assert.equal(isParokiAdmin([ROLE_PAROKI_ADMIN]), true);
  assert.equal(isParokiAdmin(["user:user"]), false);
  assert.equal(isParokiAdmin(undefined), false);
});

test("slugifyCode normalizes mixed case and punctuation", () => {
  assert.equal(slugifyCode("  Seksi Liturgi 01! "), "seksi-liturgi-01");
});

test("kindLabel maps known org kinds and falls back safely", () => {
  assert.equal(kindLabel("dpp"), "DPP");
  assert.equal(kindLabel("section"), "Bidang / Seksi");
  assert.equal(kindLabel("custom"), "custom");
});

test("inviteStatusMeta and computeInviteStatus reflect invite states", () => {
  const future = new Date(Date.now() + 3600_000).toISOString();
  const past = new Date(Date.now() - 3600_000).toISOString();

  assert.equal(
    computeInviteStatus({ expiresAt: future, usedCount: 0 }),
    "active",
  );
  assert.equal(
    computeInviteStatus({
      revokedAt: past,
      expiresAt: future,
      usedCount: 0,
    }),
    "revoked",
  );
  assert.equal(
    computeInviteStatus({ expiresAt: past, usedCount: 0 }),
    "expired",
  );
  assert.equal(
    computeInviteStatus({ expiresAt: future, maxUses: 2, usedCount: 2 }),
    "exhausted",
  );
  assert.equal(inviteStatusMeta("active").label, "Aktif");
  assert.equal(inviteStatusMeta("revoked").tone, "danger");
});

test("buildInviteJoinUrl formats shareable join URL", () => {
  assert.equal(
    buildInviteJoinUrl("https://domus.example.com/", "Abc123xyz"),
    "https://domus.example.com/org/join/Abc123xyz",
  );
});
