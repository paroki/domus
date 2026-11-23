import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  ApiError,
  api,
  getActiveParishId,
  unwrap,
  unwrapVoid,
} from "~/shared/api/client";
import { resolveErrorMessage } from "~/shared/api/errors";
import { parseSSEBuffer } from "../model";
import { AGENT_ERRORS } from "./errors";
import { agentKeys } from "./keys";
import type {
  AgentChatRequest,
  AgentSSEEvent,
  CreateAgentMemoryRequest,
  CreateAgentSessionRequest,
  UpdateAgentMemoryRequest,
  UpdateAgentSessionRequest,
} from "./types";

export function useCreateAgentSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateAgentSessionRequest = {}) =>
      unwrap(api.POST("/agent/sessions", { body }), AGENT_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: agentKeys.all }),
  });
}

export function useUpdateAgentSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      sessionId,
      body,
    }: {
      sessionId: string;
      body: UpdateAgentSessionRequest;
    }) =>
      unwrap(
        api.PATCH("/agent/sessions/{sessionId}", {
          params: { path: { sessionId } },
          body,
        }),
        AGENT_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: agentKeys.all }),
  });
}

export function useDeleteAgentSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (sessionId: string) =>
      unwrapVoid(
        api.DELETE("/agent/sessions/{sessionId}", {
          params: { path: { sessionId } },
        }),
        AGENT_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: agentKeys.all }),
  });
}

export function useCreateAgentMemory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateAgentMemoryRequest) =>
      unwrap(api.POST("/agent/memories", { body }), AGENT_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: agentKeys.memories() }),
  });
}

export function useUpdateAgentMemory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      memoryId,
      body,
    }: {
      memoryId: string;
      body: UpdateAgentMemoryRequest;
    }) =>
      unwrap(
        api.PATCH("/agent/memories/{memoryId}", {
          params: { path: { memoryId } },
          body,
        }),
        AGENT_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: agentKeys.memories() }),
  });
}

export function useDeleteAgentMemory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (memoryId: string) =>
      unwrapVoid(
        api.DELETE("/agent/memories/{memoryId}", {
          params: { path: { memoryId } },
        }),
        AGENT_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: agentKeys.memories() }),
  });
}

export async function streamAgentChat({
  body,
  onEvent,
  signal,
}: {
  body: AgentChatRequest;
  onEvent: (ev: AgentSSEEvent) => void;
  signal?: AbortSignal;
}): Promise<void> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    Accept: "text/event-stream",
  };
  const activeParishId = getActiveParishId();
  if (activeParishId) {
    headers["X-Parish-Id"] = activeParishId;
  }

  const response = await fetch("/api/agent/chat", {
    method: "POST",
    credentials: "same-origin",
    headers,
    body: JSON.stringify(body),
    signal,
  });

  if (!response.ok) {
    let code = "INTERNAL_ERROR";
    let message = response.statusText || "Terjadi kesalahan.";
    try {
      const json = (await response.json()) as {
        error?: { code?: string; message?: string };
      };
      if (json?.error?.code) {
        code = json.error.code;
      }
      if (json?.error?.message) {
        message = json.error.message;
      }
    } catch {
      // Ignore non-JSON error body.
    }
    throw new ApiError(
      response.status,
      code,
      resolveErrorMessage(code, message, AGENT_ERRORS),
    );
  }

  if (!response.body) {
    return;
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let remainder = "";

  const emitParsedEvents = (events: AgentSSEEvent[]) => {
    for (const ev of events) {
      onEvent(ev);
      if (ev.event === "error") {
        const code = ev.code ?? "INTERNAL_ERROR";
        const message = ev.message?.includes("provider HTTP")
          ? ev.message
          : resolveErrorMessage(code, ev.message, AGENT_ERRORS);
        throw new ApiError(500, code, message);
      }
    }
  };

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    const chunk = decoder.decode(value, { stream: true });
    const parsed = parseSSEBuffer(remainder + chunk);
    remainder = parsed.remainder;
    emitParsedEvents(parsed.events);
  }

  const tail = `${remainder}${decoder.decode()}`;
  if (tail.trim()) {
    const flushed = parseSSEBuffer(`${tail}\n\n`);
    emitParsedEvents(flushed.events);
  }
}
