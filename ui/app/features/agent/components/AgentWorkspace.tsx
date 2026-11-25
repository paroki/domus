import { useQueryClient } from "@tanstack/react-query";
import { Brain, Menu, RotateCcw, Terminal } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { ApiError } from "~/shared/api/client";
import { useMe } from "~/shared/api/session";
import { agentKeys } from "../api/keys";
import { useAgentSession } from "../api/queries";
import { streamAgentChat } from "../api/stream";
import { isSuperAdmin } from "../model";
import { ChatComposer } from "./ChatComposer";
import {
  ChatTimeline,
  type LiveToolCall,
  type LiveTransfer,
} from "./ChatTimeline";
import { MemoryDrawer } from "./MemoryDrawer";
import { SessionSidebar } from "./SessionSidebar";

interface AgentWorkspaceProps {
  initialSessionId?: string;
}

export function AgentWorkspace({ initialSessionId }: AgentWorkspaceProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const { data: me } = useMe();

  const [sessionId, setSessionId] = useState<string | undefined>(
    initialSessionId,
  );
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false);
  const [memoryOpen, setMemoryOpen] = useState(false);
  const [verbose, setVerbose] = useState(() => {
    if (typeof window === "undefined") return false;
    return window.localStorage.getItem("domus:agent:verbose") === "true";
  });
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [lastFailedMsg, setLastFailedMsg] = useState<string | null>(null);

  const [streaming, setStreaming] = useState(false);
  const [pendingUserMsg, setPendingUserMsg] = useState<string | null>(null);
  const [streamingAgent, setStreamingAgent] = useState("gabriel");
  const [streamingText, setStreamingText] = useState("");
  const [liveTransfers, setLiveTransfers] = useState<LiveTransfer[]>([]);
  const [liveTools, setLiveTools] = useState<LiveToolCall[]>([]);

  const abortRef = useRef<AbortController | null>(null);
  const activeStreamSessionRef = useRef<string | undefined>(initialSessionId);

  useEffect(() => {
    setSessionId(initialSessionId);
    activeStreamSessionRef.current = initialSessionId;
  }, [initialSessionId]);

  const hasParish = Boolean(me?.activeParokiId);
  const canManageProviders = isSuperAdmin(me?.roles);
  const { data: sessionDetail, isLoading: isLoadingSession } = useAgentSession(
    sessionId,
    hasParish && Boolean(sessionId),
  );

  function resetStreamUI() {
    setStreaming(false);
    setPendingUserMsg(null);
    setStreamingAgent("gabriel");
    setStreamingText("");
    setLiveTransfers([]);
    setLiveTools([]);
    abortRef.current = null;
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

    const controller = new AbortController();
    abortRef.current = controller;

    try {
      await streamAgentChat({
        body: {
          sessionId,
          message,
          ...(isRetry && sessionId ? { retry: true } : {}),
        },
        signal: controller.signal,
        onFrame: (frame) => {
          switch (frame.event) {
            case "session": {
              const sid = frame.data.sessionId;
              activeStreamSessionRef.current = sid;
              setSessionId(sid);
              if (
                typeof window !== "undefined" &&
                window.location.pathname !== `/agent/sessions/${sid}`
              ) {
                window.history.replaceState(null, "", `/agent/sessions/${sid}`);
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
        // User clicked Hentikan; exit streaming cleanly without error banner
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

  if (me && !hasParish) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-2 px-6 text-center">
        <h1 className="text-xl font-bold text-fg">
          {t("agent.noParishTitle")}
        </h1>
        <p className="max-w-md text-sm text-muted">
          {t("agent.noParishDescription")}
        </p>
      </div>
    );
  }

  return (
    <div className="flex min-h-0 flex-1 overflow-hidden">
      <SessionSidebar
        activeSession={sessionDetail}
        activeSessionId={sessionId}
        canManageProviders={canManageProviders}
        mobileOpen={mobileSidebarOpen}
        onCloseMobile={() => setMobileSidebarOpen(false)}
        onSelectNew={() => {
          setSessionId(undefined);
          activeStreamSessionRef.current = undefined;
          setErrorMsg(null);
          setLastFailedMsg(null);
          navigate("/agent");
        }}
      />

      <div className="flex min-w-0 flex-1 flex-col bg-canvas">
        <div className="flex items-center justify-between gap-2 border-b border-line bg-surface px-4 py-2.5 sm:px-6">
          <div className="flex min-w-0 items-center gap-2">
            <button
              type="button"
              aria-label={t("agent.sessionsTitle")}
              onClick={() => setMobileSidebarOpen(true)}
              className="rounded-lg p-1.5 text-muted hover:bg-fg/8 hover:text-fg lg:hidden"
            >
              <Menu aria-hidden className="size-4" />
            </button>
            <h1 className="truncate text-sm font-semibold text-fg">
              {sessionDetail?.title || t("agent.title")}
            </h1>
          </div>

          <div className="flex items-center gap-2">
            {sessionDetail?.modelId || sessionDetail?.effectiveModelId ? (
              <span className="hidden max-w-44 truncate rounded-full border border-line bg-canvas px-2.5 py-1 text-[11px] font-medium text-muted md:inline-flex">
                {sessionDetail.modelId || sessionDetail.effectiveModelId}
              </span>
            ) : null}
            <button
              type="button"
              aria-pressed={verbose}
              onClick={() => {
                setVerbose((prev) => {
                  const next = !prev;
                  if (typeof window !== "undefined") {
                    window.localStorage.setItem(
                      "domus:agent:verbose",
                      String(next),
                    );
                  }
                  return next;
                });
              }}
              className={`inline-flex items-center gap-1.5 rounded-lg border px-2.5 py-1.5 text-xs font-medium transition-colors sm:px-3 ${
                verbose
                  ? "border-primary/40 bg-primary/10 text-primary"
                  : "border-line bg-bg text-muted hover:bg-fg/6 hover:text-fg"
              }`}
            >
              <Terminal aria-hidden className="size-3.5" />
              <span className="hidden sm:inline">{t("agent.verboseMode")}</span>
            </button>

            <button
              type="button"
              onClick={() => setMemoryOpen(true)}
              className="inline-flex items-center gap-1.5 rounded-lg border border-line bg-bg px-2.5 py-1.5 text-xs font-medium text-fg hover:bg-fg/6 sm:px-3"
            >
              <Brain aria-hidden className="size-3.5 text-primary" />
              {t("agent.memoryButton")}
            </button>
          </div>
        </div>

        {errorMsg ? (
          <div
            role="alert"
            className="mx-4 mt-3 flex items-center justify-between gap-3 rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-2.5 text-xs text-red-700 dark:text-red-300 sm:mx-6"
          >
            <span className="min-w-0 flex-1">{errorMsg}</span>
            {lastFailedMsg && !streaming ? (
              <button
                type="button"
                onClick={() => void handleSend(lastFailedMsg, true)}
                className="inline-flex shrink-0 items-center gap-1 rounded-lg border border-red-500/30 bg-red-500/15 px-2.5 py-1 font-medium text-red-700 hover:bg-red-500/25 dark:text-red-200"
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
          verbose={verbose}
          onRetry={
            !streaming && !errorMsg
              ? (msg) => void handleSend(msg, true)
              : undefined
          }
          onSelectPrompt={
            !streaming ? (prompt) => void handleSend(prompt) : undefined
          }
        />

        <ChatComposer
          streaming={streaming}
          disabled={!hasParish}
          focusKey={sessionId ?? "new"}
          onSend={(msg) => void handleSend(msg)}
          onStop={handleStop}
        />
      </div>

      <MemoryDrawer open={memoryOpen} onClose={() => setMemoryOpen(false)} />
    </div>
  );
}
