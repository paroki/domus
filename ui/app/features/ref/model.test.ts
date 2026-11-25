import assert from "node:assert/strict";
import test from "node:test";
import {
  getVisibleSettingsItems,
  ROLE_PAROKI_ADMIN,
  ROLE_SUPER_ADMIN,
  slugifyCode,
} from "./model";

test("getVisibleSettingsItems returns 0 items for regular user", () => {
  assert.deepEqual(getVisibleSettingsItems(["user:user"]), []);
  assert.deepEqual(getVisibleSettingsItems([]), []);
});

test("getVisibleSettingsItems returns only lookups and paroki for super-admin", () => {
  const ids = getVisibleSettingsItems([ROLE_SUPER_ADMIN]).map((i) => i.id);
  assert.deepEqual(ids, ["lookups", "paroki"]);
});

test("getVisibleSettingsItems returns 4 paroki items for paroki-admin", () => {
  const ids = getVisibleSettingsItems([ROLE_PAROKI_ADMIN]).map((i) => i.id);
  assert.deepEqual(ids, ["profile", "wilayah", "lookups", "periods"]);
});

test("getVisibleSettingsItems returns all 5 items when holding both roles", () => {
  const ids = getVisibleSettingsItems([
    ROLE_SUPER_ADMIN,
    ROLE_PAROKI_ADMIN,
  ]).map((i) => i.id);
  assert.deepEqual(ids, ["profile", "wilayah", "lookups", "periods", "paroki"]);
});

test("getVisibleSettingsItems filters visible items by query string", () => {
  const filtered = getVisibleSettingsItems(
    [ROLE_SUPER_ADMIN, ROLE_PAROKI_ADMIN],
    "wilayah",
  );
  assert.deepEqual(
    filtered.map((i) => i.id),
    ["wilayah"],
  );
});

test("slugifyCode normalizes mixed case and spaces into valid code slugs", () => {
  assert.equal(slugifyCode("  KAJ Paroki 01! "), "kaj-paroki-01");
});
