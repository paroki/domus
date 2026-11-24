import { useQueryClient } from "@tanstack/react-query";
import { Brain, Menu } from "lucide-react";
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
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

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

  const hasParish = Boolean(me?.activeParishId);
  const canManageProviders = isSuperAdmin(me?.roles);
  const { data: sessionDetail } = useAgentSession(
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

  async function handleSend(message: string) {
    setErrorMsg(null);
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
      } else {
        setErrorMsg("Terjadi kesalahan saat menghubungi asisten AI.");
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

          <button
            type="button"
            onClick={() => setMemoryOpen(true)}
            className="inline-flex items-center gap-1.5 rounded-lg border border-line bg-bg px-3 py-1.5 text-xs font-medium text-fg hover:bg-fg/6"
          >
            <Brain aria-hidden className="size-3.5 text-primary" />
            {t("agent.memoryButton")}
          </button>
        </div>

        {errorMsg ? (
          <div
            role="alert"
            className="mx-4 mt-3 rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-2.5 text-xs text-red-700 dark:text-red-300 sm:mx-6"
          >
            {errorMsg}
          </div>
        ) : null}

        <ChatTimeline
          messages={sessionDetail?.messages ?? []}
          pendingUserMessage={pendingUserMsg}
          streaming={streaming}
          streamingAgent={streamingAgent}
          streamingText={streamingText}
          liveTransfers={liveTransfers}
          liveTools={liveTools}
        />

        <ChatComposer
          streaming={streaming}
          disabled={!hasParish}
          onSend={(msg) => void handleSend(msg)}
          onStop={handleStop}
        />
      </div>

      <MemoryDrawer open={memoryOpen} onClose={() => setMemoryOpen(false)} />
    </div>
  );
}
