import assert from "node:assert/strict";
import test from "node:test";
import {
  assertCompressedSize,
  computeTargetSize,
  DEFAULT_MAX_BYTES,
  ImageTooLargeError,
  validateImageMimeType,
} from "./image";

test("computeTargetSize keeps dimensions unchanged when both sides <= max", () => {
  assert.deepEqual(computeTargetSize(1200, 900), { width: 1200, height: 900 });
  assert.deepEqual(computeTargetSize(1600, 1600), {
    width: 1600,
    height: 1600,
  });
});

test("computeTargetSize downscales landscape and portrait while keeping aspect ratio", () => {
  assert.deepEqual(computeTargetSize(3200, 1600), { width: 1600, height: 800 });
  assert.deepEqual(computeTargetSize(1200, 2400), { width: 800, height: 1600 });
  assert.deepEqual(computeTargetSize(3000, 2001), {
    width: 1600,
    height: 1067,
  });
});

test("assertCompressedSize allows <= 2 MiB and throws ImageTooLargeError above", () => {
  assert.doesNotThrow(() => assertCompressedSize(DEFAULT_MAX_BYTES));
  assert.throws(
    () => assertCompressedSize(DEFAULT_MAX_BYTES + 1),
    (err: unknown) =>
      err instanceof ImageTooLargeError &&
      err.sizeBytes === DEFAULT_MAX_BYTES + 1 &&
      err.maxBytes === DEFAULT_MAX_BYTES,
  );
});

test("validateImageMimeType rejects non-image mime types", () => {
  assert.doesNotThrow(() => validateImageMimeType("image/jpeg"));
  assert.doesNotThrow(() => validateImageMimeType("image/webp"));
  assert.throws(() => validateImageMimeType("application/pdf"));
  assert.throws(() => validateImageMimeType("text/plain"));
});
