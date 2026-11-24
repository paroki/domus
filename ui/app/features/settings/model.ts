import { Bot, Cpu, type LucideIcon, MessageCircle, Send } from "lucide-react";
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

export const AGENT_MODEL_KEYS = [
  "agent.model.gabriel",
  "agent.model.mikael",
  "agent.model.rafael",
  "agent.model.memory_extractor",
  "agent.model.embedding",
] as const;

export const AGENT_PROMPT_KEYS = [
  "agent.prompt.gabriel",
  "agent.prompt.mikael",
  "agent.prompt.rafael",
  "agent.prompt.memory_extractor",
] as const;

export const CHANNEL_GROUP_NAMES = {
  telegram: "Telegram Bot",
  whatsapp: "GOWA (WhatsApp)",
} as const;

export type LLMProviderType = "gemini" | "anthropic" | "openai" | "openrouter";

export interface LLMProviderDef {
  id: LLMProviderType;
  label: string;
  defaultBaseURL: string;
}

export const LLM_PROVIDERS: LLMProviderDef[] = [
  {
    id: "gemini",
    label: "Google AI Studio (Gemini)",
    defaultBaseURL: "https://generativelanguage.googleapis.com",
  },
  {
    id: "anthropic",
    label: "Anthropic Claude",
    defaultBaseURL: "https://api.anthropic.com",
  },
  {
    id: "openai",
    label: "OpenAI",
    defaultBaseURL: "https://api.openai.com/v1",
  },
  {
    id: "openrouter",
    label: "OpenRouter",
    defaultBaseURL: "https://openrouter.ai/api/v1",
  },
];

export function getProviderLabel(provider?: string): string {
  return LLM_PROVIDERS.find((p) => p.id === provider)?.label ?? provider ?? "";
}

export function isSuperAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes(ROLE_SUPER_ADMIN));
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

export function findChannelGroup(
  items: SettingResponse[],
  groupName: string,
): SettingGroup | undefined {
  return groupSettings(items).find((group) => group.name === groupName);
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

export interface ProviderModelOption {
  value: string;
  label: string;
  providerId: string;
  modelId: string;
}

export function buildProviderModelOptions(
  providers: AgentProviderResponse[],
  capability: "chat" | "embedding",
): ProviderModelOption[] {
  const options: ProviderModelOption[] = [];
  for (const provider of providers) {
    if (!provider.id) continue;
    const models = filterModelsByCapability(
      (provider.models ?? []) as ProviderModelItem[],
      capability,
    );
    for (const model of models) {
      if (!model.id) continue;
      options.push({
        value: `${provider.id}:${model.id}`,
        label: `${provider.name ?? provider.provider ?? provider.id} — ${model.name || model.id}`,
        providerId: provider.id,
        modelId: model.id,
      });
    }
  }
  return options;
}

export function parseModelSelection(value?: string): {
  providerId: string;
  modelId: string;
} {
  if (!value) return { providerId: "", modelId: "" };
  const idx = value.indexOf(":");
  if (idx === -1) return { providerId: "", modelId: value };
  return {
    providerId: value.slice(0, idx),
    modelId: value.slice(idx + 1),
  };
}

export type { AgentProviderResponse };
