import { Bot, Search } from "lucide-react";
import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import {
  agentDisplayMeta,
  type CorrelatedToolStep,
  parseContentSegments,
  type TimelineItem,
  toolDisplayMeta,
} from "../model";
import { A2uiSurfaceBlock } from "./A2uiSurfaceBlock";
import { MarkdownContent } from "./MarkdownContent";

export interface LiveStreamSegment {
  agentName: string;
  thinking: string;
  delta: string;
  toolSteps: CorrelatedToolStep[];
}

export interface LiveStreamingState {
  active: boolean;
  userMessage?: string;
  baselineUserCount?: number;
  segments: LiveStreamSegment[];
}

interface ChatTimelineProps {
  items: TimelineItem[];
  streaming: LiveStreamingState;
  messageSearch: string;
  onMessageSearchChange: (q: string) => void;
  hasOlder: boolean;
  isLoadingOlder: boolean;
  onLoadOlder: () => void;
  onStarterPrompt: (prompt: string) => void;
  onA2uiAction?: (actionPrompt: string) => void;
}

function formatJsonPreview(raw?: string): string {
  if (!raw) return "";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

function ToolStepsList({ steps }: { steps: CorrelatedToolStep[] }) {
  if (steps.length === 0) return null;
  return (
    <div className="flex flex-col gap-1.5">
      {steps.map((step) => {
        const meta = toolDisplayMeta(step.name);
        const statusPrefix =
          step.status === "running"
            ? "⏳ Memanggil"
            : step.status === "error"
              ? "⚠ Gagal"
              : "✓ Selesai";
        return (
          <details
            key={step.id}
            className="rounded-lg border border-line/80 bg-canvas/70 px-3 py-1.5 text-xs"
          >
            <summary className="flex cursor-pointer flex-wrap items-center gap-2 select-none">
              <span className="font-medium text-fg">
                {statusPrefix}: <span className="font-mono">{step.name}</span>
              </span>
              <Badge tone="neutral">{meta.category}</Badge>
            </summary>
            <div className="mt-2 flex flex-col gap-1.5">
              {step.arguments ? (
                <pre className="overflow-x-auto rounded bg-surface p-2 font-mono text-[11px] text-muted">
                  {formatJsonPreview(step.arguments)}
                </pre>
              ) : null}
              {step.result ? (
                <pre className="max-h-48 overflow-auto rounded bg-surface p-2 font-mono text-[11px] text-fg">
                  {formatJsonPreview(step.result)}
                </pre>
              ) : null}
            </div>
          </details>
        );
      })}
    </div>
  );
}

function MessageBody({
  content,
  isStreaming = false,
  onA2uiAction,
}: {
  content: string;
  isStreaming?: boolean;
  onA2uiAction?: (actionPrompt: string) => void;
}) {
  const segments = parseContentSegments(content, isStreaming);
  if (segments.length === 0) return null;

  return (
    <div className="flex max-w-[90%] flex-col gap-2">
      {segments.map((seg, idx) => {
        if (seg.type === "markdown") {
          return (
            <div
              key={`md-${idx}`}
              className="rounded-2xl rounded-bl-sm border border-line bg-canvas px-4 py-2.5"
            >
              <MarkdownContent content={seg.content} />
            </div>
          );
        }
        if (seg.type === "a2ui-pending") {
          return (
            <div
              key={`a2ui-pending-${idx}`}
              className="rounded-xl border border-line bg-surface px-3 py-2 text-xs text-muted"
            >
              ⏳ Menyiapkan tampilan interaktif...
            </div>
          );
        }
        return (
          <A2uiSurfaceBlock
            key={`a2ui-${idx}`}
            messages={seg.messages}
            onAction={onA2uiAction}
          />
        );
      })}
    </div>
  );
}

function AssistantSegment({
  agentName,
  thinking,
  toolSteps,
  content,
  isStreaming = false,
  onA2uiAction,
}: {
  agentName: string;
  thinking?: string;
  toolSteps: CorrelatedToolStep[];
  content?: string;
  isStreaming?: boolean;
  onA2uiAction?: (actionPrompt: string) => void;
}) {
  const { t } = useTranslation();
  const meta = agentDisplayMeta(agentName);
  const isSubagent = agentName.toLowerCase() !== "gabriel";

  return (
    <div className="flex flex-col items-start gap-2">
      {!isSubagent ? (
        <div className="flex items-center gap-2">
          <Badge tone={meta.tone}>{meta.name}</Badge>
          <span className="text-[11px] text-muted">{meta.roleLabel}</span>
        </div>
      ) : null}

      {thinking ? (
        <details
          open={isStreaming ? true : undefined}
          className="w-full max-w-[90%] rounded-lg border border-line/70 bg-canvas/50 px-3 py-1.5 text-xs"
        >
          <summary className="cursor-pointer font-medium text-muted select-none hover:text-fg">
            {isSubagent
              ? t("agent.subagentThinkingLabel", { agent: meta.name })
              : t("agent.thinkingLabel", { agent: meta.name })}
          </summary>
          <p className="mt-1.5 whitespace-pre-wrap text-muted">{thinking}</p>
        </details>
      ) : null}

      {toolSteps.length > 0 ? (
        <div className="w-full max-w-[90%]">
          <ToolStepsList steps={toolSteps} />
        </div>
      ) : null}

      {content ? (
        <MessageBody
          content={content}
          isStreaming={isStreaming}
          onA2uiAction={onA2uiAction}
        />
      ) : null}
    </div>
  );
}

const STARTER_PROMPTS = [
  "Apa saja jadwal kegiatan minggu ini?",
  "Daftar organisasi yang saya ikuti",
  "Ingat bahwa saya suka dipanggil dengan ramah",
];

export function ChatTimeline({
  items,
  streaming,
  messageSearch,
  onMessageSearchChange,
  hasOlder,
  isLoadingOlder,
  onLoadOlder,
  onStarterPrompt,
  onA2uiAction,
}: ChatTimelineProps) {
  const { t } = useTranslation();
  const bottomRef = useRef<HTMLDivElement>(null);

  const lastSeg = streaming.segments[streaming.segments.length - 1];
  const scrollSignal = `${streaming.segments.length}:${lastSeg?.thinking.length ?? 0}:${lastSeg?.delta.length ?? 0}:${lastSeg?.toolSteps.length ?? 0}`;

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [items.length, scrollSignal]);

  const isEmpty = items.length === 0 && !streaming.active;
  const userItemCount = items.filter((item) => item.role === "user").length;
  const showStreamingUserMessage =
    streaming.active &&
    Boolean(streaming.userMessage) &&
    userItemCount <= (streaming.baselineUserCount ?? 0);

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      <div className="flex items-center justify-between gap-2 border-b border-line px-4 py-2">
        <div className="relative flex-1 max-w-xs">
          <Search className="pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted" />
          <input
            type="search"
            value={messageSearch}
            onChange={(e) => onMessageSearchChange(e.target.value)}
            placeholder={t("agent.searchMessagesPlaceholder")}
            className="h-8 w-full rounded-lg border border-line bg-canvas pr-3 pl-8 text-xs text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
          />
        </div>
        {hasOlder ? (
          <Button
            type="button"
            variant="ghost"
            disabled={isLoadingOlder}
            onClick={onLoadOlder}
            className="h-8 px-2.5 text-xs"
          >
            {t("agent.loadOlder")}
          </Button>
        ) : null}
      </div>

      <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-4">
        {isEmpty ? (
          <div className="my-auto flex flex-col items-center gap-4 text-center">
            <div className="flex size-12 items-center justify-center rounded-2xl border border-line bg-canvas text-fg">
              <Bot className="size-6" />
            </div>
            <div>
              <h2 className="text-base font-semibold text-fg">
                {t("agent.emptyChatTitle")}
              </h2>
              <p className="mt-1 text-xs text-muted">
                {t("agent.emptyChatDescription")}
              </p>
            </div>
            <div className="flex flex-wrap justify-center gap-2">
              {STARTER_PROMPTS.map((prompt) => (
                <button
                  key={prompt}
                  type="button"
                  onClick={() => onStarterPrompt(prompt)}
                  className="rounded-full border border-line bg-canvas px-3 py-1.5 text-xs text-fg transition hover:bg-fg/6"
                >
                  {prompt}
                </button>
              ))}
            </div>
          </div>
        ) : null}

        {items.map((item) => {
          if (item.role === "user") {
            return (
              <div key={item.id} className="flex justify-end">
                <div className="max-w-[82%] rounded-2xl rounded-br-sm bg-primary px-4 py-2.5 text-sm whitespace-pre-wrap text-on-primary">
                  {item.content}
                </div>
              </div>
            );
          }

          return (
            <AssistantSegment
              key={item.id}
              agentName={item.agentName}
              thinking={item.thinking}
              toolSteps={item.toolSteps}
              content={item.content}
              onA2uiAction={onA2uiAction}
            />
          );
        })}

        {showStreamingUserMessage ? (
          <div className="flex justify-end">
            <div className="max-w-[82%] rounded-2xl rounded-br-sm bg-primary px-4 py-2.5 text-sm whitespace-pre-wrap text-on-primary">
              {streaming.userMessage}
            </div>
          </div>
        ) : null}

        {streaming.active
          ? streaming.segments.map((seg, idx) => (
              <AssistantSegment
                key={`live-${seg.agentName}-${idx}`}
                agentName={seg.agentName}
                thinking={seg.thinking}
                toolSteps={seg.toolSteps}
                content={seg.delta}
                isStreaming
                onA2uiAction={onA2uiAction}
              />
            ))
          : null}

        <div ref={bottomRef} />
      </div>
    </div>
  );
}
