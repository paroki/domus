export const agentKeys = {
  all: ["agent"] as const,
  sessions: (q?: string) => ["agent", "sessions", q ?? ""] as const,
  session: (sessionId?: string) =>
    ["agent", "session", sessionId ?? ""] as const,
  memories: (q?: string, category?: string) =>
    ["agent", "memories", q ?? "", category ?? ""] as const,
  providers: () => ["agent", "providers"] as const,
};
