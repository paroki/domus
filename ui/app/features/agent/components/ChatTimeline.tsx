import {
  ArrowRight,
  Bot,
  CheckCircle2,
  Loader2,
  RotateCcw,
  XCircle,
} from "lucide-react";
import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import type { MessageResponse } from "../api/types";
import { formatAgentLabel } from "../model";
import { SafeMarkdown } from "./SafeMarkdown";

export interface LiveToolCall {
  toolCallId: string;
  agent: string;
  toolName: string;
  status: "running" | "done" | "error";
}

export interface LiveTransfer {
  from: string;
  to: string;
}

interface ChatTimelineProps {
  messages: MessageResponse[];
  pendingUserMessage?: string | null;
  streaming: boolean;
  streamingAgent: string;
  streamingText: string;
  liveTransfers: LiveTransfer[];
  liveTools: LiveToolCall[];
  onRetry?: (message: string) => void;
}

export function ChatTimeline({
  messages,
  pendingUserMessage,
  streaming,
  streamingAgent,
  streamingText,
  liveTransfers,
  liveTools,
  onRetry,
}: ChatTimelineProps) {
  const { t } = useTranslation();
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages.length, pendingUserMessage, streamingText, liveTools.length]);

  const hasContent =
    messages.length > 0 || Boolean(pendingUserMessage) || streaming;

  if (!hasContent) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-3 px-6 text-center">
        <span className="flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
          <Bot aria-hidden className="size-7" />
        </span>
        <h2 className="text-lg font-semibold text-fg">
          {t("agent.emptyChatTitle")}
        </h2>
        <p className="max-w-md text-sm text-muted">
          {t("agent.emptyChatDescription")}
        </p>
      </div>
    );
  }

  let prevAssistantAgent = "gabriel";

  return (
    <div className="flex-1 overflow-y-auto px-4 py-6 sm:px-6">
      <div className="mx-auto flex max-w-3xl flex-col gap-4">
        {messages.map((msg, idx) => {
          const isLastMessage = idx === messages.length - 1;
          if (msg.role === "user") {
            prevAssistantAgent = "gabriel";
            return (
              <div key={msg.id} className="flex flex-col items-end gap-1">
                <div className="max-w-[85%] rounded-2xl rounded-br-sm bg-primary px-4 py-2.5 text-sm whitespace-pre-wrap text-on-primary shadow-xs">
                  {msg.content}
                </div>
                {isLastMessage && !streaming && onRetry ? (
                  <button
                    type="button"
                    onClick={() => onRetry(msg.content)}
                    className="inline-flex items-center gap-1 rounded-lg px-2 py-0.5 text-xs font-medium text-muted hover:bg-fg/8 hover:text-fg"
                  >
                    <RotateCcw aria-hidden className="size-3" />
                    {t("agent.retry")}
                  </button>
                ) : null}
              </div>
            );
          }

          if (msg.role === "tool") {
            return (
              <div key={msg.id} className="flex items-center gap-2 pl-2">
                <span className="inline-flex items-center gap-1.5 rounded-full border border-line bg-surface px-2.5 py-1 text-xs text-muted">
                  <CheckCircle2
                    aria-hidden
                    className="size-3.5 text-emerald-600 dark:text-emerald-400"
                  />
                  <span>{msg.toolName || "tool"}</span>
                </span>
              </div>
            );
          }

          const agent = msg.agentName || "gabriel";
          const showHandoff =
            agent !== prevAssistantAgent && agent !== "gabriel";
          const fromAgent = prevAssistantAgent;
          prevAssistantAgent = agent;

          return (
            <div key={msg.id} className="flex flex-col gap-2">
              {showHandoff ? (
                <div className="flex items-center gap-1.5 pl-2 text-xs text-muted">
                  <span className="inline-flex items-center gap-1 rounded-full border border-line bg-canvas px-2.5 py-0.5 font-medium">
                    {formatAgentLabel(fromAgent)}
                    <ArrowRight aria-hidden className="size-3" />
                    {formatAgentLabel(agent)}
                  </span>
                </div>
              ) : null}

              {msg.toolCalls && msg.toolCalls.length > 0 ? (
                <div className="flex flex-wrap gap-1.5 pl-2">
                  {msg.toolCalls.map((tc) => (
                    <span
                      key={tc.id}
                      className="inline-flex items-center gap-1 rounded-full border border-line bg-canvas px-2.5 py-0.5 text-xs text-muted"
                    >
                      {tc.name}
                    </span>
                  ))}
                </div>
              ) : null}

              {msg.content ? (
                <div className="flex flex-col items-start gap-1">
                  <span className="px-1 text-xs font-semibold text-muted">
                    {formatAgentLabel(agent)}
                  </span>
                  <div className="max-w-[90%] rounded-2xl rounded-tl-sm border border-line bg-surface px-4 py-3 shadow-xs">
                    <SafeMarkdown content={msg.content} />
                  </div>
                </div>
              ) : null}
            </div>
          );
        })}

        {pendingUserMessage ? (
          <div className="flex justify-end">
            <div className="max-w-[85%] rounded-2xl rounded-br-sm bg-primary px-4 py-2.5 text-sm whitespace-pre-wrap text-on-primary shadow-xs">
              {pendingUserMessage}
            </div>
          </div>
        ) : null}

        {streaming ? (
          <div className="flex flex-col gap-2">
            {liveTransfers.map((tr, idx) => (
              <div
                key={`${tr.from}-${tr.to}-${idx}`}
                className="flex items-center gap-1.5 pl-2 text-xs text-muted"
              >
                <span className="inline-flex items-center gap-1 rounded-full border border-line bg-canvas px-2.5 py-0.5 font-medium">
                  {formatAgentLabel(tr.from)}
                  <ArrowRight aria-hidden className="size-3" />
                  {formatAgentLabel(tr.to)}
                </span>
              </div>
            ))}

            {liveTools.length > 0 ? (
              <div className="flex flex-wrap gap-1.5 pl-2">
                {liveTools.map((tl) => (
                  <span
                    key={tl.toolCallId}
                    className="inline-flex items-center gap-1.5 rounded-full border border-line bg-surface px-2.5 py-1 text-xs text-muted"
                  >
                    {tl.status === "running" ? (
                      <Loader2 aria-hidden className="size-3.5 animate-spin" />
                    ) : tl.status === "error" ? (
                      <XCircle aria-hidden className="size-3.5 text-red-500" />
                    ) : (
                      <CheckCircle2
                        aria-hidden
                        className="size-3.5 text-emerald-600 dark:text-emerald-400"
                      />
                    )}
                    <span>{tl.toolName}</span>
                  </span>
                ))}
              </div>
            ) : null}

            <div className="flex flex-col items-start gap-1">
              <span className="px-1 text-xs font-semibold text-muted">
                {formatAgentLabel(streamingAgent)}
              </span>
              <div className="max-w-[90%] rounded-2xl rounded-tl-sm border border-line bg-surface px-4 py-3 shadow-xs">
                {streamingText ? (
                  <SafeMarkdown content={streamingText} />
                ) : (
                  <Loader2
                    aria-hidden
                    className="size-4 animate-spin text-muted"
                  />
                )}
              </div>
            </div>
          </div>
        ) : null}

        <div ref={bottomRef} />
      </div>
    </div>
  );
}
