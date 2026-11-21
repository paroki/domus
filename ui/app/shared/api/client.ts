import createClient from "openapi-fetch";
import { resolveErrorMessage } from "./errors";
import type { paths } from "./schema";

export interface ApiMeta {
  nextCursor?: string;
  hasMore?: boolean;
  limit?: number;
  total?: number;
}

export interface FieldError {
  field: string;
  code: string;
  message: string;
}

export interface ApiErrorPayload {
  code?: string;
  message?: string;
  requestId?: string;
  details?: Partial<FieldError>[];
}

export interface ApiEnvelope<T> {
  data?: T;
  meta?: ApiMeta | null;
  error?: ApiErrorPayload | null;
}

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId?: string;
  readonly details: FieldError[];

  constructor(
    status: number,
    code: string,
    message: string,
    requestId?: string,
    details: FieldError[] = [],
  ) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.requestId = requestId;
    this.details = details;
  }
}

let activeParishId: string | null = null;

export function setActiveParishId(parishId: string | null): void {
  activeParishId = parishId;
}

export function getActiveParishId(): string | null {
  return activeParishId;
}

export const setActiveParish = setActiveParishId;
export const getActiveParish = getActiveParishId;

export const api = createClient<paths>({ baseUrl: "/api" });

api.use({
  onRequest({ request }) {
    if (activeParishId) {
      request.headers.set("X-Parish-Id", activeParishId);
    }
    return request;
  },
});

type FetchResult<T> = {
  data?: ApiEnvelope<T>;
  error?: unknown;
  response: Response;
};

function extractErrorPayload(
  data: ApiEnvelope<unknown> | undefined,
  rawError: unknown,
): ApiErrorPayload | null {
  if (data?.error) {
    return data.error;
  }
  if (rawError && typeof rawError === "object") {
    const maybeEnvelope = rawError as Partial<ApiEnvelope<unknown>>;
    if (maybeEnvelope.error && typeof maybeEnvelope.error === "object") {
      return maybeEnvelope.error;
    }
    const maybePayload = rawError as Partial<ApiErrorPayload>;
    if (typeof maybePayload.code === "string") {
      return {
        code: maybePayload.code,
        message: maybePayload.message ?? "",
        requestId: maybePayload.requestId,
        details: maybePayload.details,
      };
    }
  }
  return null;
}

function toApiError(
  result: FetchResult<unknown>,
  featureErrors?: Record<string, string>,
): ApiError {
  const payload = extractErrorPayload(result.data, result.error);
  const code = payload?.code ?? "INTERNAL_ERROR";
  const message = resolveErrorMessage(
    code,
    payload?.message || result.response.statusText,
    featureErrors,
  );
  const details: FieldError[] = (payload?.details ?? []).map((d) => ({
    field: d.field ?? "",
    code: d.code ?? "",
    message: d.message ?? "",
  }));
  return new ApiError(
    result.response.status,
    code,
    message,
    payload?.requestId,
    details,
  );
}

export async function unwrapVoid(
  promise: Promise<{ data?: unknown; error?: unknown; response: Response }>,
  featureErrors?: Record<string, string>,
): Promise<void> {
  const result = await promise;
  const env =
    result.data && typeof result.data === "object"
      ? (result.data as ApiEnvelope<unknown>)
      : undefined;
  if (!result.response.ok || env?.error) {
    throw toApiError({ ...result, data: env }, featureErrors);
  }
}

export async function unwrap<T>(
  promise: Promise<FetchResult<T>>,
  featureErrors?: Record<string, string>,
): Promise<T> {
  const result = await promise;
  if (!result.response.ok || result.data?.error) {
    throw toApiError(result, featureErrors);
  }
  if (!result.data || result.data.data === undefined) {
    throw new ApiError(
      result.response.status,
      "INTERNAL_ERROR",
      "Empty response body.",
    );
  }
  return result.data.data;
}

export async function unwrapMeta<T>(
  promise: Promise<FetchResult<T>>,
  featureErrors?: Record<string, string>,
): Promise<{ data: T; meta: ApiMeta }> {
  const result = await promise;
  if (!result.response.ok || result.data?.error) {
    throw toApiError(result, featureErrors);
  }
  if (!result.data || result.data.data === undefined) {
    throw new ApiError(
      result.response.status,
      "INTERNAL_ERROR",
      "Empty response body.",
    );
  }
  return {
    data: result.data.data,
    meta: result.data.meta ?? { nextCursor: "", hasMore: false, limit: 20 },
  };
}
