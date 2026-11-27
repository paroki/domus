import assert from "node:assert/strict";
import test from "node:test";
import {
  DEFAULT_PAROKI_TIMEZONE,
  formatParokiTimezone,
  GEREJAWI_STEPS,
  getLookupCategoryLabel,
  hasAnyRefAdminRole,
  isParokiAdmin,
  isSuperAdmin,
  LOOKUP_CATEGORIES,
  LOOKUP_CATEGORY_META,
  NIL_UUID,
  PAROKI_STEPS,
  PAROKI_TIMEZONES,
  ROLE_PAROKI_ADMIN,
  ROLE_SUPER_ADMIN,
  SIPIL_STEPS,
  slugifyCode,
  validateCoordinatePair,
} from "./model";

test("role helpers detect user:super-admin and ref:paroki-admin", () => {
  assert.equal(isSuperAdmin([ROLE_SUPER_ADMIN]), true);
  assert.equal(isParokiAdmin([ROLE_PAROKI_ADMIN]), true);
  assert.equal(hasAnyRefAdminRole(["user:user"]), false);
  assert.equal(hasAnyRefAdminRole([ROLE_PAROKI_ADMIN]), true);
});

test("LOOKUP_CATEGORIES contains the 4 non-umat Bahasa Indonesia categories", () => {
  assert.deepEqual(LOOKUP_CATEGORIES, [
    "status_paroki",
    "org_position",
    "event_type",
    "liturgi_song_part",
  ]);
  assert.equal(LOOKUP_CATEGORY_META.length, 4);
  assert.equal(getLookupCategoryLabel("status_paroki"), "Status Paroki");
  assert.equal(getLookupCategoryLabel("event_type"), "Jenis Kegiatan");
  assert.equal(
    getLookupCategoryLabel("liturgi_song_part"),
    "Bagian Lagu Liturgi",
  );
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
    ],
  );
  assert.match(PAROKI_STEPS[0]?.description ?? "", /\/users/);
  assert.doesNotMatch(GEREJAWI_STEPS[2]?.description ?? "", /tunjuk Admin/i);
});

test("slugifyCode normalizes mixed case and spaces into valid code slugs", () => {
  assert.equal(slugifyCode("  KAJ Paroki 01! "), "kaj-paroki-01");
  assert.equal(
    slugifyCode("Keuskupan Agung Jakarta"),
    "keuskupan-agung-jakarta",
  );
  assert.equal(slugifyCode("---abc_123---"), "abc_123");
});

test("SIPIL_STEPS, timezone helpers, and validateCoordinatePair work", () => {
  assert.deepEqual(
    SIPIL_STEPS.map((s) => s.path),
    [
      "/settings/sipil/kota",
      "/settings/sipil/kecamatan",
      "/settings/sipil/kelurahan",
    ],
  );
  assert.equal(NIL_UUID, "00000000-0000-0000-0000-000000000000");
  assert.equal(DEFAULT_PAROKI_TIMEZONE, "Asia/Makassar");
  assert.equal(PAROKI_TIMEZONES.length, 3);
  assert.equal(formatParokiTimezone("Asia/Jakarta"), "WIB");
  assert.equal(formatParokiTimezone("Asia/Makassar"), "WITA");
  assert.equal(formatParokiTimezone("Asia/Jayapura"), "WIT");
  assert.equal(formatParokiTimezone("UTC"), "UTC");

  assert.deepEqual(validateCoordinatePair("", ""), { valid: true });
  assert.equal(validateCoordinatePair("-6.2", "").valid, false);
  assert.equal(validateCoordinatePair("91", "106").valid, false);
  assert.equal(validateCoordinatePair("-6.2", "181").valid, false);
  assert.deepEqual(validateCoordinatePair("-6.2", "106.8"), {
    valid: true,
    latitude: -6.2,
    longitude: 106.8,
  });
});
