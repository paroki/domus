import assert from "node:assert/strict";
import test from "node:test";
import {
  canManageLiturgiGlobal,
  DEFAULT_LITURGI_SONG_PARTS,
  formatLiturgiRange,
  formatParokiTimezoneShort,
  fromLocalDatetimeInputValue,
  getLiturgiMinistryMeta,
  groupPetugasByRole,
  isUserAssignedToSchedule,
  LITURGI_MINISTRIES,
  LITURGI_ROLE_SLUGS,
  petugasStatusMeta,
  ROLE_LITURGI_ADMIN,
  ROLE_PAROKI_ADMIN,
  toLocalDatetimeInputValue,
} from "./model";

test("LITURGI_MINISTRIES defines 6 canonical roles in fixed order", () => {
  assert.deepEqual(LITURGI_ROLE_SLUGS, [
    "lektor",
    "mazmur",
    "pembagi-komuni",
    "misdinar",
    "organis",
    "kor",
  ]);
  assert.equal(LITURGI_MINISTRIES.length, 6);
  assert.equal(getLiturgiMinistryMeta("mazmur")?.label, "Pemazmur");
  assert.equal(getLiturgiMinistryMeta("unknown"), undefined);
});

test("DEFAULT_LITURGI_SONG_PARTS has 11 standard Mass song parts", () => {
  assert.equal(DEFAULT_LITURGI_SONG_PARTS.length, 11);
  assert.equal(DEFAULT_LITURGI_SONG_PARTS[0], "Pembuka");
  assert.equal(DEFAULT_LITURGI_SONG_PARTS[10], "Penutup");
});

test("canManageLiturgiGlobal recognizes liturgi:admin and ref:paroki-admin", () => {
  assert.equal(canManageLiturgiGlobal([ROLE_LITURGI_ADMIN]), true);
  assert.equal(canManageLiturgiGlobal([ROLE_PAROKI_ADMIN]), true);
  assert.equal(canManageLiturgiGlobal(["user:user"]), false);
  assert.equal(canManageLiturgiGlobal(null), false);
});

test("petugasStatusMeta maps status codes to Indonesian labels and tones", () => {
  assert.deepEqual(petugasStatusMeta("confirmed"), {
    label: "Siap",
    tone: "success",
  });
  assert.deepEqual(petugasStatusMeta("declined"), {
    label: "Berhalangan",
    tone: "danger",
  });
  assert.deepEqual(petugasStatusMeta("assigned"), {
    label: "Ditugaskan",
    tone: "neutral",
  });
});

test("groupPetugasByRole and isUserAssignedToSchedule group and detect assignments", () => {
  const list = [
    { role: "lektor", userId: "u1", name: "Budi" },
    { role: "kor", userId: "u2", name: "Paduan Suara" },
    { role: "lektor", userId: "u3", name: "Sari" },
  ];
  const grouped = groupPetugasByRole(list);
  assert.equal(grouped.lektor.length, 2);
  assert.equal(grouped.kor.length, 1);
  assert.equal(grouped.misdinar.length, 0);
  assert.equal(isUserAssignedToSchedule(list, "u2"), true);
  assert.equal(isUserAssignedToSchedule(list, "u99"), false);
});

test("paroki timezone conversions and range formatting work for WITA", () => {
  assert.equal(formatParokiTimezoneShort("Asia/Jakarta"), "WIB");
  assert.equal(formatParokiTimezoneShort("Asia/Makassar"), "WITA");
  assert.equal(formatParokiTimezoneShort("Asia/Jayapura"), "WIT");

  const iso = fromLocalDatetimeInputValue("2026-12-24T19:00", "Asia/Makassar");
  assert.equal(iso, "2026-12-24T11:00:00.000Z");
  assert.equal(
    toLocalDatetimeInputValue(iso, "Asia/Makassar"),
    "2026-12-24T19:00",
  );

  const formatted = formatLiturgiRange(
    "2026-12-24T11:00:00.000Z",
    "2026-12-24T13:00:00.000Z",
    "Asia/Makassar",
  );
  assert.ok(formatted.includes("WITA"));
});
