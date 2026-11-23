import type {
  AgentMessageResponse,
  AgentSSEEvent,
  AgentSSEEventType,
  AgentToolCallDTO,
} from "./api/types";

export const ROLE_PARISH_ADMIN = "ref:parish-admin";
export const MAX_PINNED_MEMORIES = 5;
export const MAX_MEMORY_CONTENT_LEN = 500;

export function isParishAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes(ROLE_PARISH_ADMIN));
}

export function canMutateMemory(
  memory: { scope?: string; userId?: string },
  me?: { id?: string; roles?: string[] },
): boolean {
  if (memory.scope === "parish") {
    return isParishAdmin(me?.roles);
  }
  return Boolean(me?.id && (!memory.userId || memory.userId === me.id));
}

export type AgentTone = "neutral" | "success" | "warning";

export interface AgentMeta {
  name: string;
  roleLabel: string;
  tone: AgentTone;
}

export function agentDisplayMeta(agentName?: string): AgentMeta {
  switch (agentName?.toLowerCase()) {
    case "mikael":
      return {
        name: "Mikael",
        roleLabel: "Spesialis Kegiatan",
        tone: "success",
      };
    case "rafael":
      return {
        name: "Rafael",
        roleLabel: "Spesialis Organisasi",
        tone: "warning",
      };
    default:
      return {
        name: "Gabriel",
        roleLabel: "Orkestrator",
        tone: "neutral",
      };
  }
}

export const MEMORY_SCOPES = [
  { value: "user", label: "Memori Pribadi" },
  { value: "parish", label: "FAQ & Pengetahuan Paroki" },
] as const;

export const MEMORY_CATEGORIES = [
  { value: "profile", label: "Profil" },
  { value: "preference", label: "Preferensi" },
  { value: "ministry", label: "Pelayanan" },
  { value: "parish_faq", label: "FAQ Paroki" },
] as const;

export function memoryCategoryMeta(category?: string): {
  label: string;
  tone: AgentTone;
} {
  switch (category) {
    case "profile":
      return { label: "Profil", tone: "neutral" };
    case "preference":
      return { label: "Preferensi", tone: "success" };
    case "ministry":
      return { label: "Pelayanan", tone: "warning" };
    case "parish_faq":
      return { label: "FAQ Paroki", tone: "neutral" };
    default:
      return { label: category || "Umum", tone: "neutral" };
  }
}

export function toolDisplayMeta(toolName?: string): {
  label: string;
  category: string;
} {
  const name = toolName ?? "";
  if (name === "delegate_specialist") {
    return { label: name, category: "Delegasi Spesialis" };
  }
  if (name.startsWith("org_")) {
    return { label: name, category: "Organisasi" };
  }
  if (name.startsWith("event_")) {
    return { label: name, category: "Kegiatan" };
  }
  if (name.startsWith("memory_")) {
    return { label: name, category: "Memori" };
  }
  return { label: name || "alat", category: "Alat" };
}

export function sortMessagesChronologically(
  messages: AgentMessageResponse[],
): AgentMessageResponse[] {
  return [...messages].sort((a, b) => {
    const ta = a.createdAt ? new Date(a.createdAt).getTime() : 0;
    const tb = b.createdAt ? new Date(b.createdAt).getTime() : 0;
    if (ta !== tb) return ta - tb;
    return (a.id ?? "").localeCompare(b.id ?? "");
  });
}

export function hasInterruptedTurn(messages: AgentMessageResponse[]): boolean {
  if (messages.length === 0) return false;
  const sorted = sortMessagesChronologically(messages);
  const last = sorted[sorted.length - 1];
  if (!last) return false;
  if (last.role === "user" || last.role === "tool") return true;
  return last.role === "assistant" && (last.toolCalls?.length ?? 0) > 0;
}

export interface CorrelatedToolStep {
  id: string;
  name: string;
  arguments: string;
  agentName: string;
  status: "running" | "done" | "error";
  result?: string;
}

export interface TimelineItem {
  id: string;
  role: "user" | "assistant";
  agentName: string;
  content: string;
  thinking?: string;
  toolSteps: CorrelatedToolStep[];
  createdAt?: string;
}

export function isToolErrorResult(raw?: string): boolean {
  if (!raw) return false;
  try {
    const parsed: unknown = JSON.parse(raw);
    return Boolean(parsed && typeof parsed === "object" && "error" in parsed);
  } catch {
    return false;
  }
}

export function correlateToolCalls(
  messages: AgentMessageResponse[],
): Map<string, AgentMessageResponse> {
  const results = new Map<string, AgentMessageResponse>();
  for (const m of messages) {
    if (m.role === "tool" && m.toolCallId) {
      results.set(m.toolCallId, m);
    }
  }
  return results;
}

