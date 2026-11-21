import type { ProviderModelItem, SettingResponse } from "./api/types";

export const ROLE_SUPER_ADMIN = "user:super-admin";

export type SettingsTab = "channels" | "providers" | "agents";

export interface LLMProviderDef {
  id: "gemini" | "anthropic" | "openai" | "openrouter";
  label: string;
  defaultBaseURL: string;
  apiKeyKey: string;
  baseURLKey: string;
  modelsKey: string;
}

export const LLM_PROVIDERS: LLMProviderDef[] = [
  {
    id: "gemini",
    label: "Google Gemini",
    defaultBaseURL: "https://generativelanguage.googleapis.com",
    apiKeyKey: "agent.provider.gemini.api_key",
    baseURLKey: "agent.provider.gemini.base_url",
    modelsKey: "agent.provider.gemini.models",
  },
  {
    id: "anthropic",
    label: "Anthropic Claude",
    defaultBaseURL: "https://api.anthropic.com",
    apiKeyKey: "agent.provider.anthropic.api_key",
    baseURLKey: "agent.provider.anthropic.base_url",
    modelsKey: "agent.provider.anthropic.models",
  },
  {
    id: "openai",
    label: "OpenAI",
    defaultBaseURL: "https://api.openai.com/v1",
    apiKeyKey: "agent.provider.openai.api_key",
    baseURLKey: "agent.provider.openai.base_url",
    modelsKey: "agent.provider.openai.models",
  },
  {
    id: "openrouter",
    label: "OpenRouter",
    defaultBaseURL: "https://openrouter.ai/api/v1",
    apiKeyKey: "agent.provider.openrouter.api_key",
    baseURLKey: "agent.provider.openrouter.base_url",
    modelsKey: "agent.provider.openrouter.models",
  },
];

export interface AgentPersonaDef {
  id: "gabriel" | "mikael" | "rafael";
  name: string;
  role: string;
  tools: string[];
  providerKey: string;
  modelKey: string;
  effortKey: string;
  instructionKey: string;
}

export const AGENT_PERSONAS: AgentPersonaDef[] = [
  {
    id: "gabriel",
    name: "Gabriel",
    role: "Orkestrator & Memori Utama",
    tools: ["delegate_specialist", "memory_*"],
    providerKey: "agent.gabriel.provider",
    modelKey: "agent.gabriel.model",
    effortKey: "agent.gabriel.effort",
    instructionKey: "agent.gabriel.instruction",
  },
  {
    id: "mikael",
    name: "Mikael",
    role: "Spesialis Kegiatan & Jadwal",
    tools: ["event_*"],
    providerKey: "agent.mikael.provider",
    modelKey: "agent.mikael.model",
    effortKey: "agent.mikael.effort",
    instructionKey: "agent.mikael.instruction",
  },
  {
    id: "rafael",
    name: "Rafael",
    role: "Spesialis Organisasi & Komunitas",
    tools: ["org_*"],
    providerKey: "agent.rafael.provider",
    modelKey: "agent.rafael.model",
    effortKey: "agent.rafael.effort",
    instructionKey: "agent.rafael.instruction",
  },
];

export const REASONING_EFFORTS = [
  { value: "none", label: "Tanpa Penalaran (None)" },
  { value: "low", label: "Rendah (Low — 1.024 token)" },
  { value: "medium", label: "Sedang (Medium — 4.096 token)" },
  { value: "high", label: "Tinggi (High — 16.384 token)" },
] as const;

export const EMBEDDING_PROVIDERS = [
  { id: "gemini", label: "Google Gemini" },
  { id: "openai", label: "OpenAI" },
  { id: "openrouter", label: "OpenRouter" },
  { id: "none", label: "Nonaktif (BM25 Saja)" },
] as const;

export const EMBEDDING_KEYS = {
  providerKey: "agent.embedding.provider",
  modelKey: "agent.embedding.model",
} as const;

export function isSuperAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes(ROLE_SUPER_ADMIN));
}

export function isAgentSettingKey(key?: string): boolean {
  return Boolean(key?.startsWith("agent."));
}

export interface SettingGroup {
  name: string;
  items: SettingResponse[];
  configured: boolean;
}

export function groupSettings(items: SettingResponse[]): SettingGroup[] {
  const groups = new Map<string, SettingResponse[]>();
  for (const item of items) {
    const name = item.group || "Umum";
    const list = groups.get(name);
    if (list) {
      list.push(item);
    } else {
      groups.set(name, [item]);
    }
  }
  return Array.from(groups.entries()).map(([name, groupItems]) => ({
    name,
    items: groupItems,
    configured: groupItems.some((item) => Boolean(item.configured)),
  }));
}

export function groupChannelSettings(items: SettingResponse[]): SettingGroup[] {
  return groupSettings(items.filter((item) => !isAgentSettingKey(item.key)));
}

export function buildClearValues(
  items: SettingResponse[],
): Record<string, string> {
  const values: Record<string, string> = {};
  for (const item of items) {
    if (item.key) {
      values[item.key] = "";
    }
  }
  return values;
}

export function parseProviderModels(rawJson?: string): ProviderModelItem[] {
  if (!rawJson) return [];
  const trimmed = rawJson.trim();
  if (!trimmed) return [];
  try {
    const parsed: unknown = JSON.parse(trimmed);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(
      (item): item is ProviderModelItem =>
        Boolean(item) && typeof item === "object" && "id" in item,
    );
  } catch {
    return [];
  }
}

export function filterModelsByCapability(
  models: ProviderModelItem[],
  capability: "chat" | "embedding",
): ProviderModelItem[] {
  return models.filter((model) => {
    const caps = model.capabilities;
    if (!caps || caps.length === 0) {
      return capability === "chat";
    }
    return caps.includes(capability);
  });
}

export function isProviderConfigured(
  itemsByKey: Record<string, SettingResponse>,
  draft: Record<string, string>,
  provider: string,
): boolean {
  const apiKeyKey = `agent.provider.${provider}.api_key`;
  const persisted = itemsByKey[apiKeyKey]?.configured;
  const draftVal = draft[apiKeyKey]?.trim() ?? "";
  return Boolean(persisted) || draftVal.length > 0;
}

export interface ProviderOption {
  id: string;
  label: string;
  configured: boolean;
}

export function buildProviderOptionList(
  itemsByKey: Record<string, SettingResponse>,
  draft: Record<string, string>,
  includeNone = false,
): ProviderOption[] {
  const source = includeNone ? EMBEDDING_PROVIDERS : LLM_PROVIDERS;
  return source.map((p) => {
    const configured =
      p.id === "none" ? true : isProviderConfigured(itemsByKey, draft, p.id);
    return {
      id: p.id,
      label:
        p.id === "none"
          ? p.label
          : `${p.label} (${configured ? "Aktif" : "Belum dikonfigurasi"})`,
      configured,
    };
  });
}
