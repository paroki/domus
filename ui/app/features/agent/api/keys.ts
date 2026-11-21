import type {
  SearchAgentMemoriesRequest,
  SearchAgentMessagesRequest,
  SearchAgentSessionsRequest,
} from "./types";

export const agentKeys = {
  all: ["agent"] as const,
  sessions: (params?: Omit<SearchAgentSessionsRequest, "cursor">) =>
    [...agentKeys.all, "sessions", params ?? {}] as const,
  session: (sessionId: string) =>
    [...agentKeys.all, "session", sessionId] as const,
  messages: (
    sessionId: string,
    params?: Omit<SearchAgentMessagesRequest, "cursor">,
  ) => [...agentKeys.all, "messages", sessionId, params ?? {}] as const,
  memories: (params?: Omit<SearchAgentMemoriesRequest, "cursor">) =>
    [...agentKeys.all, "memories", params ?? {}] as const,
};
