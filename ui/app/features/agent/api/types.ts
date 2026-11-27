import type { MemoryCategory } from "../model";

export interface ToolCallRecord {
  id: string;
  name: string;
  arguments?: string;
}

export interface SessionResponse {
  id: string;
  parokiId: string;
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
  parokiId: string;
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
  parokiId: string;
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

export interface QuestionOption {
  label: string;
  value: string;
}

export interface QuestionPayload {
  question: string;
  options: QuestionOption[];
  multiSelect?: boolean;
  allowCustom?: boolean;
}

export interface NavigationPayload {
  path: string;
  label: string;
  description?: string;
  autoRedirect?: boolean;
}

export interface PageContextInput {
  path: string;
  title?: string;
}

export interface ChatRequest {
  sessionId?: string;
  message: string;
  retry?: boolean;
  pageContext?: PageContextInput;
}
