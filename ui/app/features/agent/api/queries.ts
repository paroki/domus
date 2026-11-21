import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useQuery,
} from "@tanstack/react-query";
import { api, unwrap, unwrapMeta } from "~/shared/api/client";
import { agentKeys } from "./keys";
import type {
  SearchAgentMemoriesRequest,
  SearchAgentMessagesRequest,
  SearchAgentSessionsRequest,
} from "./types";

export function agentSessionsInfiniteOptions(
  params: Omit<SearchAgentSessionsRequest, "cursor"> = {},
  enabled = true,
) {
  return infiniteQueryOptions({
    queryKey: agentKeys.sessions(params),
    initialPageParam: undefined as string | undefined,
    enabled,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/agent/sessions/search", {
          body: { ...params, cursor: pageParam, limit: params.limit ?? 20 },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useAgentSessions(
  params: Omit<SearchAgentSessionsRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(agentSessionsInfiniteOptions(params, enabled));
}

export function agentSessionOptions(sessionId: string, enabled = true) {
  return queryOptions({
    queryKey: agentKeys.session(sessionId),
    enabled: Boolean(sessionId) && enabled,
    queryFn: () =>
      unwrap(
        api.GET("/agent/sessions/{sessionId}", {
          params: { path: { sessionId } },
        }),
      ),
  });
}

export function useAgentSession(sessionId: string, enabled = true) {
  return useQuery(agentSessionOptions(sessionId, enabled));
}

export function agentMessagesInfiniteOptions(
  sessionId: string,
  params: Omit<SearchAgentMessagesRequest, "cursor"> = {},
  enabled = true,
) {
  return infiniteQueryOptions({
    queryKey: agentKeys.messages(sessionId, params),
    initialPageParam: undefined as string | undefined,
    enabled: Boolean(sessionId) && enabled,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/agent/sessions/{sessionId}/messages/search", {
          params: { path: { sessionId } },
          body: { ...params, cursor: pageParam, limit: params.limit ?? 50 },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useAgentMessages(
  sessionId: string,
  params: Omit<SearchAgentMessagesRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    agentMessagesInfiniteOptions(sessionId, params, enabled),
  );
}

export function agentMemoriesInfiniteOptions(
  params: Omit<SearchAgentMemoriesRequest, "cursor"> = {},
  enabled = true,
) {
  return infiniteQueryOptions({
    queryKey: agentKeys.memories(params),
    initialPageParam: undefined as string | undefined,
    enabled,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/agent/memories/search", {
          body: { ...params, cursor: pageParam, limit: params.limit ?? 50 },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useAgentMemories(
  params: Omit<SearchAgentMemoriesRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(agentMemoriesInfiniteOptions(params, enabled));
}
