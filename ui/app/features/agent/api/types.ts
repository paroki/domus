import type { MemoryCategory } from "../model";

export interface ToolCallRecord {
  id: string;
  name: string;
  arguments?: string;
}

export interface SessionResponse {
  id: string;
  parishId: string;
  userId: string;
  title: string;
  providerId?: string;
  modelId?: string;
  lastMessageAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface MessageResponse {
  id: string;
  sessionId: string;
  parishId: string;
  userId: string;
  agentName: string;
  role: "user" | "assistant" | "tool";
  content: string;
  toolCalls?: ToolCallRecord[];
  toolCallId?: string;
  toolName?: string;
  meta?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface SessionDetailResponse extends SessionResponse {
  effectiveProviderId?: string;
  effectiveModelId?: string;
  messages: MessageResponse[];
}

export interface UpdateSessionRequest {
  title?: string;
  providerId?: string;
  modelId?: string;
}

export interface MemoryResponse {
  id: string;
  parishId: string;
  userId: string;
  category: MemoryCategory;
  content: string;
  sourceSessionId?: string;
  score?: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateMemoryRequest {
  content: string;
  category?: MemoryCategory;
}

export interface UpdateMemoryRequest {
  content?: string;
  category?: MemoryCategory;
}

export interface ProviderModelItem {
  id?: string;
  name?: string;
  capabilities?: string[];
}

export interface AgentProviderResponse {
  id?: string;
  name?: string;
  provider?: string;
  models?: ProviderModelItem[];
}

export interface ChatRequest {
  sessionId?: string;
  message: string;
}
