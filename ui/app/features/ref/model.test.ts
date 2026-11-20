import assert from "node:assert/strict";
import test from "node:test";
import {
  getVisibleSettingsItems,
  ROLE_PARISH_ADMIN,
  ROLE_SUPER_ADMIN,
  slugifyCode,
} from "./model";

test("getVisibleSettingsItems returns 0 items for regular user", () => {
  assert.deepEqual(getVisibleSettingsItems(["user:user"]), []);
  assert.deepEqual(getVisibleSettingsItems([]), []);
});

test("getVisibleSettingsItems returns only lookups and parishes for super-admin", () => {
  const ids = getVisibleSettingsItems([ROLE_SUPER_ADMIN]).map((i) => i.id);
  assert.deepEqual(ids, ["lookups", "parishes"]);
});

test("getVisibleSettingsItems returns 4 parish items for parish-admin", () => {
  const ids = getVisibleSettingsItems([ROLE_PARISH_ADMIN]).map((i) => i.id);
  assert.deepEqual(ids, ["profile", "territory", "lookups", "periods"]);
});

test("getVisibleSettingsItems returns all 5 items when holding both roles", () => {
  const ids = getVisibleSettingsItems([
    ROLE_SUPER_ADMIN,
    ROLE_PARISH_ADMIN,
  ]).map((i) => i.id);
  assert.deepEqual(ids, [
    "profile",
    "territory",
    "lookups",
    "periods",
    "parishes",
  ]);
});

test("getVisibleSettingsItems filters visible items by query string", () => {
  const filtered = getVisibleSettingsItems(
    [ROLE_SUPER_ADMIN, ROLE_PARISH_ADMIN],
    "wilayah",
  );
  assert.deepEqual(
    filtered.map((i) => i.id),
    ["territory"],
  );
});

test("slugifyCode normalizes mixed case and spaces into valid code slugs", () => {
  assert.equal(slugifyCode("  KAJ Paroki 01! "), "kaj-paroki-01");
});
