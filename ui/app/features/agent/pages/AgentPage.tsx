import { useQueryClient } from "@tanstack/react-query";
import { Bot, Brain, PanelLeft, Send, Square } from "lucide-react";
import { type FormEvent, type KeyboardEvent, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { streamAgentChat } from "../api/actions";
import { resolveAgentError } from "../api/errors";
import { agentKeys } from "../api/keys";
import { useAgentMessages, useAgentSessions } from "../api/queries";
import {
  ChatTimeline,
  type LiveStreamingState,
} from "../components/ChatTimeline";
import { MemoryDrawer } from "../components/MemoryDrawer";
import { SessionSidebar } from "../components/SessionSidebar";
import { buildTimelineItems, isToolErrorResult } from "../model";

export function meta() {
  return [{ title: `${i18n.t("agent.title")} · Domus` }];
}

const EMPTY_STREAM: LiveStreamingState = {
  active: false,
  agentName: "gabriel",
  thinking: "",
  delta: "",
  toolSteps: [],
};

export default function AgentPage() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const { data: me } = useMe();
  const hasParish = Boolean(me?.activeParishId);

  const [searchParams, setSearchParams] = useSearchParams();
  const activeSessionId = searchParams.get("session") ?? "";

  const [sessionSearch, setSessionSearch] = useState("");
  const [messageSearch, setMessageSearch] = useState("");
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [input, setInput] = useState("");
  const [chatError, setChatError] = useState<string | null>(null);
  const [streaming, setStreaming] = useState<LiveStreamingState>(EMPTY_STREAM);

  const abortRef = useRef<AbortController | null>(null);

  const sessionsQuery = useAgentSessions({}, hasParish);
  const allSessions =
    sessionsQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const filteredSessions = sessionSearch.trim()
    ? allSessions.filter((s) =>
        (s.title ?? "")
          .toLowerCase()
          .includes(sessionSearch.trim().toLowerCase()),
      )
    : allSessions;

  const messagesQuery = useAgentMessages(
    activeSessionId,
    { q: messageSearch.trim() || undefined },
    hasParish && Boolean(activeSessionId),
  );
  const rawMessages =
    messagesQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const timelineItems = buildTimelineItems(rawMessages);

  function setSessionParam(id: string) {
    if (id) {
      setSearchParams({ session: id }, { replace: true });
    } else {
      setSearchParams({}, { replace: true });
    }
  }

  async function sendMessage(text: string) {
    const trimmed = text.trim();
    if (!trimmed || streaming.active) return;

    setChatError(null);
    setInput("");

    const controller = new AbortController();
    abortRef.current = controller;

    let resolvedSessionId = activeSessionId;
    let touchedMemory = false;

    setStreaming({
      active: true,
      userMessage: trimmed,
      agentName: "gabriel",
      thinking: "",
      delta: "",
      toolSteps: [],
    });

    try {
      await streamAgentChat({
        body: {
          sessionId: activeSessionId || undefined,
          message: trimmed,
        },
        signal: controller.signal,
        onEvent: (ev) => {
          switch (ev.event) {
            case "session":
              if (ev.sessionId) {
                resolvedSessionId = ev.sessionId;
                setSessionParam(ev.sessionId);
              }
              break;
            case "thinking":
              setStreaming((prev) => ({
                ...prev,
                agentName: ev.agentName || prev.agentName,
                thinking: prev.thinking + (ev.content ?? ""),
              }));
              break;
            case "tool_call":
              if (ev.toolName?.startsWith("memory_")) {
                touchedMemory = true;
              }
              setStreaming((prev) => ({
                ...prev,
                agentName: ev.agentName || prev.agentName,
                toolSteps: [
                  ...prev.toolSteps,
                  {
                    id: ev.toolCallId || `tc-${prev.toolSteps.length}`,
                    name: ev.toolName || "",
                    arguments: ev.arguments || "",
                    agentName: ev.agentName || prev.agentName,
                    status: "running",
                  },
                ],
              }));
              break;
            case "tool_result":
              setStreaming((prev) => ({
                ...prev,
                toolSteps: prev.toolSteps.map((step) =>
                  step.id === ev.toolCallId
                    ? {
                        ...step,
                        status: isToolErrorResult(ev.content)
                          ? "error"
                          : "done",
                        result: ev.content,
                      }
                    : step,
                ),
              }));
              break;
            case "delta":
              setStreaming((prev) => ({
                ...prev,
                agentName: ev.agentName || prev.agentName,
                delta: prev.delta + (ev.content ?? ""),
              }));
              break;
            case "done":
              break;
          }
        },
      });
    } catch (err) {
      if (!controller.signal.aborted) {
        setChatError(resolveAgentError(err));
      }
    } finally {
      abortRef.current = null;
      setStreaming(EMPTY_STREAM);
      void qc.invalidateQueries({ queryKey: agentKeys.sessions() });
      if (resolvedSessionId) {
        void qc.invalidateQueries({
          queryKey: agentKeys.messages(resolvedSessionId),
        });
      }
      if (touchedMemory) {
        void qc.invalidateQueries({ queryKey: agentKeys.memories() });
      }
    }
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    void sendMessage(input);
  }

  function handleKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      void sendMessage(input);
    }
  }

  if (!hasParish) {
    return (
      <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-4 px-6 py-10">
        <Card className="p-6 text-center">
          <div className="mx-auto mb-3 flex size-11 items-center justify-center rounded-xl border border-line bg-canvas">
            <Bot className="size-5 text-fg" />
          </div>
          <h1 className="text-lg font-semibold text-fg">
            {t("agent.noParishTitle")}
          </h1>
          <p className="mt-1 text-sm text-muted">
            {t("agent.noParishDescription")}
          </p>
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto flex h-[calc(100vh-4.5rem)] w-full max-w-7xl flex-col gap-3 px-4 py-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <Button
            type="button"
            variant="ghost"
            onClick={() => setSidebarOpen((v) => !v)}
            className="h-8 px-2.5 text-xs lg:hidden"
          >
            <PanelLeft className="size-4" />
          </Button>
          <div>
            <h1 className="text-lg font-bold text-fg">{t("agent.title")}</h1>
            <p className="text-xs text-muted">{t("agent.subtitle")}</p>
          </div>
        </div>

        <Button
          type="button"
          variant={drawerOpen ? "primary" : "ghost"}
          onClick={() => setDrawerOpen((v) => !v)}
          className="h-8 px-3 text-xs"
        >
          <Brain className="size-4" />
          {t("agent.memoryButton")}
        </Button>
      </div>

      <div className="flex min-h-0 flex-1 gap-3">
        <div className={`${sidebarOpen ? "flex" : "hidden"} lg:flex h-full`}>
          <SessionSidebar
            sessions={filteredSessions}
            activeSessionId={activeSessionId}
            searchQuery={sessionSearch}
            onSearchChange={setSessionSearch}
            onSelectSession={(id) => {
              setSessionParam(id);
              setSidebarOpen(false);
            }}
            onNewSession={() => {
              setSessionParam("");
              setChatError(null);
              setSidebarOpen(false);
            }}
            hasNextPage={Boolean(sessionsQuery.hasNextPage)}
            isFetchingNextPage={sessionsQuery.isFetchingNextPage}
            onLoadMore={() => void sessionsQuery.fetchNextPage()}
            onSessionDeleted={(id) => {
              if (id === activeSessionId) {
                setSessionParam("");
              }
            }}
          />
        </div>

        <main className="flex min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-line bg-surface">
          {chatError ? (
            <div
              role="alert"
              className="mx-4 mt-3 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs text-red-700 dark:text-red-300"
            >
              {chatError}
            </div>
          ) : null}

          <ChatTimeline
            items={timelineItems}
            streaming={streaming}
            messageSearch={messageSearch}
            onMessageSearchChange={setMessageSearch}
            hasOlder={Boolean(messagesQuery.hasNextPage)}
            isLoadingOlder={messagesQuery.isFetchingNextPage}
            onLoadOlder={() => void messagesQuery.fetchNextPage()}
            onStarterPrompt={(prompt) => void sendMessage(prompt)}
          />

          <form
            onSubmit={handleSubmit}
            className="flex items-end gap-2 border-t border-line bg-canvas/50 p-3"
          >
            <textarea
              rows={2}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={t("agent.composerPlaceholder")}
              className="max-h-36 min-h-10 flex-1 resize-none rounded-lg border border-line bg-canvas px-3 py-2 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
            />
            {streaming.active ? (
              <Button
                type="button"
                variant="ghost"
                onClick={() => abortRef.current?.abort()}
              >
                <Square className="size-4" />
                {t("agent.stop")}
              </Button>
            ) : (
              <Button type="submit" disabled={!input.trim()}>
                <Send className="size-4" />
                {t("agent.send")}
              </Button>
            )}
          </form>
        </main>

        <MemoryDrawer open={drawerOpen} onClose={() => setDrawerOpen(false)} />
      </div>
    </div>
  );
}
