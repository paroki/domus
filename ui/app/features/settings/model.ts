import {
  Bot,
  Calendar,
  Cpu,
  type LucideIcon,
  MessageCircle,
  Send,
  Sparkles,
  Users,
} from "lucide-react";
import type {
  AgentProviderResponse,
  ProviderModelItem,
  SettingResponse,
} from "./api/types";

export const ROLE_SUPER_ADMIN = "user:super-admin";

export interface SettingsLauncherItem {
  id: "telegram" | "whatsapp" | "aiProvider" | "aiAgent";
  path: string;
  plate: 1 | 2 | 3 | 4 | 5 | 6;
  icon: LucideIcon;
}

export const SETTINGS_LAUNCHERS: SettingsLauncherItem[] = [
  { id: "telegram", path: "/settings/telegram", plate: 1, icon: Send },
  {
    id: "whatsapp",
    path: "/settings/whatsapp",
    plate: 2,
    icon: MessageCircle,
  },
  { id: "aiProvider", path: "/settings/ai-provider", plate: 3, icon: Cpu },
  { id: "aiAgent", path: "/settings/ai-agent", plate: 4, icon: Bot },
];

export const CHANNEL_GROUP_NAMES = {
  telegram: "Telegram Bot",
  whatsapp: "GOWA (WhatsApp)",
} as const;

export type LLMProviderType = "gemini" | "anthropic" | "openai" | "openrouter";

export interface LLMProviderDef {
  id: LLMProviderType;
  label: string;
  defaultBaseURL: string;
  supportsEmbedding: boolean;
}

export const LLM_PROVIDERS: LLMProviderDef[] = [
  {
    id: "gemini",
    label: "Google AI Studio (Gemini)",
    defaultBaseURL: "https://generativelanguage.googleapis.com",
    supportsEmbedding: true,
  },
  {
    id: "anthropic",
    label: "Anthropic Claude",
    defaultBaseURL: "https://api.anthropic.com",
    supportsEmbedding: false,
  },
  {
    id: "openai",
    label: "OpenAI",
    defaultBaseURL: "https://api.openai.com/v1",
    supportsEmbedding: true,
  },
  {
    id: "openrouter",
    label: "OpenRouter",
    defaultBaseURL: "https://openrouter.ai/api/v1",
    supportsEmbedding: true,
  },
];

export function getProviderLabel(provider?: string): string {
  return LLM_PROVIDERS.find((p) => p.id === provider)?.label ?? provider ?? "";
}

export interface AgentPersonaDef {
  id: "gabriel" | "mikael" | "rafael";
  name: string;
  role: string;
  plate: 1 | 2 | 3 | 4 | 5 | 6;
  icon: LucideIcon;
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
    plate: 1,
    icon: Sparkles,
    tools: ["delegate_specialist", "memory_*"],
    providerKey: "agent.gabriel.provider_id",
    modelKey: "agent.gabriel.model",
    effortKey: "agent.gabriel.effort",
    instructionKey: "agent.gabriel.instruction",
  },
  {
    id: "mikael",
    name: "Mikael",
    role: "Spesialis Kegiatan & Jadwal",
    plate: 2,
    icon: Calendar,
    tools: ["event_*"],
    providerKey: "agent.mikael.provider_id",
    modelKey: "agent.mikael.model",
    effortKey: "agent.mikael.effort",
    instructionKey: "agent.mikael.instruction",
  },
  {
    id: "rafael",
    name: "Rafael",
    role: "Spesialis Organisasi & Komunitas",
    plate: 3,
    icon: Users,
    tools: ["org_*"],
    providerKey: "agent.rafael.provider_id",
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

export const EMBEDDING_KEYS = {
  providerKey: "agent.embedding.provider_id",
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

export interface ProviderOption {
  id: string;
  label: string;
  provider?: string;
}

export function findAgentPersona(id?: string): AgentPersonaDef | undefined {
  if (!id) return undefined;
  return AGENT_PERSONAS.find((persona) => persona.id === id);
}

export function findChannelGroup(
  items: SettingResponse[],
  groupName: string,
): SettingGroup | undefined {
  return groupChannelSettings(items).find((group) => group.name === groupName);
}

export function indexSettingsByKey(
  items?: SettingResponse[],
): Record<string, SettingResponse> {
  const map: Record<string, SettingResponse> = {};
  for (const item of items ?? []) {
    if (item.key) {
      map[item.key] = item;
    }
  }
  return map;
}

export function buildSettingsDraft(
  items?: SettingResponse[],
): Record<string, string> {
  const next: Record<string, string> = {};
  for (const item of items ?? []) {
    if (item.key) {
      next[item.key] = item.value ?? "";
    }
  }
  return next;
}

export function getProviderModelsByCapability(
  providers: AgentProviderResponse[],
  providerId: string,
  capability: "chat" | "embedding",
): ProviderModelItem[] {
  if (!providerId || providerId === "none") return [];
  const found = providers.find((p) => p.id === providerId);
  if (!found?.models) return [];
  return filterModelsByCapability(found.models, capability);
}

export function buildProviderOptionList(
  providers: AgentProviderResponse[],
  includeNone = false,
): ProviderOption[] {
  const filtered = includeNone
    ? providers.filter((p) => {
        const def = LLM_PROVIDERS.find((d) => d.id === p.provider);
        return def?.supportsEmbedding ?? false;
      })
    : providers;

  const opts: ProviderOption[] = filtered.map((p) => ({
    id: p.id ?? "",
    label: `${p.name ?? p.id} (${getProviderLabel(p.provider)})`,
    provider: p.provider,
  }));

  if (includeNone) {
    opts.push({ id: "none", label: "Nonaktif (BM25 Saja)" });
  }
  return opts;
}
