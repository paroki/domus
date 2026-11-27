import { Check, Pencil, Plus, Search, Trash2, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import { useDeleteAgentSession, useUpdateAgentSession } from "../api/actions";
import { useAgentProviders, useAgentSessions } from "../api/queries";
import type { SessionDetailResponse } from "../api/types";

interface SessionSidebarProps {
  activeSession?: SessionDetailResponse;
  activeSessionId?: string;
  canManageProviders: boolean;
  mobileOpen: boolean;
  onCloseMobile: () => void;
  onSelectNew: () => void;
}

export function SessionSidebar({
  activeSession,
  activeSessionId,
  canManageProviders,
  mobileOpen,
  onCloseMobile,
  onSelectNew,
}: SessionSidebarProps) {
  const { t } = useTranslation();
  const [query, setQuery] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingTitle, setEditingTitle] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    title: string;
  } | null>(null);
  const renameInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editingId) {
      renameInputRef.current?.focus();
      renameInputRef.current?.select();
    }
  }, [editingId]);

  const { data: sessions = [] } = useAgentSessions(query);
  const { data: providers = [] } = useAgentProviders(canManageProviders);
  const updateSession = useUpdateAgentSession();
  const deleteSession = useDeleteAgentSession();

  useEffect(() => {
    if (!mobileOpen) return;
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") onCloseMobile();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [mobileOpen, onCloseMobile]);

  const modelOptions: { value: string; label: string }[] = [];
  if (canManageProviders) {
    for (const p of providers) {
      if (!p.id) continue;
      for (const m of p.models ?? []) {
        if (!m.id) continue;
        const caps = m.capabilities ?? [];
        if (caps.length > 0 && !caps.includes("chat")) continue;
        modelOptions.push({
          value: `${p.id}:${m.id}`,
          label: `${p.name ?? p.provider} — ${m.name || m.id}`,
        });
      }
    }
  }

  async function handleSaveTitle(sessionId: string) {
    const trimmed = editingTitle.trim();
    if (!trimmed) {
      setEditingId(null);
      return;
    }
    await updateSession.mutateAsync({
      sessionId,
      body: { title: trimmed },
    });
    setEditingId(null);
  }

  async function handleDelete(sessionId: string) {
    await deleteSession.mutateAsync(sessionId);
    if (sessionId === activeSessionId) {
      onSelectNew();
    }
  }

  async function handleModelChange(value: string) {
    if (!activeSession) return;
    if (!value) {
      await updateSession.mutateAsync({
        sessionId: activeSession.id,
        body: { providerId: "", modelId: "" },
      });
      return;
    }
    const idx = value.indexOf(":");
    if (idx === -1) return;
    await updateSession.mutateAsync({
      sessionId: activeSession.id,
      body: {
        providerId: value.slice(0, idx),
        modelId: value.slice(idx + 1),
      },
    });
  }

  const currentOverrideValue =
    activeSession?.providerId && activeSession?.modelId
      ? `${activeSession.providerId}:${activeSession.modelId}`
      : "";

  const sidebarContent = (
    <div className="flex h-full w-72 flex-col border-r border-line bg-surface">
      <div className="flex items-center justify-between gap-2 border-b border-line p-3">
        <button
          type="button"
          onClick={() => {
            onSelectNew();
            onCloseMobile();
          }}
          className="inline-flex min-h-11 flex-1 items-center justify-center gap-1.5 rounded-lg bg-primary px-3 py-2 text-xs font-semibold text-on-primary hover:opacity-90"
        >
          <Plus aria-hidden className="size-4" />
          {t("agent.newSession")}
        </button>
        <button
          type="button"
          aria-label={t("agent.cancel")}
          onClick={onCloseMobile}
          className="inline-flex size-11 items-center justify-center rounded-lg text-muted hover:bg-fg/8 hover:text-fg lg:hidden"
        >
          <X aria-hidden className="size-4" />
        </button>
      </div>

      <div className="border-b border-line p-3">
        <div className="relative">
          <Search
            aria-hidden
            className="pointer-events-none absolute top-1/2 left-3 size-3.5 -translate-y-1/2 text-muted"
          />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t("agent.searchSessionsPlaceholder")}
            className="h-11 w-full rounded-lg border border-line bg-bg pr-9 pl-8 text-base text-fg placeholder:text-muted/70 focus:border-accent focus:outline-none sm:text-sm"
          />
          {query ? (
            <button
              type="button"
              aria-label={t("agent.cancel")}
              onClick={() => setQuery("")}
              className="absolute top-1/2 right-1.5 inline-flex size-8 -translate-y-1/2 items-center justify-center rounded text-muted hover:text-fg"
            >
              <X aria-hidden className="size-3.5" />
            </button>
          ) : null}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-2">
        {sessions.length === 0 ? (
          <p className="px-2 py-4 text-center text-xs text-muted">
            {t("agent.emptySessions")}
          </p>
        ) : (
          <ul className="flex flex-col gap-1">
            {sessions.map((sess) => {
              const isActive = sess.id === activeSessionId;
              const isEditing = editingId === sess.id;
              return (
                <li
                  key={sess.id}
                  className={`group flex min-h-11 items-center gap-1 rounded-lg px-2.5 py-1.5 text-xs ${
                    isActive
                      ? "bg-primary/10 font-semibold text-fg"
                      : "text-muted hover:bg-fg/6 hover:text-fg"
                  }`}
                >
                  {isEditing ? (
                    <div className="flex flex-1 items-center gap-1">
                      <input
                        ref={renameInputRef}
                        type="text"
                        value={editingTitle}
                        onChange={(e) => setEditingTitle(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter") {
                            void handleSaveTitle(sess.id);
                          } else if (e.key === "Escape") {
                            setEditingId(null);
                          }
                        }}
                        className="h-10 min-w-0 flex-1 rounded border border-line bg-bg px-2 text-base text-fg focus:outline-none sm:text-sm"
                      />
                      <button
                        type="button"
                        aria-label={t("agent.renameSession")}
                        onClick={() => void handleSaveTitle(sess.id)}
                        className="inline-flex min-h-9 min-w-9 items-center justify-center rounded text-emerald-600 hover:bg-emerald-600/10"
                      >
                        <Check aria-hidden className="size-3.5" />
                      </button>
                    </div>
                  ) : (
                    <>
                      <Link
                        to={`/agent/sessions/${sess.id}`}
                        aria-current={isActive ? "page" : undefined}
                        onClick={onCloseMobile}
                        className="min-w-0 flex-1 truncate py-1.5 text-inherit no-underline"
                      >
                        {sess.title}
                      </Link>
                      <button
                        type="button"
                        aria-label={t("agent.renameSession")}
                        onClick={() => {
                          setDeleteTarget(null);
                          setEditingId(sess.id);
                          setEditingTitle(sess.title);
                        }}
                        className="inline-flex min-h-9 min-w-9 items-center justify-center rounded text-muted opacity-100 transition-opacity hover:text-fg sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100"
                      >
                        <Pencil aria-hidden className="size-3.5" />
                      </button>
                      <button
                        type="button"
                        aria-label={t("agent.deleteSession")}
                        onClick={() =>
                          setDeleteTarget({ id: sess.id, title: sess.title })
                        }
                        className="inline-flex min-h-9 min-w-9 items-center justify-center rounded text-muted opacity-100 transition-opacity hover:text-red-600 sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100"
                      >
                        <Trash2 aria-hidden className="size-3.5" />
                      </button>
                    </>
                  )}
                </li>
              );
            })}
          </ul>
        )}
      </div>

      {activeSession ? (
        <div className="border-t border-line p-3">
          <p className="mb-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted">
            {t("agent.modelOverrideLabel")}
          </p>
          {canManageProviders ? (
            <select
              aria-label={t("agent.modelOverrideLabel")}
              value={currentOverrideValue}
              disabled={updateSession.isPending}
              onChange={(e) => void handleModelChange(e.target.value)}
              className="h-11 w-full rounded-lg border border-line bg-bg px-2.5 text-base text-fg focus:border-accent focus:outline-none sm:text-sm"
            >
              <option value="">{t("agent.modelDefaultBadge")}</option>
              {modelOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          ) : (
            <div className="flex items-center justify-between gap-2">
              <span className="truncate rounded-md bg-canvas px-2 py-1 text-xs text-fg">
                {activeSession.modelId ||
                  activeSession.effectiveModelId ||
                  t("agent.modelDefaultBadge")}
              </span>
              {activeSession.modelId ? (
                <button
                  type="button"
                  disabled={updateSession.isPending}
                  onClick={() => void handleModelChange("")}
                  className="shrink-0 text-xs font-medium text-accent hover:underline"
                >
                  {t("agent.resetModelOverride")}
                </button>
              ) : null}
            </div>
          )}
        </div>
      ) : null}
    </div>
  );

  return (
    <>
      <aside className="hidden h-full shrink-0 lg:block">
        {sidebarContent}
      </aside>
      {mobileOpen ? (
        <div className="fixed inset-0 z-40 flex lg:hidden">
          <div className="relative z-10 h-full">{sidebarContent}</div>
          <button
            type="button"
            aria-label={t("agent.cancel")}
            onClick={onCloseMobile}
            className="flex-1 bg-black/40"
          />
        </div>
      ) : null}
      <ConfirmDialog
        open={Boolean(deleteTarget)}
        title={t("agent.confirmDeleteSession")}
        description={
          deleteTarget
            ? `${t("agent.confirmDeleteSessionDescription")} (${deleteTarget.title})`
            : t("agent.confirmDeleteSessionDescription")
        }
        confirmLabel={t("agent.deleteSession")}
        cancelLabel={t("agent.cancel")}
        pending={deleteSession.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          const id = deleteTarget.id;
          setDeleteTarget(null);
          void handleDelete(id);
        }}
        onCancel={() => setDeleteTarget(null)}
      />
    </>
  );
}
