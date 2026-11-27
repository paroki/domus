import assert from "node:assert/strict";
import test from "node:test";
import {
  canManageLiturgiGlobal,
  DEFAULT_LITURGI_SONG_PARTS,
  formatLiturgiRange,
  formatParokiTimezoneShort,
  formatRespondedAt,
  formatSubstituteLabel,
  fromLocalDatetimeInputValue,
  getLiturgiMinistryMeta,
  groupPetugasByRole,
  hasSubstituteProposal,
  isUserAssignedToSchedule,
  LITURGI_MINISTRIES,
  LITURGI_ROLE_SLUGS,
  mergeCantusIntoLaguForm,
  petugasStatusMeta,
  ROLE_LITURGI_ADMIN,
  ROLE_PAROKI_ADMIN,
  resolveSongPartOptions,
  scheduleStatusMeta,
  songPartLabel,
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

test("scheduleStatusMeta and formatRespondedAt format status and timestamps", () => {
  assert.deepEqual(scheduleStatusMeta("scheduled"), {
    label: "Terjadwal",
    tone: "success",
  });
  assert.deepEqual(scheduleStatusMeta("cancelled"), {
    label: "Dibatalkan",
    tone: "danger",
  });
  const ts = formatRespondedAt("2026-12-24T11:00:00.000Z", "Asia/Makassar");
  assert.ok(ts.includes("WITA"));
  assert.equal(formatRespondedAt(""), "");
});

test("substitute helpers detect and format declined substitute proposals", () => {
  assert.equal(
    hasSubstituteProposal({ status: "confirmed", substituteName: "Budi" }),
    false,
  );
  assert.equal(
    hasSubstituteProposal({ status: "declined", substituteName: "Budi" }),
    true,
  );
  assert.equal(
    formatSubstituteLabel({ status: "declined", substituteName: "Sari" }),
    "Usulan Pengganti: Sari",
  );
  assert.equal(
    formatSubstituteLabel({ status: "declined", substituteOrgId: "org-1" }),
    "Usulan Pengganti: Organisasi / Kelompok",
  );
  assert.equal(formatSubstituteLabel({ status: "declined" }), null);
});

test("song part helpers resolve lookups/defaults and merge Cantus catalog auto-fill", () => {
  const defaults = resolveSongPartOptions();
  assert.equal(defaults.length, 11);
  assert.deepEqual(defaults[0], { value: "pembuka", label: "Pembuka" });
  assert.deepEqual(defaults[7], { value: "agnus-dei", label: "Agnus Dei" });

  assert.equal(songPartLabel("agnus-dei"), "Agnus Dei");
  assert.equal(
    songPartLabel("pembuka", [
      { code: "pembuka", label: "Lagu Pembukaan", active: true },
    ]),
    "Lagu Pembukaan",
  );

  const merged = mergeCantusIntoLaguForm(
    {
      cantusId: "",
      sumber: "",
      nomor: "",
      judul: "",
      nadaDasar: "",
      note: "Pakai bait 1-3",
    },
    {
      id: "cantus-1",
      sumber: "PS",
      nomor: "PS-320",
      judul: "Tuhan Kasihanilah Kami",
      nadaDasar: "D",
    },
  );
  assert.equal(merged.cantusId, "cantus-1");
  assert.equal(merged.sumber, "PS");
  assert.equal(merged.nomor, "PS-320");
  assert.equal(merged.judul, "Tuhan Kasihanilah Kami");
  assert.equal(merged.nadaDasar, "D");
  assert.equal(merged.note, "Pakai bait 1-3");
});
