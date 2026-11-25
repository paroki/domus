import assert from "node:assert/strict";
import test from "node:test";
import {
  ApiError,
  api,
  getActiveParokiId,
  setActiveParokiId,
  unwrap,
  unwrapMeta,
  unwrapVoid,
} from "./client";
import { createQueryClient } from "./query-client";

test("unwrap returns data on 200 envelope", async () => {
  const res = new Response(
    JSON.stringify({ data: { status: "ok" }, meta: null, error: null }),
    { status: 200, headers: { "Content-Type": "application/json" } },
  );

  const data = await unwrap(
    Promise.resolve({
      data: { data: { status: "ok" }, meta: null, error: null },
      response: res,
    }),
  );
  assert.deepEqual(data, { status: "ok" });
});

test("unwrapVoid resolves on 204 No Content and throws ApiError on failure", async () => {
  const res204 = new Response(null, { status: 204 });
  await unwrapVoid(
    Promise.resolve({
      data: undefined,
      response: res204,
    }),
  );

  const res403 = new Response("{}", { status: 403 });
  await assert.rejects(
    () =>
      unwrapVoid(
        Promise.resolve({
          error: {
            data: null,
            meta: null,
            error: { code: "REF_FORBIDDEN", message: "forbidden" },
          },
          response: res403,
        }),
        { REF_FORBIDDEN: "Akses ditolak" },
      ),
    (err: unknown) => {
      assert.ok(err instanceof ApiError);
      assert.equal(err.status, 403);
      assert.equal(err.code, "REF_FORBIDDEN");
      assert.equal(err.message, "Akses ditolak");
      return true;
    },
  );
});

test("unwrapMeta returns data and pagination meta", async () => {
  const res = new Response("{}", { status: 200 });
  const result = await unwrapMeta(
    Promise.resolve({
      data: {
        data: [{ id: "1" }],
        meta: { nextCursor: "cur-2", hasMore: true, limit: 10, total: 25 },
        error: null,
      },
      response: res,
    }),
  );

  assert.deepEqual(result.data, [{ id: "1" }]);
  assert.deepEqual(result.meta, {
    nextCursor: "cur-2",
    hasMore: true,
    limit: 10,
    total: 25,
  });
});

test("throws ApiError with structured details on 4xx and 5xx", async () => {
  const errBody = {
    data: null,
    meta: null,
    error: {
      code: "VALIDATION_ERROR",
      message: "One or more fields failed validation.",
      requestId: "req-123",
      details: [{ field: "email", code: "EMAIL", message: "invalid email" }],
    },
  };
  const res = new Response(JSON.stringify(errBody), { status: 422 });

  let caught: unknown;
  try {
    await unwrap(
      Promise.resolve({
        error: errBody,
        response: res,
      }),
    );
  } catch (err) {
    caught = err;
  }

  assert.ok(caught instanceof ApiError);
  assert.equal(caught.status, 422);
  assert.equal(caught.code, "VALIDATION_ERROR");
  assert.equal(caught.requestId, "req-123");
  assert.deepEqual(caught.details, [
    { field: "email", code: "EMAIL", message: "invalid email" },
  ]);

  // Feature-specific error override on 5xx
  const res500 = new Response("{}", { status: 500 });
  let caught500: unknown;
  try {
    await unwrap(
      Promise.resolve({
        error: {
          data: null,
          meta: null,
          error: { code: "CUSTOM_ERR", message: "raw" },
        },
        response: res500,
      }),
      { CUSTOM_ERR: "Friendly feature message" },
    );
  } catch (err) {
    caught500 = err;
  }
  assert.ok(caught500 instanceof ApiError);
  assert.equal(caught500.status, 500);
  assert.equal(caught500.message, "Friendly feature message");
});

test("injects X-Paroki-Id header when active paroki is set", async () => {
  const seenHeaders: (string | null)[] = [];
  const mockFetch = (async (input: RequestInfo | URL) => {
    const req = input as Request;
    seenHeaders.push(req.headers.get("X-Paroki-Id"));
    return new Response(
      JSON.stringify({ data: { status: "ok" }, meta: null, error: null }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    );
  }) as typeof fetch;

  try {
    setActiveParokiId("paroki-abc");
    assert.equal(getActiveParokiId(), "paroki-abc");
    await unwrap(
      api.GET("/health", { baseUrl: "http://localhost/api", fetch: mockFetch }),
    );

    setActiveParokiId(null);
    await unwrap(
      api.GET("/health", { baseUrl: "http://localhost/api", fetch: mockFetch }),
    );

    assert.deepEqual(seenHeaders, ["paroki-abc", null]);
  } finally {
    setActiveParokiId(null);
  }
});

test("createQueryClient suppresses retries on 4xx ApiError and retries 5xx up to 2 times", () => {
  const qc = createQueryClient();
  const retry = qc.getDefaultOptions().queries?.retry as (
    failureCount: number,
    error: unknown,
  ) => boolean;

  assert.equal(retry(0, new ApiError(404, "NOT_FOUND", "missing")), false);
  assert.equal(retry(0, new ApiError(422, "VALIDATION_ERROR", "bad")), false);
  assert.equal(retry(0, new ApiError(500, "INTERNAL_ERROR", "boom")), true);
  assert.equal(retry(1, new ApiError(500, "INTERNAL_ERROR", "boom")), true);
  assert.equal(retry(2, new ApiError(500, "INTERNAL_ERROR", "boom")), false);
});
