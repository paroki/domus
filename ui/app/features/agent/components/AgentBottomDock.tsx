import { useQueryClient } from "@tanstack/react-query";
import { ChevronDown, Maximize2, Plus, RotateCcw } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useLocation, useNavigate } from "react-router";
import { ApiError } from "~/shared/api/client";
import { useMe } from "~/shared/api/session";
import { agentKeys } from "../api/keys";
import { useAgentSession } from "../api/queries";
import { streamAgentChat } from "../api/stream";
import type { NavigationPayload, QuestionPayload } from "../api/types";
import {
  buildPageContext,
  getDockSessionKey,
  performAgentNavigate,
} from "../model";
import { ChatComposer } from "./ChatComposer";
import {
  AgentAvatar,
  ChatTimeline,
  type LiveToolCall,
  type LiveTransfer,
} from "./ChatTimeline";

function readStoredDockSession(parokiId: string): string | undefined {
  if (typeof window === "undefined" || !parokiId) return undefined;
  return window.localStorage.getItem(getDockSessionKey(parokiId)) || undefined;
}

export function AgentBottomDock() {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const { data: me } = useMe();
  const parokiId = me?.activeParokiId ?? "";

  const [expanded, setExpanded] = useState(false);
  const [sessionId, setSessionId] = useState<string | undefined>(() =>
    readStoredDockSession(parokiId),
  );
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [lastFailedMsg, setLastFailedMsg] = useState<string | null>(null);

  const [streaming, setStreaming] = useState(false);
  const [pendingUserMsg, setPendingUserMsg] = useState<string | null>(null);
  const [streamingAgent, setStreamingAgent] = useState("gabriel");
  const [streamingText, setStreamingText] = useState("");
  const [liveTransfers, setLiveTransfers] = useState<LiveTransfer[]>([]);
  const [liveTools, setLiveTools] = useState<LiveToolCall[]>([]);
  const [liveQuestion, setLiveQuestion] = useState<QuestionPayload | null>(
    null,
  );
  const [liveNavigation, setLiveNavigation] =
    useState<NavigationPayload | null>(null);

  const abortRef = useRef<AbortController | null>(null);
  const activeStreamSessionRef = useRef<string | undefined>(sessionId);

  useEffect(() => {
    const stored = readStoredDockSession(parokiId);
    setSessionId(stored);
    activeStreamSessionRef.current = stored;
    setErrorMsg(null);
    setLastFailedMsg(null);
  }, [parokiId]);

  const hasParish = Boolean(parokiId);
  const { data: sessionDetail, isLoading: isLoadingSession } = useAgentSession(
    sessionId,
    hasParish && expanded && Boolean(sessionId),
  );

  if (!hasParish) {
    return null;
  }

  function resetStreamUI() {
    setStreaming(false);
    setPendingUserMsg(null);
    setStreamingAgent("gabriel");
    setStreamingText("");
    setLiveTransfers([]);
    setLiveTools([]);
    setLiveNavigation(null);
    abortRef.current = null;
  }

  function handleNewSession() {
    abortRef.current?.abort();
    if (typeof window !== "undefined" && parokiId) {
      window.localStorage.removeItem(getDockSessionKey(parokiId));
    }
    setSessionId(undefined);
    activeStreamSessionRef.current = undefined;
    setErrorMsg(null);
    setLastFailedMsg(null);
    setLiveQuestion(null);
    resetStreamUI();
  }

  async function handleSend(message: string, isRetry = false) {
    setErrorMsg(null);
    setLastFailedMsg(null);
    setPendingUserMsg(message);
    setStreaming(true);
    setStreamingAgent("gabriel");
    setStreamingText("");
    setLiveTransfers([]);
    setLiveTools([]);
    setLiveQuestion(null);
    setLiveNavigation(null);

    const controller = new AbortController();
    abortRef.current = controller;

    try {
      await streamAgentChat({
        body: {
          sessionId,
          message,
          ...(isRetry && sessionId ? { retry: true } : {}),
          pageContext: buildPageContext(location.pathname, location.hash),
        },
        signal: controller.signal,
        onFrame: (frame) => {
          switch (frame.event) {
            case "session": {
              const sid = frame.data.sessionId;
              activeStreamSessionRef.current = sid;
              setSessionId(sid);
              if (typeof window !== "undefined" && parokiId) {
                window.localStorage.setItem(getDockSessionKey(parokiId), sid);
              }
              void qc.invalidateQueries({ queryKey: ["agent", "sessions"] });
              break;
            }
            case "agent_transfer": {
              setLiveTransfers((prev) => [...prev, frame.data]);
              setStreamingAgent(frame.data.to);
              break;
            }
            case "tool_start": {
              setLiveTools((prev) => [
                ...prev,
                {
                  toolCallId: frame.data.toolCallId,
                  agent: frame.data.agent,
                  toolName: frame.data.toolName,
                  status: "running",
                  args: frame.data.args,
                },
              ]);
              break;
            }
            case "tool_end": {
              setLiveTools((prev) =>
                prev.map((tl) =>
                  tl.toolCallId === frame.data.toolCallId
                    ? {
                        ...tl,
                        status: frame.data.isError ? "error" : "done",
                        output: frame.data.output,
                      }
                    : tl,
                ),
              );
              break;
            }
            case "token": {
              setStreamingAgent(frame.data.agent || "gabriel");
              setStreamingText((prev) => prev + frame.data.delta);
              break;
            }
            case "question": {
              setLiveQuestion(frame.data);
              break;
            }
            case "navigate": {
              setLiveNavigation(frame.data);
              if (frame.data.autoRedirect === true) {
                void qc.invalidateQueries();
                performAgentNavigate(
                  frame.data.path,
                  location.pathname,
                  navigate,
                );
              }
              break;
            }
            case "error": {
              setErrorMsg(frame.data.message);
              setLastFailedMsg(message);
              break;
            }
            case "done": {
              break;
            }
          }
        },
      });
    } catch (err) {
      if (err instanceof DOMException && err.name === "AbortError") {
        // User stopped stream
      } else if (err instanceof ApiError || err instanceof Error) {
        setErrorMsg(err.message);
        setLastFailedMsg(message);
      } else {
        setErrorMsg("Terjadi kesalahan saat menghubungi asisten AI.");
        setLastFailedMsg(message);
      }
    } finally {
      const targetSid = activeStreamSessionRef.current ?? sessionId;
      resetStreamUI();
      if (targetSid) {
        void qc.invalidateQueries({ queryKey: agentKeys.session(targetSid) });
      }
      void qc.invalidateQueries({ queryKey: ["agent", "sessions"] });
    }
  }

  function handleStop() {
    abortRef.current?.abort();
  }

  if (!expanded) {
    return (
      <div className="pointer-events-none fixed inset-x-0 bottom-0 z-40 flex justify-center px-3 pt-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] sm:px-4">
        <div className="pointer-events-auto flex w-full items-center gap-2 rounded-2xl border border-line bg-surface/95 px-2.5 py-1.5 shadow-lg backdrop-blur-md sm:max-w-2xl">
          <AgentAvatar agent="gabriel" />
          <span className="hidden max-w-36 truncate rounded-full border border-line bg-canvas px-2.5 py-0.5 font-mono text-[11px] text-muted sm:inline-flex">
            {location.pathname}
          </span>
          <button
            type="button"
            onClick={() => setExpanded(true)}
            className="flex min-h-11 min-w-0 flex-1 items-center rounded-xl border border-line bg-bg px-3 text-left text-base text-muted/80 transition-colors hover:border-primary/40 hover:text-fg sm:text-sm"
          >
            <span className="truncate">{t("agent.dockQuickPlaceholder")}</span>
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="pointer-events-none fixed inset-x-0 bottom-0 z-40 flex justify-center px-2 sm:px-4">
      <section
        aria-label={t("agent.title")}
        className="pointer-events-auto flex max-h-[75dvh] w-full flex-col overflow-hidden rounded-t-2xl border border-b-0 border-line bg-canvas shadow-2xl pb-[env(safe-area-inset-bottom)] sm:max-w-2xl"
      >
        <header className="flex items-center justify-between gap-2 border-b border-line bg-surface px-3 py-2 sm:px-4">
          <div className="flex min-w-0 items-center gap-2">
            <AgentAvatar agent={streaming ? streamingAgent : "gabriel"} />
            <h2 className="truncate text-sm font-semibold text-fg">
              {sessionDetail?.title || t("agent.title")}
            </h2>
            <span className="hidden max-w-36 truncate rounded-full border border-line bg-canvas px-2 py-0.5 font-mono text-[11px] text-muted sm:inline-flex">
              {location.pathname}
            </span>
          </div>

          <div className="flex shrink-0 items-center gap-1">
            <button
              type="button"
              onClick={handleNewSession}
              aria-label={t("agent.newSession")}
              title={t("agent.newSession")}
              className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-xl border border-line bg-bg px-2.5 text-xs font-medium text-fg transition-colors hover:bg-fg/6"
            >
              <Plus aria-hidden className="size-4" />
            </button>
            <Link
              to={sessionId ? `/agent/sessions/${sessionId}` : "/agent"}
              onClick={() => setExpanded(false)}
              className="inline-flex min-h-11 items-center gap-1.5 rounded-xl border border-line bg-bg px-3 text-xs font-medium text-fg transition-colors hover:bg-fg/6"
            >
              <Maximize2 aria-hidden className="size-3.5 text-primary" />
              <span>{t("agent.dockOpenFull")}</span>
            </Link>
            <button
              type="button"
              onClick={() => setExpanded(false)}
              aria-label={t("agent.dockMinimize")}
              title={t("agent.dockMinimize")}
              className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-xl border border-line bg-bg text-muted transition-colors hover:bg-fg/6 hover:text-fg"
            >
              <ChevronDown aria-hidden className="size-4" />
            </button>
          </div>
        </header>

        {errorMsg ? (
          <div
            role="alert"
            className="mx-3 mt-2 flex items-center justify-between gap-2 rounded-xl border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs text-red-700 dark:text-red-300"
          >
            <span className="min-w-0 flex-1">{errorMsg}</span>
            {lastFailedMsg && !streaming ? (
              <button
                type="button"
                onClick={() => void handleSend(lastFailedMsg, true)}
                className="inline-flex min-h-9 shrink-0 items-center gap-1 rounded-lg border border-red-500/30 bg-red-500/15 px-2.5 py-1 font-medium text-red-700 hover:bg-red-500/25 dark:text-red-200"
              >
                <RotateCcw aria-hidden className="size-3" />
                {t("agent.retry")}
              </button>
            ) : null}
          </div>
        ) : null}

        <ChatTimeline
          messages={sessionDetail?.messages ?? []}
          loading={Boolean(sessionId) && isLoadingSession}
          pendingUserMessage={pendingUserMsg}
          streaming={streaming}
          streamingAgent={streamingAgent}
          streamingText={streamingText}
          liveTransfers={liveTransfers}
          liveTools={liveTools}
          liveQuestion={liveQuestion}
          liveNavigation={liveNavigation}
          onRetry={
            !streaming && !errorMsg
              ? (msg) => void handleSend(msg, true)
              : undefined
          }
          onSelectPrompt={
            !streaming ? (prompt) => void handleSend(prompt) : undefined
          }
          onSendMessage={!streaming ? (msg) => void handleSend(msg) : undefined}
        />

        <ChatComposer
          streaming={streaming}
          disabled={!hasParish}
          focusKey={expanded ? (sessionId ?? "dock-new") : "dock-closed"}
          onSend={(msg) => void handleSend(msg)}
          onStop={handleStop}
        />
      </section>
    </div>
  );
}
