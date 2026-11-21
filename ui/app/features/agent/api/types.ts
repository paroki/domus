import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type AgentSessionResponse = Schemas["domain.AgentSessionResponse"];
export type SearchAgentSessionsRequest =
  Schemas["domain.SearchAgentSessionsRequest"];
export type AgentSessionSearchFilter =
  Schemas["domain.AgentSessionSearchFilter"];
export type CreateAgentSessionRequest =
  Schemas["domain.CreateAgentSessionRequest"];
export type UpdateAgentSessionRequest =
  Schemas["domain.UpdateAgentSessionRequest"];

export type AgentMessageResponse = Schemas["domain.AgentMessageResponse"];
export type AgentToolCallDTO = Schemas["domain.AgentToolCallDTO"];
export type SearchAgentMessagesRequest =
  Schemas["domain.SearchAgentMessagesRequest"];

export type AgentMemoryResponse = Schemas["domain.AgentMemoryResponse"];
export type SearchAgentMemoriesRequest =
  Schemas["domain.SearchAgentMemoriesRequest"];
export type AgentMemorySearchFilter = Schemas["domain.AgentMemorySearchFilter"];
export type CreateAgentMemoryRequest =
  Schemas["domain.CreateAgentMemoryRequest"];
export type UpdateAgentMemoryRequest =
  Schemas["domain.UpdateAgentMemoryRequest"];

export type AgentChatRequest = Schemas["domain.AgentChatRequest"];

export type AgentSSEEventType =
  | "session"
  | "thinking"
  | "tool_call"
  | "tool_result"
  | "delta"
  | "done"
  | "error";

export interface AgentSSEEvent {
  event: AgentSSEEventType;
  sessionId?: string;
  agentName?: string;
  content?: string;
  toolCallId?: string;
  toolName?: string;
  arguments?: string;
  code?: string;
  message?: string;
}
