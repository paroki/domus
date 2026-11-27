import assert from "node:assert/strict";
import test from "node:test";
import type { User } from "./api/types";
import {
  canEditUmatReferensi,
  genderLabel,
  getUmatReferensiCategoryLabel,
  isSuperAdmin,
  isUmatAdmin,
  NIL_UUID,
  optionalUuidForPatch,
  parseKeluargaDirectoryFilters,
  parseUmatDirectoryFilters,
  ROLE_SUPER_ADMIN,
  ROLE_UMAT_ADMIN,
  slugifyReferensiCode,
  UMAT_REFERENSI_CATEGORIES,
  validatePIIInput,
} from "./model";

test("isUmatAdmin checks umat:admin role", () => {
  assert.equal(isUmatAdmin({ roles: [ROLE_UMAT_ADMIN] } as User), true);
  assert.equal(isUmatAdmin({ roles: ["ref:paroki-admin"] } as User), false);
  assert.equal(isUmatAdmin(null), false);
});

test("isSuperAdmin and canEditUmatReferensi enforce parish vs global permissions", () => {
  const superAdminUser = { roles: [ROLE_SUPER_ADMIN] } as User;
  const umatAdminUser = { roles: [ROLE_UMAT_ADMIN] } as User;
  const plainUser = { roles: ["user:user"] } as User;

  assert.equal(isSuperAdmin(superAdminUser), true);
  assert.equal(isSuperAdmin(umatAdminUser), false);
  assert.equal(isSuperAdmin(null), false);

  assert.equal(canEditUmatReferensi(umatAdminUser, "paroki-1"), true);
  assert.equal(canEditUmatReferensi(umatAdminUser, undefined), false);
  assert.equal(canEditUmatReferensi(superAdminUser, undefined), true);
  assert.equal(canEditUmatReferensi(superAdminUser, "paroki-1"), false);
  assert.equal(canEditUmatReferensi(plainUser, "paroki-1"), false);
  assert.equal(canEditUmatReferensi(plainUser, undefined), false);
});

test("UMAT_REFERENSI_CATEGORIES has all 12 Data Umat categories and helpers", () => {
  assert.deepEqual(
    UMAT_REFERENSI_CATEGORIES.map((c) => c.slug),
    [
      "agama",
      "golongan_darah",
      "suku",
      "hubungan_keluarga",
      "status_pernikahan",
      "status_rumah",
      "status_ekonomi",
      "pendidikan",
      "pekerjaan",
      "profesi",
      "disabilitas",
      "aktivitas_sosial",
    ],
  );
  assert.equal(
    getUmatReferensiCategoryLabel("golongan_darah"),
    "Golongan Darah",
  );
  assert.equal(getUmatReferensiCategoryLabel("unknown_slug"), "unknown_slug");
  assert.equal(
    slugifyReferensiCode("  S1 Teknik Informatika! "),
    "s1-teknik-informatika",
  );
});

test("optionalUuidForPatch returns NIL_UUID for empty input", () => {
  assert.equal(optionalUuidForPatch(""), NIL_UUID);
  assert.equal(optionalUuidForPatch("   "), NIL_UUID);
  assert.equal(optionalUuidForPatch("abc-123"), "abc-123");
});

test("validatePIIInput enforces 4..32 runes without control chars", () => {
  assert.equal(validatePIIInput("", "NIK"), null);
  assert.equal(validatePIIInput("  1234  ", "NIK"), null);
  assert.ok(validatePIIInput("123", "NIK"));
  assert.ok(validatePIIInput("a".repeat(33), "NIK"));
  assert.ok(validatePIIInput("1234\n5678", "NIK"));
});

test("genderLabel formats pria and wanita", () => {
  assert.equal(genderLabel("pria"), "Pria");
  assert.equal(genderLabel("wanita"), "Wanita");
});

test("parseUmatDirectoryFilters and parseKeluargaDirectoryFilters parse URLSearchParams", () => {
  const umatParams = new URLSearchParams(
    "q=Budi&nik=3174&jenisKelamin=pria&almarhum=alive&wilayahId=w1",
  );
  assert.deepEqual(parseUmatDirectoryFilters(umatParams), {
    q: "Budi",
    nik: "3174",
    wilayahId: "w1",
    lingkunganId: "",
    jenisKelamin: "pria",
    agama: "",
    statusPernikahan: "",
    almarhum: "alive",
  });

  const kkParams = new URLSearchParams("q=Santoso&nomorKk=9999&aktif=false");
  assert.deepEqual(parseKeluargaDirectoryFilters(kkParams), {
    q: "Santoso",
    nomorKk: "9999",
    wilayahId: "",
    lingkunganId: "",
    aktif: "false",
  });
});
