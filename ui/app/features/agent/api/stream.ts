import {
  type ApiEnvelope,
  ApiError,
  getActiveParishId,
} from "~/shared/api/client";
import { resolveErrorMessage } from "~/shared/api/errors";
import { createSSEParser, type SSEFrame } from "../model";
import { AGENT_ERRORS } from "./errors";
import type { ChatRequest } from "./types";

export interface StreamAgentChatParams {
  body: ChatRequest;
  signal: AbortSignal;
  onFrame: (frame: SSEFrame) => void;
}

export async function streamAgentChat({
  body,
  signal,
  onFrame,
}: StreamAgentChatParams): Promise<void> {
  const headers = new Headers({
    "Content-Type": "application/json",
    Accept: "text/event-stream",
  });
  const parishId = getActiveParishId();
  if (parishId) {
    headers.set("X-Paroki-Id", parishId);
  }

  const response = await fetch("/api/agent/chat", {
    method: "POST",
    credentials: "include",
    headers,
    body: JSON.stringify(body),
    signal,
  });

  if (!response.ok) {
    let envelope: ApiEnvelope<unknown> | undefined;
    try {
      envelope = (await response.json()) as ApiEnvelope<unknown>;
    } catch {
      envelope = undefined;
    }
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

  if (!response.body) {
    return;
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder("utf-8");
  const parser = createSSEParser();

  try {
    while (true) {
      const { value, done } = await reader.read();
      if (done) {
        const rest = decoder.decode();
        if (rest) {
          for (const frame of parser.push(rest)) {
            onFrame(frame);
          }
        }
        for (const frame of parser.flush()) {
          onFrame(frame);
        }
        break;
      }
      if (value) {
        const text = decoder.decode(value, { stream: true });
        for (const frame of parser.push(text)) {
          onFrame(frame);
        }
      }
    }
  } finally {
    reader.releaseLock();
  }
}
