import {
  ArrowDown,
  ArrowRight,
  ArrowUpRight,
  BookUser,
  Bot,
  Calendar,
  Check,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Church,
  Copy,
  Loader2,
  Music,
  RotateCcw,
  Sparkles,
  Users,
  XCircle,
} from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import type {
  MessageResponse,
  NavigationPayload,
  QuestionPayload,
} from "../api/types";
import {
  extractNavigationMeta,
  extractQuestionMeta,
  formatAgentLabel,
  formatToolJSON,
  isToolErrorContent,
} from "../model";
import { SafeMarkdown } from "./SafeMarkdown";

export interface LiveToolCall {
  toolCallId: string;
  agent: string;
  toolName: string;
  status: "running" | "done" | "error";
  args?: string;
  output?: string;
}

export interface LiveTransfer {
  from: string;
  to: string;
}

interface ChatTimelineProps {
  messages: MessageResponse[];
  loading?: boolean;
  pendingUserMessage?: string | null;
  streaming: boolean;
  streamingAgent: string;
  streamingText: string;
  liveTransfers: LiveTransfer[];
  liveTools: LiveToolCall[];
  liveQuestion?: QuestionPayload | null;
  liveNavigation?: NavigationPayload | null;
  verbose?: boolean;
  onRetry?: (message: string) => void;
  onSelectPrompt?: (prompt: string) => void;
  onSendMessage?: (message: string) => void;
}

function formatMessageTime(iso?: string): string {
  if (!iso) return "";
  try {
    const date = new Date(iso);
    if (Number.isNaN(date.getTime())) return "";
    return new Intl.DateTimeFormat("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  } catch {
    return "";
  }
}

export function AgentAvatar({ agent }: { agent: string }) {
  const key = agent.toLowerCase();
  if (key === "mikael") {
    return (
      <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-sky-500/15 text-sky-700 dark:text-sky-300">
        <Calendar aria-hidden className="size-3.5" />
      </span>
    );
  }
  if (key === "rafael") {
    return (
      <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-amber-500/15 text-amber-700 dark:text-amber-300">
        <Users aria-hidden className="size-3.5" />
      </span>
    );
  }
  if (key === "yosef") {
    return (
      <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-700 dark:text-emerald-300">
        <BookUser aria-hidden className="size-3.5" />
      </span>
    );
  }
  if (key === "gregorius") {
    return (
      <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-violet-500/15 text-violet-700 dark:text-violet-300">
        <Church aria-hidden className="size-3.5" />
      </span>
    );
  }
  if (key === "sesilia") {
    return (
      <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-rose-500/15 text-rose-700 dark:text-rose-300">
        <Music aria-hidden className="size-3.5" />
      </span>
    );
  }
  return (
    <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-primary/15 text-primary">
      <Bot aria-hidden className="size-3.5" />
    </span>
  );
}

function NavigationCard({ nav }: { nav: NavigationPayload }) {
  return (
    <Link
      to={nav.path}
      className="group/nav flex min-h-11 w-full max-w-md items-center justify-between gap-3 rounded-xl border border-primary/25 bg-primary/5 px-3.5 py-2.5 text-left transition-colors hover:border-primary/45 hover:bg-primary/10"
    >
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-primary">{nav.label}</p>
        {nav.description ? (
          <p className="mt-0.5 line-clamp-2 text-xs text-muted">
            {nav.description}
          </p>
        ) : null}
      </div>
      <ArrowUpRight
        aria-hidden
        className="size-4 shrink-0 text-primary transition-transform group-hover/nav:translate-x-0.5 group-hover/nav:-translate-y-0.5"
      />
    </Link>
  );
}

function QuestionChips({
  payload,
  onSendMessage,
}: {
  payload: QuestionPayload;
  onSendMessage: (message: string) => void;
}) {
  const { t } = useTranslation();
  const [selected, setSelected] = useState<string[]>([]);

  useEffect(() => {
    setSelected([]);
  }, [payload]);

  function toggleValue(val: string) {
    setSelected((prev) =>
      prev.includes(val) ? prev.filter((item) => item !== val) : [...prev, val],
    );
  }

  return (
    <div className="flex w-full max-w-xl flex-col gap-2 pt-1">
      <div className="flex flex-wrap gap-2">
        {payload.options.map((opt) => {
          const isSelected = selected.includes(opt.value);
          return (
            <button
              key={`${opt.value}-${opt.label}`}
              type="button"
              onClick={() => {
                if (payload.multiSelect === true) {
                  toggleValue(opt.value);
                } else {
                  onSendMessage(opt.value);
                }
              }}
              aria-pressed={
                payload.multiSelect === true ? isSelected : undefined
              }
              className={`inline-flex min-h-11 items-center gap-1.5 rounded-xl border px-3.5 py-2 text-left text-xs font-medium transition-colors ${
                isSelected
                  ? "border-primary bg-primary text-on-primary shadow-2xs"
                  : "border-line bg-surface text-fg hover:border-primary/40 hover:bg-primary/5"
              }`}
            >
              {payload.multiSelect === true && isSelected ? (
                <Check aria-hidden className="size-3.5 shrink-0" />
              ) : null}
              <span>{opt.label}</span>
            </button>
          );
        })}
      </div>

      {payload.multiSelect === true ? (
        <div>
          <button
            type="button"
            disabled={selected.length === 0}
            onClick={() => {
              if (selected.length > 0) {
                onSendMessage(selected.join(", "));
              }
            }}
            className="inline-flex min-h-11 items-center gap-1.5 rounded-xl bg-primary px-4 py-2 text-xs font-medium text-on-primary shadow-2xs transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-45"
          >
            {t("agent.questionSubmitSelected")}
          </button>
        </div>
      ) : null}
    </div>
  );
}

function CopyMessageButton({ content }: { content: string }) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  function handleCopy() {
    if (!content || typeof navigator === "undefined" || !navigator.clipboard) {
      return;
    }
    void navigator.clipboard.writeText(content).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }

  return (
    <button
      type="button"
      onClick={handleCopy}
      aria-label={copied ? t("agent.copiedMessage") : t("agent.copyMessage")}
      className="pointer-events-auto inline-flex min-h-9 items-center gap-1 rounded-lg px-2.5 py-1 text-xs font-medium text-muted opacity-100 transition-opacity hover:bg-fg/8 hover:text-fg focus:opacity-100 sm:pointer-events-none sm:opacity-0 sm:group-hover/msg:pointer-events-auto sm:group-hover/msg:opacity-100"
    >
      {copied ? (
        <>
          <Check
            aria-hidden
            className="size-3 text-emerald-600 dark:text-emerald-400"
          />
          {t("agent.copiedMessage")}
        </>
      ) : (
        <>
          <Copy aria-hidden className="size-3" />
          {t("agent.copyMessage")}
        </>
      )}
    </button>
  );
}

