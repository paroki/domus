import { Check, Pencil, Plus, Search, Trash2, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { useDeleteAgentSession, useUpdateAgentSession } from "../api/actions";
import type { AgentSessionResponse } from "../api/types";

interface SessionSidebarProps {
  sessions: AgentSessionResponse[];
  activeSessionId: string;
  searchQuery: string;
  onSearchChange: (q: string) => void;
  onSelectSession: (id: string) => void;
  onNewSession: () => void;
  hasNextPage: boolean;
  isFetchingNextPage: boolean;
  onLoadMore: () => void;
  onSessionDeleted: (id: string) => void;
}

function channelLabel(ch?: string): string {
  switch (ch) {
    case "telegram":
      return "Telegram";
    case "whatsapp":
      return "WhatsApp";
    default:
      return "Web";
  }
}

export function SessionSidebar({
  sessions,
  activeSessionId,
  searchQuery,
  onSearchChange,
  onSelectSession,
  onNewSession,
  hasNextPage,
  isFetchingNextPage,
  onLoadMore,
  onSessionDeleted,
}: SessionSidebarProps) {
  const { t } = useTranslation();
  const updateSession = useUpdateAgentSession();
  const deleteSession = useDeleteAgentSession();

  const [editingId, setEditingId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState("");

  async function submitRename(id: string) {
    const trimmed = editTitle.trim();
    if (!trimmed) return;
    await updateSession.mutateAsync({
      sessionId: id,
      body: { title: trimmed },
    });
    setEditingId(null);
  }

  async function handleDelete(id: string) {
    await deleteSession.mutateAsync(id);
    onSessionDeleted(id);
  }

  return (
    <aside className="flex h-full w-full flex-col gap-3 rounded-xl border border-line bg-surface p-3 lg:w-72 lg:shrink-0">
      <Button
        type="button"
        variant="primary"
        data-testid="agent-new-session"
        onClick={onNewSession}
        className="w-full"
      >
        <Plus className="size-4" />
        {t("agent.newSession")}
      </Button>

      <div className="relative">
        <Search className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted" />
        <input
          type="search"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder={t("agent.searchSessionsPlaceholder")}
          className="h-9 w-full rounded-lg border border-line bg-canvas pr-3 pl-8 text-xs text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
        />
      </div>

      <div className="flex flex-1 flex-col gap-1.5 overflow-y-auto pr-0.5">
        {sessions.length === 0 ? (
          <p className="px-2 py-4 text-center text-xs text-muted">
            {t("agent.emptySessions")}
          </p>
        ) : (
          sessions.map((sess) => {
            const id = sess.id ?? "";
            const isActive = id === activeSessionId;
            const isEditing = editingId === id;

            return (
              <div
                key={id}
                className={`group flex flex-col gap-1 rounded-lg border px-2.5 py-2 transition ${
                  isActive
                    ? "border-fg/30 bg-fg/8"
                    : "border-transparent hover:bg-fg/4"
                }`}
              >
                {isEditing ? (
                  <div className="flex items-center gap-1">
                    <input
                      type="text"
                      value={editTitle}
                      onChange={(e) => setEditTitle(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") {
                          e.preventDefault();
                          void submitRename(id);
                        } else if (e.key === "Escape") {
                          setEditingId(null);
                        }
                      }}
                      className="h-7 flex-1 rounded border border-line bg-canvas px-2 text-xs text-fg focus:outline-none"
                    />
                    <button
                      type="button"
                      aria-label="Simpan"
                      onClick={() => void submitRename(id)}
                      className="rounded p-1 text-emerald-600 hover:bg-fg/8"
                    >
                      <Check className="size-3.5" />
                    </button>
                    <button
                      type="button"
                      aria-label="Batal"
                      onClick={() => setEditingId(null)}
                      className="rounded p-1 text-muted hover:bg-fg/8"
                    >
                      <X className="size-3.5" />
                    </button>
                  </div>
                ) : (
                  <div className="flex items-start justify-between gap-2">
                    <button
                      type="button"
                      onClick={() => onSelectSession(id)}
                      className="flex-1 text-left"
                    >
                      <div className="line-clamp-1 text-xs font-medium text-fg">
                        {sess.title || t("agent.newSession")}
                      </div>
                      <div className="mt-1 flex items-center gap-1.5">
                        <Badge tone="neutral">
                          {channelLabel(sess.channel)}
                        </Badge>
                        {sess.lastMessageAt ? (
                          <span className="text-[10px] text-muted">
                            {new Date(sess.lastMessageAt).toLocaleDateString(
                              "id-ID",
                              { day: "numeric", month: "short" },
                            )}
                          </span>
                        ) : null}
                      </div>
                    </button>

                    <div className="flex shrink-0 items-center gap-0.5 opacity-0 transition group-hover:opacity-100 focus-within:opacity-100">
                      <button
                        type="button"
                        aria-label={t("agent.renameSession")}
                        onClick={() => {
                          setEditingId(id);
                          setEditTitle(sess.title ?? "");
                        }}
                        className="rounded p-1 text-muted hover:bg-fg/8 hover:text-fg"
                      >
                        <Pencil className="size-3" />
                      </button>
                      <button
                        type="button"
                        aria-label={t("agent.deleteSession")}
                        onClick={() => void handleDelete(id)}
                        className="rounded p-1 text-muted hover:bg-red-500/10 hover:text-red-600"
                      >
                        <Trash2 className="size-3" />
                      </button>
                    </div>
                  </div>
                )}
              </div>
            );
          })
        )}

        {hasNextPage ? (
          <Button
            type="button"
            variant="ghost"
            disabled={isFetchingNextPage}
            onClick={onLoadMore}
            className="h-8 px-2.5 text-xs"
          >
            {t("agent.loadMore")}
          </Button>
        ) : null}
      </div>
    </aside>
  );
}
