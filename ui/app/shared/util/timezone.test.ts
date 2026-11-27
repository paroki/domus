import assert from "node:assert/strict";
import test from "node:test";
import {
  DEFAULT_TIMEZONE,
  detectBrowserTimezone,
  getTimezoneOptions,
  INDONESIA_TIMEZONES,
} from "./timezone";

test("timezone utilities: default and Indonesian timezones", () => {
  assert.equal(DEFAULT_TIMEZONE, "Asia/Makassar");
  assert.deepEqual(
    INDONESIA_TIMEZONES.map((t) => t.value),
    ["Asia/Jakarta", "Asia/Makassar", "Asia/Jayapura"],
  );
});

test("detectBrowserTimezone falls back to DEFAULT_TIMEZONE for UTC or empty", () => {
  const orig = Intl.DateTimeFormat;
  try {
    for (const utcZone of ["UTC", "Etc/UTC", ""]) {
      Intl.DateTimeFormat = (() => ({
        resolvedOptions: () => ({ timeZone: utcZone }),
      })) as unknown as typeof Intl.DateTimeFormat;
      assert.equal(detectBrowserTimezone(), DEFAULT_TIMEZONE);
    }
    Intl.DateTimeFormat = (() => ({
      resolvedOptions: () => ({ timeZone: "Asia/Jakarta" }),
    })) as unknown as typeof Intl.DateTimeFormat;
    assert.equal(detectBrowserTimezone(), "Asia/Jakarta");
  } finally {
    Intl.DateTimeFormat = orig;
  }
});

test("getTimezoneOptions appends non-Indonesian valid IANA timezone without duplicating", () => {
  const base = getTimezoneOptions("Asia/Jakarta");
  assert.equal(base.length, 3);

  const custom = getTimezoneOptions("America/New_York");
  assert.equal(custom.length, 4);
  assert.deepEqual(custom[3], {
    value: "America/New_York",
    label: "America/New_York",
    short: "America/New_York",
  });

  const invalid = getTimezoneOptions("Invalid/Zone");
  assert.equal(
    invalid.some((t) => t.value === "Invalid/Zone"),
    false,
  );
});
