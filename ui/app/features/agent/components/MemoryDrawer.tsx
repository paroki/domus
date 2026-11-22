import { Pencil, Pin, Plus, Search, Trash2, X } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMe } from "~/shared/api/session";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import {
  useCreateAgentMemory,
  useDeleteAgentMemory,
  useUpdateAgentMemory,
} from "../api/actions";
import { resolveAgentError } from "../api/errors";
import { useAgentMemories } from "../api/queries";
import type { AgentMemoryResponse } from "../api/types";
import {
  canMutateMemory,
  isParishAdmin,
  MAX_MEMORY_CONTENT_LEN,
  MAX_PINNED_MEMORIES,
  MEMORY_CATEGORIES,
  MEMORY_SCOPES,
  memoryCategoryMeta,
} from "../model";

interface MemoryDrawerProps {
  open: boolean;
  onClose: () => void;
}

type MemoryScope = "user" | "parish";
type MemoryCategory = "profile" | "preference" | "ministry" | "parish_faq";

export function MemoryDrawer({ open, onClose }: MemoryDrawerProps) {
  const { t } = useTranslation();
  const { data: me } = useMe();
  const hasParish = Boolean(me?.activeParishId);
  const parishAdmin = isParishAdmin(me?.roles);

  const [scope, setScope] = useState<MemoryScope>("user");
  const [q, setQ] = useState("");
  const [categoryFilter, setCategoryFilter] = useState<string>("all");
  const [pinnedOnly, setPinnedOnly] = useState(false);

  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<AgentMemoryResponse | null>(null);
  const [category, setCategory] = useState<MemoryCategory>("preference");
  const [content, setContent] = useState("");
  const [isPinned, setIsPinned] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const memoriesQuery = useAgentMemories(
    {
      q: q.trim() || undefined,
      filter: {
        scope,
        category:
          categoryFilter !== "all"
            ? (categoryFilter as MemoryCategory)
            : undefined,
        isPinned: pinnedOnly ? true : undefined,
      },
    },
    open && hasParish,
  );

  const createMemory = useCreateAgentMemory();
  const updateMemory = useUpdateAgentMemory();
  const deleteMemory = useDeleteAgentMemory();

  if (!open) return null;

  const items = memoriesQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const pinnedCount = items.filter((m) => m.isPinned).length;
  const canCreateInScope = scope === "user" || parishAdmin;

  function openCreate() {
    setEditing(null);
    setCategory(scope === "parish" ? "parish_faq" : "preference");
    setContent("");
    setIsPinned(false);
    setFormError(null);
    setShowForm(true);
  }

  function openEdit(m: AgentMemoryResponse) {
    setEditing(m);
    setCategory((m.category as MemoryCategory) || "preference");
    setContent(m.content ?? "");
    setIsPinned(Boolean(m.isPinned));
    setFormError(null);
    setShowForm(true);
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setFormError(null);
    const trimmed = content.trim();
    if (!trimmed) return;
    try {
      if (editing?.id) {
        await updateMemory.mutateAsync({
          memoryId: editing.id,
          body: { category, content: trimmed, isPinned },
        });
      } else {
        await createMemory.mutateAsync({
          scope,
          category,
          content: trimmed,
          isPinned,
        });
      }
      setShowForm(false);
      setEditing(null);
      setContent("");
    } catch (err) {
      setFormError(resolveAgentError(err));
    }
  }

  async function handleTogglePin(m: AgentMemoryResponse) {
    if (!m.id) return;
    setFormError(null);
    try {
      await updateMemory.mutateAsync({
        memoryId: m.id,
        body: { isPinned: !m.isPinned },
      });
    } catch (err) {
      setFormError(resolveAgentError(err));
    }
  }

  async function handleDelete(id: string) {
    setFormError(null);
    try {
      await deleteMemory.mutateAsync(id);
    } catch (err) {
      setFormError(resolveAgentError(err));
    }
  }

  return (
    <aside className="flex h-full w-full flex-col gap-3 rounded-xl border border-line bg-surface p-4 lg:w-96 lg:shrink-0">
      <div className="flex items-center justify-between gap-2">
        <h2 className="text-sm font-semibold text-fg">
          {t("agent.memoryDrawerTitle")}
        </h2>
        <button
          type="button"
          onClick={onClose}
          aria-label="Tutup"
          className="rounded-lg p-1 text-muted hover:bg-fg/6 hover:text-fg"
        >
          <X className="size-4" />
        </button>
      </div>

      <div
        role="tablist"
        aria-label={t("agent.memoryDrawerTitle")}
        className="grid grid-cols-2 gap-1 rounded-lg bg-canvas p-1"
      >
        {MEMORY_SCOPES.map((s) => {
          const selected = scope === s.value;
          return (
            <button
              key={s.value}
              type="button"
              role="tab"
              aria-selected={selected}
              onClick={() => {
                setScope(s.value);
                setShowForm(false);
                setFormError(null);
              }}
              className={`rounded-md px-2 py-1.5 text-xs font-medium transition ${
                selected
                  ? "bg-surface text-fg shadow-xs"
                  : "text-muted hover:text-fg"
              }`}
            >
              {s.label}
            </button>
          );
        })}
      </div>

      <div className="flex items-center justify-between gap-2 text-xs">
        <Badge tone="neutral">
          {t("agent.memoryPinnedCount", {
            count: Math.min(pinnedCount, MAX_PINNED_MEMORIES),
          })}
        </Badge>
        {canCreateInScope ? (
          <Button
            type="button"
            onClick={openCreate}
            className="h-8 px-2.5 text-xs"
          >
            <Plus className="size-3.5" />
            {t("agent.memoryAdd")}
          </Button>
        ) : null}
      </div>

      {scope === "parish" && !parishAdmin ? (
        <div className="rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-xs text-amber-800 dark:text-amber-200">
          {t("agent.memoryParishReadOnly")}
        </div>
      ) : null}

      <div className="flex flex-col gap-2">
        <div className="relative">
          <Search className="pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted" />
          <input
            type="search"
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder={t("agent.memorySearchPlaceholder")}
            className="h-8 w-full rounded-lg border border-line bg-canvas pr-2.5 pl-8 text-xs text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
          />
        </div>
        <div className="flex items-center gap-2">
          <select
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
            className="h-8 flex-1 rounded-lg border border-line bg-canvas px-2 text-xs text-fg focus:outline-none"
          >
            <option value="all">Semua Kategori</option>
            {MEMORY_CATEGORIES.map((c) => (
              <option key={c.value} value={c.value}>
                {c.label}
              </option>
            ))}
          </select>
          <label className="flex cursor-pointer items-center gap-1.5 text-xs text-muted">
            <input
              type="checkbox"
              checked={pinnedOnly}
              onChange={(e) => setPinnedOnly(e.target.checked)}
            />
            Disematkan
          </label>
        </div>
      </div>

      {formError ? (
        <div
          role="alert"
          className="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs text-red-700 dark:text-red-300"
        >
          {formError}
        </div>
      ) : null}

      {showForm && canCreateInScope ? (
        <form
          onSubmit={handleSubmit}
          className="flex flex-col gap-2.5 rounded-lg border border-line bg-canvas p-3"
        >
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value as MemoryCategory)}
            className="h-8 rounded-lg border border-line bg-surface px-2 text-xs text-fg"
          >
            {MEMORY_CATEGORIES.map((c) => (
              <option key={c.value} value={c.value}>
                {c.label}
              </option>
            ))}
          </select>
          <textarea
            rows={3}
            maxLength={MAX_MEMORY_CONTENT_LEN}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Tulis fakta ringkas..."
            className="rounded-lg border border-line bg-surface p-2 text-xs text-fg focus:outline-none"
          />
          <div className="flex items-center justify-between text-[11px] text-muted">
            <label className="flex cursor-pointer items-center gap-1.5">
              <input
                type="checkbox"
                checked={isPinned}
                onChange={(e) => setIsPinned(e.target.checked)}
              />
              Sematkan (maks 5)
            </label>
            <span>
              {content.length}/{MAX_MEMORY_CONTENT_LEN}
            </span>
          </div>
          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setShowForm(false)}
              className="h-8 px-2.5 text-xs"
            >
              Batal
            </Button>
            <Button
              type="submit"
              disabled={createMemory.isPending || updateMemory.isPending}
              className="h-8 px-2.5 text-xs"
            >
              Simpan
            </Button>
          </div>
        </form>
      ) : null}

      <div className="flex flex-1 flex-col gap-2 overflow-y-auto pr-0.5">
        {items.length === 0 ? (
          <p className="py-6 text-center text-xs text-muted">
            {t("agent.memoryEmpty")}
          </p>
        ) : (
          items.map((m) => {
            const cat = memoryCategoryMeta(m.category);
            const mutable = canMutateMemory(m, me);
            return (
              <div
                key={m.id}
                className="flex flex-col gap-1.5 rounded-lg border border-line bg-canvas p-2.5"
              >
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-1.5">
                    <Badge tone={cat.tone}>{cat.label}</Badge>
                    {m.isPinned ? <Badge tone="success">Pin</Badge> : null}
                  </div>
                  {mutable && m.id ? (
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        aria-label="Pin"
                        onClick={() => void handleTogglePin(m)}
                        className={`rounded p-1 hover:bg-fg/8 ${
                          m.isPinned ? "text-emerald-600" : "text-muted"
                        }`}
                      >
                        <Pin className="size-3.5" />
                      </button>
                      <button
                        type="button"
                        aria-label="Ubah"
                        onClick={() => openEdit(m)}
                        className="rounded p-1 text-muted hover:bg-fg/8 hover:text-fg"
                      >
                        <Pencil className="size-3.5" />
                      </button>
                      <button
                        type="button"
                        aria-label="Hapus"
                        onClick={() => void handleDelete(m.id ?? "")}
                        className="rounded p-1 text-muted hover:bg-red-500/10 hover:text-red-600"
                      >
                        <Trash2 className="size-3.5" />
                      </button>
                    </div>
                  ) : null}
                </div>
                <p className="text-xs leading-relaxed text-fg">{m.content}</p>
              </div>
            );
          })
        )}

        {memoriesQuery.hasNextPage ? (
          <Button
            type="button"
            variant="ghost"
            disabled={memoriesQuery.isFetchingNextPage}
            onClick={() => void memoriesQuery.fetchNextPage()}
            className="h-8 px-2.5 text-xs"
          >
            {t("agent.loadMore")}
          </Button>
        ) : null}
      </div>
    </aside>
  );
}
