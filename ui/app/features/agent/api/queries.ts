import { queryOptions, useQuery } from "@tanstack/react-query";
import { api, unwrap } from "~/shared/api/client";
import { AGENT_ERRORS } from "./errors";
import { agentRequest } from "./http";
import { agentKeys } from "./keys";
import type {
  AgentProviderResponse,
  MemoryResponse,
  SessionDetailResponse,
  SessionResponse,
} from "./types";

export function agentSessionsQueryOptions(q?: string) {
  return queryOptions({
    queryKey: agentKeys.sessions(q),
    queryFn: async () => {
      const res = await agentRequest<SessionResponse[]>(
        "/agent/sessions/search",
        {
          method: "POST",
          body: JSON.stringify({ q: q || undefined, limit: 50 }),
        },
      );
      return res.data ?? [];
    },
  });
}

export function useAgentSessions(q?: string, enabled = true) {
  return useQuery({
    ...agentSessionsQueryOptions(q),
    enabled,
  });
}

export function agentSessionQueryOptions(sessionId: string) {
  return queryOptions({
    queryKey: agentKeys.session(sessionId),
    queryFn: async () => {
      const res = await agentRequest<SessionDetailResponse>(
        `/agent/sessions/${sessionId}`,
      );
      return res.data;
    },
  });
}

export function useAgentSession(sessionId?: string, enabled = true) {
  return useQuery({
    ...agentSessionQueryOptions(sessionId ?? ""),
    enabled: enabled && Boolean(sessionId),
  });
}

export function agentMemoriesQueryOptions(q?: string, category?: string) {
  return queryOptions({
    queryKey: agentKeys.memories(q, category),
    queryFn: async () => {
      const res = await agentRequest<MemoryResponse[]>(
        "/agent/memories/search",
        {
          method: "POST",
          body: JSON.stringify({
            q: q || undefined,
            filter: category ? { category } : undefined,
            limit: 50,
          }),
        },
      );
      return res.data ?? [];
    },
  });
}

export function useAgentMemories(
  q?: string,
  category?: string,
  enabled = true,
) {
  return useQuery({
    ...agentMemoriesQueryOptions(q, category),
    enabled,
  });
}

export function useAgentProviders(enabled = false) {
  return useQuery({
    queryKey: agentKeys.providers(),
    queryFn: async () => {
      const res = await unwrap(api.GET("/agent/providers"), AGENT_ERRORS);
      return (res ?? []) as AgentProviderResponse[];
    },
    enabled,
  });
}
