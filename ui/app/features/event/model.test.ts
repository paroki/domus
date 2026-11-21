import assert from "node:assert/strict";
import test from "node:test";
import { ApiError } from "~/shared/api/client";
import { resolveEventError } from "./api/errors";
import {
  computeEventLifecycle,
  fromLocalDatetimeInputValue,
  lifecycleMeta,
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

test("datetime-local conversion helpers round-trip valid ISO timestamps", () => {
  const iso = "2026-11-21T08:30:00.000Z";
  const local = toLocalDatetimeInputValue(iso);
  assert.match(local, /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/);
  assert.equal(fromLocalDatetimeInputValue(local), iso);
  assert.equal(toLocalDatetimeInputValue(undefined), "");
  assert.equal(fromLocalDatetimeInputValue(""), "");
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
