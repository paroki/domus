export type MemoryCategory = "preference" | "fact" | "context";

export const MEMORY_CATEGORIES: MemoryCategory[] = [
  "preference",
  "fact",
  "context",
];

export const MAX_MEMORY_CHARS = 2000;
export const MAX_CHAT_MESSAGE_CHARS = 16000;

export function isSuperAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes("user:super-admin"));
}

export function formatAgentLabel(agentName?: string): string {
  switch ((agentName ?? "").toLowerCase()) {
    case "mikael":
      return "Mikael";
    case "rafael":
      return "Rafael";
    default:
      return "Gabriel";
  }
}

export function safeUrlTransform(url: string): string {
  const trimmed = url.trim();
  if (!trimmed) return "";
  try {
    const parsed = new URL(trimmed);
    const proto = parsed.protocol.toLowerCase();
    if (proto === "http:" || proto === "https:" || proto === "mailto:") {
      return trimmed;
    }
    return "";
  } catch {
    return "";
  }
}

export function validateMemoryContent(raw: string): string | null {
  const trimmed = raw.trim();
  if (!trimmed) {
    return "Konten memori wajib diisi.";
  }
  if (Array.from(trimmed).length > MAX_MEMORY_CHARS) {
    return "Konten memori maksimal 2000 karakter.";
  }
  return null;
}

export function validateChatMessage(raw: string): string | null {
  const trimmed = raw.trim();
  if (!trimmed) {
    return "Pesan wajib diisi.";
  }
  if (Array.from(trimmed).length > MAX_CHAT_MESSAGE_CHARS) {
    return "Pesan maksimal 16000 karakter.";
  }
  return null;
}

export type SSEFrame =
  | {
      event: "session";
      data: { sessionId: string; title: string; isNew: boolean };
    }
  | { event: "agent_transfer"; data: { from: string; to: string } }
  | {
      event: "tool_start";
      data: {
        agent: string;
        toolCallId: string;
        toolName: string;
        args?: string;
      };
    }
  | {
      event: "tool_end";
      data: {
        agent: string;
        toolCallId: string;
        toolName: string;
        isError: boolean;
      };
    }
  | { event: "token"; data: { agent: string; delta: string } }
  | {
      event: "done";
      data: { sessionId: string; messageId: string; agent: string };
    }
  | { event: "error"; data: { code: string; message: string } };

export interface SSEParser {
  push(chunk: string): SSEFrame[];
  flush(): SSEFrame[];
}

export function createSSEParser(): SSEParser {
  let buffer = "";
  let currentEvent = "";
  let dataLines: string[] = [];

  function dispatchFrame(out: SSEFrame[]) {
    if (!currentEvent && dataLines.length === 0) return;
    const rawData = dataLines.join("\n");
    const ev = currentEvent;
    currentEvent = "";
    dataLines = [];
    if (!ev || !rawData) return;
    try {
      const parsed: unknown = JSON.parse(rawData);
      if (parsed && typeof parsed === "object") {
        out.push({ event: ev, data: parsed } as SSEFrame);
      }
    } catch {
      // ignore malformed JSON frames
    }
  }

  function processLine(line: string, out: SSEFrame[]) {
    if (line === "") {
      dispatchFrame(out);
      return;
    }
    if (line.startsWith(":")) {
      return;
    }
    if (line.startsWith("event:")) {
      currentEvent = line.slice(6).trim();
      return;
    }
    if (line.startsWith("data:")) {
      const rest = line.slice(5);
      dataLines.push(rest.startsWith(" ") ? rest.slice(1) : rest);
    }
  }

  return {
    push(chunk: string): SSEFrame[] {
      const out: SSEFrame[] = [];
      buffer += chunk;
      let idx = buffer.indexOf("\n");
      while (idx !== -1) {
        const line = buffer.slice(0, idx).replace(/\r$/, "");
        buffer = buffer.slice(idx + 1);
        processLine(line, out);
        idx = buffer.indexOf("\n");
      }
      return out;
    },
    flush(): SSEFrame[] {
      const out: SSEFrame[] = [];
      if (buffer.length > 0) {
        processLine(buffer.replace(/\r$/, ""), out);
        buffer = "";
      }
      dispatchFrame(out);
      return out;
    },
  };
}
