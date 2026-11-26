import assert from "node:assert/strict";
import test from "node:test";
import type { User } from "./api/types";
import {
  genderLabel,
  isUmatAdmin,
  NIL_UUID,
  optionalUuidForPatch,
  parseKeluargaDirectoryFilters,
  parseUmatDirectoryFilters,
  ROLE_UMAT_ADMIN,
  validatePIIInput,
} from "./model";

test("isUmatAdmin checks umat:admin role", () => {
  assert.equal(isUmatAdmin({ roles: [ROLE_UMAT_ADMIN] } as User), true);
  assert.equal(isUmatAdmin({ roles: ["ref:paroki-admin"] } as User), false);
  assert.equal(isUmatAdmin(null), false);
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