export function buildTimelineItems(
  messages: AgentMessageResponse[],
): TimelineItem[] {
  const sorted = sortMessagesChronologically(messages);
  const toolResults = correlateToolCalls(sorted);
  const items: TimelineItem[] = [];

  for (const m of sorted) {
    if (m.role === "tool") {
      continue;
    }
    const toolSteps: CorrelatedToolStep[] = (m.toolCalls ?? []).map(
      (tc: AgentToolCallDTO) => {
        const res = tc.id ? toolResults.get(tc.id) : undefined;
        const resultContent = res?.content;
        const status: CorrelatedToolStep["status"] = res
          ? isToolErrorResult(resultContent)
            ? "error"
            : "done"
          : "done";
        return {
          id: tc.id ?? `${m.id}-${tc.name}`,
          name: tc.name ?? "",
          arguments: tc.arguments ?? "",
          agentName: tc.agentName || m.agentName || "gabriel",
          status,
          result: resultContent,
        };
      },
    );

    items.push({
      id: m.id ?? `${m.createdAt}-${items.length}`,
      role: m.role === "user" ? "user" : "assistant",
      agentName: m.agentName || "gabriel",
      content: m.content ?? "",
      thinking: m.thinking,
      toolSteps,
      createdAt: m.createdAt,
    });
  }

  return items;
}

export function parseSSEBuffer(buffer: string): {
  events: AgentSSEEvent[];
  remainder: string;
} {
  const normalized = buffer.replace(/\r\n/g, "\n");
  const blocks = normalized.split("\n\n");
  const remainder = blocks.pop() ?? "";
  const events: AgentSSEEvent[] = [];

  for (const block of blocks) {
    const trimmed = block.trim();
    if (!trimmed) continue;

    let eventType: AgentSSEEventType | "" = "";
    const dataLines: string[] = [];

    for (const rawLine of trimmed.split("\n")) {
      if (rawLine.startsWith(":")) {
        continue;
      }
      if (rawLine.startsWith("event:")) {
        eventType = rawLine.slice("event:".length).trim() as AgentSSEEventType;
      } else if (rawLine.startsWith("data:")) {
        dataLines.push(rawLine.slice("data:".length).trimStart());
      }
    }

    if (dataLines.length === 0) continue;
    const rawData = dataLines.join("\n");
    try {
      const parsed = JSON.parse(rawData) as AgentSSEEvent;
      if (!parsed.event && eventType) {
        parsed.event = eventType;
      }
      if (parsed.event) {
        events.push(parsed);
      }
    } catch {
      // Ignore malformed JSON frames safely.
    }
  }

  return { events, remainder };
}

export type ContentSegment =
  | { type: "markdown"; content: string }
  | { type: "a2ui"; messages: Record<string, unknown>[] }
  | { type: "a2ui-pending" };

function isValidA2uiEnvelopeArray(
  raw: string,
): Record<string, unknown>[] | null {
  try {
    const parsed: unknown = JSON.parse(raw.trim());
    if (!Array.isArray(parsed) || parsed.length === 0) return null;
    for (const item of parsed) {
      if (!item || typeof item !== "object") return null;
      const rec = item as Record<string, unknown>;
      if (rec.version !== "v0.9") return null;
      if (
        !("createSurface" in rec) &&
        !("updateComponents" in rec) &&
        !("updateDataModel" in rec) &&
        !("deleteSurface" in rec)
      ) {
        return null;
      }
    }
    return parsed as Record<string, unknown>[];
  } catch {
    return null;
  }
}

export function parseContentSegments(
  content: string,
  isStreaming = false,
): ContentSegment[] {
  if (!content.includes("```a2ui")) {
    return content ? [{ type: "markdown", content }] : [];
  }

  const segments: ContentSegment[] = [];
  const fenceRe = /```a2ui[ \t]*\r?\n([\s\S]*?)\r?\n```/g;
  let lastIndex = 0;

  for (const match of content.matchAll(fenceRe)) {
    const idx = match.index ?? 0;
    const prose = content.slice(lastIndex, idx).trim();
    if (prose) {
      segments.push({ type: "markdown", content: prose });
    }
    const envelopes = isValidA2uiEnvelopeArray(match[1] ?? "");
    if (envelopes) {
      segments.push({ type: "a2ui", messages: envelopes });
    }
    lastIndex = idx + match[0].length;
  }

  const tail = content.slice(lastIndex);
  if (isStreaming) {
    const openIdx = tail.indexOf("```a2ui");
    if (openIdx !== -1) {
      const prose = tail.slice(0, openIdx).trim();
      if (prose) {
        segments.push({ type: "markdown", content: prose });
      }
      segments.push({ type: "a2ui-pending" });
      return segments;
    }
  }

  const trimmedTail = tail.trim();
  if (trimmedTail) {
    segments.push({ type: "markdown", content: trimmedTail });
  }
  return segments;
}
