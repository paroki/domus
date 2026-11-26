import assert from "node:assert/strict";
import test from "node:test";
import { ApiError } from "~/shared/api/client";
import { resolveEventError } from "./api/errors";
import {
  computeEventLifecycle,
  DEFAULT_EVENT_ROLES,
  DEFAULT_EVENT_TYPES,
  DEFAULT_PAROKI_TIMEZONE,
  eventTypeLabel,
  formatEventRange,
  formatParokiTimezoneShort,
  fromLocalDatetimeInputValue,
  lifecycleMeta,
  mergeEventTypeOptions,
  petugasStatusMeta,
  rsvpStatusMeta,
  toLocalDatetimeInputValue,
  visibilityMeta,
} from "./model";

test("computeEventLifecycle derives cancelled, completed, ongoing, and upcoming", () => {
  const now = Date.parse("2026-11-21T10:00:00Z");

  assert.equal(
    computeEventLifecycle(
      {
        status: "cancelled",
        startAt: "2026-11-21T09:00:00Z",
        endAt: "2026-11-21T11:00:00Z",
      },
      now,
    ),
    "cancelled",
  );
  assert.equal(
    computeEventLifecycle(
      {
        status: "scheduled",
        startAt: "2026-11-21T07:00:00Z",
        endAt: "2026-11-21T09:00:00Z",
      },
      now,
    ),
    "completed",
  );
  assert.equal(
    computeEventLifecycle(
      {
        status: "scheduled",
        startAt: "2026-11-21T09:00:00Z",
        endAt: "2026-11-21T11:00:00Z",
      },
      now,
    ),
    "ongoing",
  );
  assert.equal(
    computeEventLifecycle(
      {
        status: "scheduled",
        startAt: "2026-11-21T12:00:00Z",
        endAt: "2026-11-21T14:00:00Z",
      },
      now,
    ),
    "upcoming",
  );
});

test("metadata helpers return expected Indonesian labels and Badge tones", () => {
  assert.deepEqual(lifecycleMeta("upcoming"), {
    label: "Akan Datang",
    tone: "neutral",
  });
  assert.deepEqual(lifecycleMeta("ongoing"), {
    label: "Berlangsung",
    tone: "success",
  });
  assert.deepEqual(lifecycleMeta("completed"), {
    label: "Selesai",
    tone: "neutral",
  });
  assert.deepEqual(lifecycleMeta("cancelled"), {
    label: "Dibatalkan",
    tone: "danger",
  });

  assert.deepEqual(visibilityMeta("public"), {
    label: "Publik Paroki",
    tone: "success",
  });
  assert.deepEqual(visibilityMeta("private"), {
    label: "Internal Organisasi",
    tone: "warning",
  });

  assert.deepEqual(rsvpStatusMeta("going"), {
    label: "Hadir",
    tone: "success",
  });
  assert.deepEqual(rsvpStatusMeta("maybe"), {
    label: "Mungkin",
    tone: "warning",
  });
  assert.deepEqual(rsvpStatusMeta("not_going"), {
    label: "Tidak Hadir",
    tone: "danger",
  });
});

test("datetime-local conversion helpers round-trip valid ISO timestamps in paroki timezone", () => {
  const iso = "2026-11-21T08:30:00.000Z";
  assert.equal(DEFAULT_PAROKI_TIMEZONE, "Asia/Makassar");
  assert.equal(formatParokiTimezoneShort("Asia/Jakarta"), "WIB");
  assert.equal(formatParokiTimezoneShort("Asia/Makassar"), "WITA");
  assert.equal(formatParokiTimezoneShort("Asia/Jayapura"), "WIT");
  assert.equal(formatParokiTimezoneShort(undefined), "WITA");

  const localWita = toLocalDatetimeInputValue(iso, "Asia/Makassar");
  assert.equal(localWita, "2026-11-21T16:30");
  assert.equal(fromLocalDatetimeInputValue(localWita, "Asia/Makassar"), iso);

  const localWib = toLocalDatetimeInputValue(iso, "Asia/Jakarta");
  assert.equal(localWib, "2026-11-21T15:30");
  assert.equal(fromLocalDatetimeInputValue(localWib, "Asia/Jakarta"), iso);

  assert.equal(toLocalDatetimeInputValue(undefined), "");
  assert.equal(fromLocalDatetimeInputValue(""), "");

  const range = formatEventRange(
    "2026-12-01T00:00:00Z",
    "2026-12-01T01:30:00Z",
    "Asia/Makassar",
  );
  assert.match(range, /WITA$/);
  assert.match(range, /08[.:]00/);
});

test("resolveEventError maps known event error codes and falls back safely", () => {
  assert.equal(
    resolveEventError(new ApiError(409, "EVENT_RSVP_CLOSED", "closed")),
    "Konfirmasi kehadiran (RSVP) untuk kegiatan ini sudah ditutup.",
  );
  assert.equal(
    resolveEventError(new ApiError(403, "EVENT_FORBIDDEN", "forbidden")),
    "Anda tidak memiliki izin untuk mengakses atau mengelola kegiatan ini.",
  );
  assert.equal(resolveEventError(new Error("boom"), "Fallback"), "Fallback");
});

test("event type helpers merge built-in defaults and custom lookups", () => {
  assert.equal(DEFAULT_EVENT_TYPES.length, 8);
  const merged = mergeEventTypeOptions([
    { code: "misa", label: "Perayaan Ekaristi" },
    { code: "ziarah", label: "Ziarah Rohani" },
  ]);
  assert.equal(
    merged.find((o) => o.code === "misa")?.label,
    "Perayaan Ekaristi",
  );
  assert.equal(merged.find((o) => o.code === "ziarah")?.label, "Ziarah Rohani");
  assert.equal(eventTypeLabel("misa"), "Misa / Ekaristi");
  assert.equal(
    eventTypeLabel("ziarah", [{ code: "ziarah", label: "Ziarah Rohani" }]),
    "Ziarah Rohani",
  );
  assert.equal(eventTypeLabel("custom_code"), "custom_code");
});

test("petugas role defaults and status metadata helpers work", () => {
  assert.equal(DEFAULT_EVENT_ROLES.includes("Organis"), true);
  assert.deepEqual(petugasStatusMeta("assigned"), {
    label: "Ditugaskan",
    tone: "neutral",
  });
  assert.deepEqual(petugasStatusMeta("confirmed"), {
    label: "Siap Bertugas",
    tone: "success",
  });
  assert.deepEqual(petugasStatusMeta("declined"), {
    label: "Berhalangan",
    tone: "danger",
  });
});
