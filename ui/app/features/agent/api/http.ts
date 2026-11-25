import {
  type ApiEnvelope,
  ApiError,
  type ApiMeta,
  getActiveParishId,
} from "~/shared/api/client";
import { resolveErrorMessage } from "~/shared/api/errors";
import { AGENT_ERRORS } from "./errors";

export interface AgentFetchResult<T> {
  data: T;
  meta: ApiMeta | null;
}

export async function agentRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<AgentFetchResult<T>> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const parishId = getActiveParishId();
  if (parishId) {
    headers.set("X-Paroki-Id", parishId);
  }

  const response = await fetch(`/api${path}`, {
    ...init,
    credentials: "include",
    headers,
  });

  if (response.status === 204) {
    return { data: undefined as T, meta: null };
  }

  let envelope: ApiEnvelope<T> | undefined;
  try {
    envelope = (await response.json()) as ApiEnvelope<T>;
  } catch {
    envelope = undefined;
  }

  if (!response.ok || envelope?.error) {
    const errPayload = envelope?.error;
    const code = errPayload?.code ?? `HTTP_${response.status}`;
    const details = (errPayload?.details ?? []).map((d) => ({
      field: d.field ?? "",
      code: d.code ?? code,
      message: d.message ?? "",
    }));
    const msg = resolveErrorMessage(
      code,
      errPayload?.message ?? response.statusText,
      AGENT_ERRORS,
    );
    throw new ApiError(
      response.status,
      code,
      msg,
      errPayload?.requestId,
      details,
    );
  }

  return {
    data: envelope?.data as T,
    meta: envelope?.meta ?? null,
  };
}
