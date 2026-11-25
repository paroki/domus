import assert from "node:assert/strict";
import test from "node:test";
import {
  GEREJAWI_STEPS,
  getLookupCategoryLabel,
  hasAnyRefAdminRole,
  isParokiAdmin,
  isSuperAdmin,
  LOOKUP_CATEGORIES,
  LOOKUP_CATEGORY_META,
  PAROKI_STEPS,
  ROLE_PAROKI_ADMIN,
  ROLE_SUPER_ADMIN,
  slugifyCode,
} from "./model";

test("role helpers detect user:super-admin and ref:paroki-admin", () => {
  assert.equal(isSuperAdmin([ROLE_SUPER_ADMIN]), true);
  assert.equal(isParokiAdmin([ROLE_PAROKI_ADMIN]), true);
  assert.equal(hasAnyRefAdminRole(["user:user"]), false);
  assert.equal(hasAnyRefAdminRole([ROLE_PAROKI_ADMIN]), true);
});

test("LOOKUP_CATEGORIES contains all 14 Bahasa Indonesia categories", () => {
  assert.deepEqual(LOOKUP_CATEGORIES, [
    "status_paroki",
    "agama",
    "golongan_darah",
    "suku",
    "hubungan_keluarga",
    "status_pernikahan",
    "status_rumah",
    "pendidikan",
    "pekerjaan",
    "profesi",
    "disabilitas",
    "status_ekonomi",
    "aktivitas_sosial",
    "org_position",
  ]);
  assert.equal(LOOKUP_CATEGORY_META.length, 14);
  assert.equal(getLookupCategoryLabel("status_paroki"), "Status Paroki");
});

test("GEREJAWI_STEPS and PAROKI_STEPS point to /settings/* IUI routes", () => {
  assert.deepEqual(
    GEREJAWI_STEPS.map((s) => s.path),
    [
      "/settings/gerejawi/keuskupan",
      "/settings/gerejawi/kevikepan",
      "/settings/gerejawi/paroki",
    ],
  );
  assert.deepEqual(
    PAROKI_STEPS.map((s) => s.path),
    [
      "/settings/paroki/profile",
      "/settings/paroki/wilayah",
      "/settings/paroki/lingkungan",
      "/settings/paroki/periods",
    ],
  );
});

test("slugifyCode normalizes mixed case and spaces into valid code slugs", () => {
  assert.equal(slugifyCode("  KAJ Paroki 01! "), "kaj-paroki-01");
});