function CollapsibleToolItem({
  name,
  agentName,
  status,
  args,
  output,
  verbose,
}: {
  name: string;
  agentName?: string;
  status: "running" | "done" | "error";
  args?: string;
  output?: string;
  verbose: boolean;
}) {
  const { t } = useTranslation();
  const hasDetails = Boolean(args || output);
  const isError = status === "error";
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (isError || verbose) {
      setOpen(true);
    } else {
      setOpen(false);
    }
  }, [isError, verbose]);

  const pillClass = `inline-flex min-h-9 items-center gap-1.5 rounded-full border px-3 py-1 text-xs transition-colors ${
    isError
      ? "border-red-500/30 bg-red-500/10 text-red-700 dark:text-red-300"
      : "border-line bg-surface text-muted"
  } ${hasDetails ? "cursor-pointer hover:border-fg/30 hover:text-fg" : ""}`;

  const icon =
    status === "running" ? (
      <Loader2 aria-hidden className="size-3.5 animate-spin" />
    ) : isError ? (
      <XCircle aria-hidden className="size-3.5 text-red-500" />
    ) : (
      <CheckCircle2
        aria-hidden
        className="size-3.5 text-emerald-600 dark:text-emerald-400"
      />
    );

  return (
    <div className="flex flex-col items-start gap-1.5">
      {hasDetails ? (
        <button
          type="button"
          aria-expanded={open}
          onClick={() => setOpen((prev) => !prev)}
          className={pillClass}
        >
          {icon}
          <span className="font-mono text-[11px]">{name}</span>
          {verbose && agentName ? (
            <span className="opacity-60">({formatAgentLabel(agentName)})</span>
          ) : null}
          {open ? (
            <ChevronDown aria-hidden className="size-3 opacity-70" />
          ) : (
            <ChevronRight aria-hidden className="size-3 opacity-70" />
          )}
        </button>
      ) : (
        <span className={pillClass}>
          {icon}
          <span className="font-mono text-[11px]">{name}</span>
          {verbose && agentName ? (
            <span className="opacity-60">({formatAgentLabel(agentName)})</span>
          ) : null}
        </span>
      )}

      {open && hasDetails ? (
        <div className="flex w-full max-w-[90%] flex-col gap-1">
          {args ? (
            <pre className="max-h-60 overflow-auto rounded-xl border border-line bg-canvas px-3 py-2 font-mono text-xs whitespace-pre-wrap break-words text-muted">
              {`${t("agent.toolArgsLabel")}: ${formatToolJSON(args)}`}
            </pre>
          ) : null}
          {output ? (
            <pre
              className={`max-h-60 overflow-auto rounded-xl border px-3 py-2 font-mono text-xs whitespace-pre-wrap break-words ${
                isError
                  ? "border-red-500/30 bg-red-500/5 text-red-700 dark:text-red-300"
                  : "border-line bg-surface text-muted"
              }`}
            >
              {output}
            </pre>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

export function ChatTimeline({
  messages,
  loading = false,
  pendingUserMessage,
  streaming,
  streamingAgent,
  streamingText,
  liveTransfers,
  liveTools,
  liveQuestion,
  liveNavigation,
  verbose = false,
  onRetry,
  onSelectPrompt,
  onSendMessage,
}: ChatTimelineProps) {
  const { t } = useTranslation();
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const isNearBottomRef = useRef(true);
  const [showScrollBtn, setShowScrollBtn] = useState(false);

  function handleScroll() {
    const el = scrollContainerRef.current;
    if (!el) return;
    const distance = el.scrollHeight - el.scrollTop - el.clientHeight;
    const near = distance < 96;
    isNearBottomRef.current = near;
    setShowScrollBtn(!near);
  }

  function scrollToBottom(behavior: ScrollBehavior = "smooth") {
    isNearBottomRef.current = true;
    setShowScrollBtn(false);
    bottomRef.current?.scrollIntoView({ behavior });
  }

  useEffect(() => {
    if (pendingUserMessage) {
      isNearBottomRef.current = true;
      setShowScrollBtn(false);
    }
  }, [pendingUserMessage]);

  useEffect(() => {
    if (!isNearBottomRef.current) return;
    bottomRef.current?.scrollIntoView({
      behavior: streaming ? "auto" : "smooth",
    });
  }, [
    messages.length,
    pendingUserMessage,
    streamingText,
    liveTools.length,
    streaming,
  ]);

  const hasContent =
    messages.length > 0 || Boolean(pendingUserMessage) || streaming;

  if (loading && !hasContent) {
    return (
      <div className="flex-1 overflow-y-auto px-4 py-6 sm:px-6">
        <div className="mx-auto flex max-w-3xl animate-pulse flex-col gap-5">
          <div className="flex justify-end">
            <div className="h-10 w-52 rounded-2xl rounded-br-sm bg-fg/10" />
          </div>
          <div className="flex items-start gap-2.5">
            <div className="size-6 shrink-0 rounded-full bg-fg/10" />
            <div className="flex flex-1 flex-col gap-2">
              <div className="h-3 w-20 rounded bg-fg/10" />
              <div className="h-24 w-4/5 rounded-2xl rounded-tl-sm bg-fg/8" />
            </div>
          </div>
          <div className="flex justify-end">
            <div className="h-10 w-64 rounded-2xl rounded-br-sm bg-fg/10" />
          </div>
        </div>
      </div>
    );
  }

  if (!hasContent) {
    const starterPrompts = [
      t("agent.starterPrompt1"),
      t("agent.starterPrompt2"),
      t("agent.starterPrompt3"),
    ];
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-5 px-6 text-center">
        <span className="flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary shadow-2xs">
          <Bot aria-hidden className="size-7" />
        </span>
        <div className="space-y-1.5">
          <h2 className="text-lg font-semibold text-fg">
            {t("agent.emptyChatTitle")}
          </h2>
          <p className="max-w-md text-sm text-muted">
            {t("agent.emptyChatDescription")}
          </p>
        </div>
        {onSelectPrompt ? (
          <div className="mt-1 grid w-full max-w-xl grid-cols-1 gap-2.5 sm:grid-cols-3">
            {starterPrompts.map((prompt) => (
              <button
                key={prompt}
                type="button"
                onClick={() => onSelectPrompt(prompt)}
                className="group flex flex-col items-start gap-2 rounded-2xl border border-line bg-surface p-3.5 text-left text-xs font-medium text-fg shadow-2xs transition-all hover:-translate-y-0.5 hover:border-primary/40 hover:bg-primary/5 hover:shadow-xs"
              >
                <Sparkles
                  aria-hidden
                  className="size-4 text-primary transition-transform group-hover:scale-110"
                />
                <span className="leading-snug">{prompt}</span>
              </button>
            ))}
          </div>
        ) : null}
      </div>
    );
  }

  let prevAssistantAgent = "gabriel";
  let lastUserIdx = -1;
  for (let i = messages.length - 1; i >= 0; i--) {
    if (messages[i]?.role === "user") {
      lastUserIdx = i;
      break;
    }
  }

  return (
    <div className="relative flex min-h-0 flex-1 flex-col">
      <div
        ref={scrollContainerRef}
        onScroll={handleScroll}
        className="flex-1 overflow-y-auto px-4 py-6 sm:px-6"
      >
        <div className="mx-auto flex max-w-3xl flex-col gap-5">
          {messages.map((msg, idx) => {
            const isLastUserMessage = idx === lastUserIdx;
            const timeLabel = formatMessageTime(msg.createdAt);

            if (msg.role === "user") {
              prevAssistantAgent = "gabriel";
              return (
                <div
                  key={msg.id}
                  className="group/msg flex flex-col items-end gap-1"
                >
                  <div className="max-w-[85%] rounded-2xl rounded-br-sm bg-primary px-4 py-2.5 text-sm leading-relaxed whitespace-pre-wrap text-on-primary shadow-xs">
                    {msg.content}
                  </div>
                  <div className="flex items-center gap-2 px-1">
                    {timeLabel ? (
                      <span className="text-[11px] text-muted/80">
                        {timeLabel}
                      </span>
                    ) : null}
                    {isLastUserMessage && !streaming && onRetry ? (
                      <button
                        type="button"
                        onClick={() => onRetry(msg.content)}
                        className="inline-flex min-h-9 items-center gap-1 rounded-lg px-2.5 py-1 text-xs font-medium text-muted opacity-100 transition-opacity hover:bg-fg/8 hover:text-fg focus:opacity-100 sm:opacity-0 sm:group-hover/msg:opacity-100"
                      >
                        <RotateCcw aria-hidden className="size-3" />
                        {t("agent.retry")}
                      </button>
                    ) : null}
                  </div>
                </div>
              );
            }

            if (msg.role === "tool") {
              const isError = isToolErrorContent(msg.content);
              return (
                <div key={msg.id} className="pl-8">
                  <CollapsibleToolItem
                    name={msg.toolName || "tool"}
                    agentName={msg.agentName}
                    status={isError ? "error" : "done"}
                    output={msg.content || undefined}
                    verbose={verbose}
                  />
                </div>
              );
            }

            const agent = msg.agentName || "gabriel";
            const showHandoff =
              agent !== prevAssistantAgent && agent !== "gabriel";
            const fromAgent = prevAssistantAgent;
            prevAssistantAgent = agent;
            const navMeta = extractNavigationMeta(msg.meta);
            const isLastMessage = idx === messages.length - 1;
            const questionMeta =
              isLastMessage && !streaming
                ? (extractQuestionMeta(msg.meta) ?? liveQuestion ?? null)
                : null;

            return (
              <div key={msg.id} className="flex flex-col gap-2">
                {showHandoff ? (
                  <div className="flex items-center gap-1.5 pl-8 text-xs text-muted">
                    <span className="inline-flex items-center gap-1 rounded-full border border-line bg-canvas px-2.5 py-0.5 font-medium">
                      {formatAgentLabel(fromAgent)}
                      <ArrowRight aria-hidden className="size-3" />
                      {formatAgentLabel(agent)}
                    </span>
                  </div>
                ) : null}

                {msg.toolCalls && msg.toolCalls.length > 0 ? (
                  <div className="flex flex-col gap-1.5 pl-8">
                    {msg.toolCalls.map((tc) => (
                      <CollapsibleToolItem
                        key={tc.id}
                        name={tc.name}
                        agentName={msg.agentName}
                        status="done"
                        args={tc.arguments}
                        verbose={verbose}
                      />
                    ))}
                  </div>
                ) : null}

                {msg.content || navMeta || questionMeta ? (
                  <div className="group/msg flex items-start gap-2.5">
                    <AgentAvatar agent={agent} />
                    <div className="flex min-w-0 flex-1 flex-col items-start gap-1.5">
                      <div className="flex items-center gap-2 px-0.5">
                        <span className="text-xs font-semibold text-fg">
                          {formatAgentLabel(agent)}
                        </span>
                        {timeLabel ? (
                          <span className="text-[11px] text-muted/80">
                            {timeLabel}
                          </span>
                        ) : null}
                      </div>
                      {msg.content ? (
                        <div className="max-w-full rounded-2xl rounded-tl-sm border border-line bg-surface px-4 py-3 shadow-xs">
                          <SafeMarkdown content={msg.content} />
                        </div>
                      ) : null}
                      {navMeta ? <NavigationCard nav={navMeta} /> : null}
                      {questionMeta && onSendMessage ? (
                        <QuestionChips
                          payload={questionMeta}
                          onSendMessage={onSendMessage}
                        />
                      ) : null}
                      {msg.content ? (
                        <CopyMessageButton content={msg.content} />
                      ) : null}
                    </div>
                  </div>
                ) : null}
              </div>
            );
          })}

          {pendingUserMessage ? (
            <div className="flex justify-end">
              <div className="max-w-[85%] rounded-2xl rounded-br-sm bg-primary px-4 py-2.5 text-sm leading-relaxed whitespace-pre-wrap text-on-primary shadow-xs">
                {pendingUserMessage}
              </div>
            </div>
          ) : null}

          {streaming ? (
            <div className="flex flex-col gap-2">
              {liveTransfers.map((tr, idx) => (
                <div
                  key={`${tr.from}-${tr.to}-${idx}`}
                  className="flex items-center gap-1.5 pl-8 text-xs text-muted"
                >
                  <span className="inline-flex items-center gap-1 rounded-full border border-line bg-canvas px-2.5 py-0.5 font-medium">
                    {formatAgentLabel(tr.from)}
                    <ArrowRight aria-hidden className="size-3" />
                    {formatAgentLabel(tr.to)}
                  </span>
                </div>
              ))}

              {liveTools.length > 0 ? (
                <div className="flex flex-col gap-1.5 pl-8">
                  {liveTools.map((tl) => (
                    <CollapsibleToolItem
                      key={tl.toolCallId}
                      name={tl.toolName}
                      agentName={tl.agent}
                      status={tl.status}
                      args={tl.args}
                      output={tl.output}
                      verbose={verbose}
                    />
                  ))}
                </div>
              ) : null}

              <div className="flex items-start gap-2.5">
                <AgentAvatar agent={streamingAgent} />
                <div className="flex min-w-0 flex-1 flex-col items-start gap-1">
                  <div className="flex items-center gap-2 px-0.5">
                    <span className="text-xs font-semibold text-fg">
                      {formatAgentLabel(streamingAgent)}
                    </span>
                    <span className="inline-flex items-center gap-1 text-[11px] text-primary">
                      <span className="size-1.5 animate-pulse rounded-full bg-primary" />
                      {t("agent.typingLabel")}
                    </span>
                  </div>
                  <div className="max-w-full rounded-2xl rounded-tl-sm border border-line bg-surface px-4 py-3 shadow-xs">
                    {streamingText ? (
                      <div className="relative">
                        <SafeMarkdown content={streamingText} />
                        <span
                          aria-hidden
                          className="ml-1 inline-block h-3.5 w-1.5 animate-pulse rounded-xs bg-primary align-middle"
                        />
                      </div>
                    ) : (
                      <div className="flex items-center gap-1 py-1">
                        <span className="size-1.5 animate-bounce rounded-full bg-muted [animation-delay:-0.3s]" />
                        <span className="size-1.5 animate-bounce rounded-full bg-muted [animation-delay:-0.15s]" />
                        <span className="size-1.5 animate-bounce rounded-full bg-muted" />
                      </div>
                    )}
                  </div>
                  {liveNavigation ? (
                    <NavigationCard nav={liveNavigation} />
                  ) : null}
                </div>
              </div>
            </div>
          ) : null}

          <div ref={bottomRef} />
        </div>
      </div>

      {showScrollBtn ? (
        <div className="pointer-events-none absolute right-0 bottom-3 left-0 flex justify-center">
          <button
            type="button"
            onClick={() => scrollToBottom("smooth")}
            className="pointer-events-auto inline-flex min-h-10 items-center gap-1.5 rounded-full border border-line bg-surface/95 px-3.5 py-2 text-xs font-medium text-fg shadow-md backdrop-blur-xs transition-transform hover:scale-105 hover:bg-surface"
          >
            <ArrowDown aria-hidden className="size-3.5 text-primary" />
            {t("agent.scrollToBottom")}
          </button>
        </div>
      ) : null}
    </div>
  );
}
