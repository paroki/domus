import { Pencil, Plus, Search, Trash2, X } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/shared/ui/Button";
import {
  useCreateAgentMemory,
  useDeleteAgentMemory,
  useUpdateAgentMemory,
} from "../api/actions";
import { useAgentMemories } from "../api/queries";
import {
  MAX_MEMORY_CHARS,
  MEMORY_CATEGORIES,
  type MemoryCategory,
  validateMemoryContent,
} from "../model";
import { SafeMarkdown } from "./SafeMarkdown";

interface MemoryDrawerProps {
  open: boolean;
  onClose: () => void;
}

export function MemoryDrawer({ open, onClose }: MemoryDrawerProps) {
  const { t } = useTranslation();
  const [query, setQuery] = useState("");
  const [filterCat, setFilterCat] = useState<string>("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [draftContent, setDraftContent] = useState("");
  const [draftCategory, setDraftCategory] = useState<MemoryCategory>("fact");
  const [formErr, setFormErr] = useState<string | null>(null);

  const { data: memories = [], isLoading } = useAgentMemories(
    query,
    filterCat,
    open,
  );
  const createMemory = useCreateAgentMemory();
  const updateMemory = useUpdateAgentMemory();
  const deleteMemory = useDeleteAgentMemory();

  useEffect(() => {
    if (!open) return;
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onClose]);

  if (!open) return null;

  function categoryLabel(cat: MemoryCategory): string {
    switch (cat) {
      case "preference":
        return t("agent.categoryPreference");
      case "context":
        return t("agent.categoryContext");
      default:
        return t("agent.categoryFact");
    }
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const err = validateMemoryContent(draftContent);
    if (err) {
      setFormErr(err);
      return;
    }
    setFormErr(null);
    const trimmed = draftContent.trim();
    if (editingId) {
      await updateMemory.mutateAsync({
        memoryId: editingId,
        body: { content: trimmed, category: draftCategory },
      });
      setEditingId(null);
    } else {
      await createMemory.mutateAsync({
        content: trimmed,
        category: draftCategory,
      });
    }
    setDraftContent("");
    setDraftCategory("fact");
  }

  return (
    <div className="fixed inset-0 z-50 flex justify-end">
      <button
        type="button"
        aria-label={t("agent.cancel")}
        onClick={onClose}
        className="flex-1 bg-black/40"
      />
      <section
        aria-label={t("agent.memoryDrawerTitle")}
        className="flex h-full w-full max-w-md flex-col border-l border-line bg-surface shadow-xl"
      >
        <div className="flex items-center justify-between border-b border-line px-4 py-3">
          <h2 className="text-base font-semibold text-fg">
            {t("agent.memoryDrawerTitle")}
          </h2>
          <button
            type="button"
            aria-label={t("agent.cancel")}
            onClick={onClose}
            className="rounded-lg p-1.5 text-muted hover:bg-fg/8 hover:text-fg"
          >
            <X aria-hidden className="size-4" />
          </button>
        </div>

        <form
          onSubmit={handleSubmit}
          className="flex flex-col gap-2 border-b border-line p-4"
        >
          <div className="flex items-center justify-between gap-2">
            <select
              value={draftCategory}
              onChange={(e) =>
                setDraftCategory(e.target.value as MemoryCategory)
              }
              className="rounded-lg border border-line bg-bg px-2.5 py-1.5 text-xs text-fg focus:border-accent focus:outline-none"
            >
              {MEMORY_CATEGORIES.map((c) => (
                <option key={c} value={c}>
                  {categoryLabel(c)}
                </option>
              ))}
            </select>
            {editingId ? (
              <button
                type="button"
                onClick={() => {
                  setEditingId(null);
                  setDraftContent("");
                  setFormErr(null);
                }}
                className="text-xs text-muted hover:text-fg"
              >
                {t("agent.cancel")}
              </button>
            ) : null}
          </div>

          <textarea
            rows={3}
            value={draftContent}
            maxLength={MAX_MEMORY_CHARS}
            onChange={(e) => {
              setDraftContent(e.target.value);
              if (formErr) setFormErr(null);
            }}
            placeholder={t("agent.memoryContentPlaceholder")}
            className="w-full rounded-lg border border-line bg-bg px-3 py-2 text-xs text-fg placeholder:text-muted/70 focus:border-accent focus:outline-none"
          />
          {draftContent.length > MAX_MEMORY_CHARS * 0.8 ? (
            <p className="text-right text-[11px] tabular-nums text-muted">
              {draftContent.length} / {MAX_MEMORY_CHARS}
            </p>
          ) : null}
          {formErr ? <p className="text-xs text-red-600">{formErr}</p> : null}
          <Button
            type="submit"
            disabled={createMemory.isPending || updateMemory.isPending}
            className="h-9 gap-1.5 text-xs"
          >
            <Plus aria-hidden className="size-3.5" />
            {editingId ? t("agent.saveMemory") : t("agent.memoryAdd")}
          </Button>
        </form>

        <div className="flex items-center gap-2 border-b border-line px-4 py-2.5">
          <div className="relative flex-1">
            <Search
              aria-hidden
              className="pointer-events-none absolute top-2 left-2.5 size-3.5 text-muted"
            />
            <input
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder={t("agent.memorySearchPlaceholder")}
              className="w-full rounded-lg border border-line bg-bg py-1.5 pr-2.5 pl-8 text-xs text-fg placeholder:text-muted/70 focus:border-accent focus:outline-none"
            />
          </div>
          <select
            value={filterCat}
            onChange={(e) => setFilterCat(e.target.value)}
            className="rounded-lg border border-line bg-bg px-2 py-1.5 text-xs text-fg focus:border-accent focus:outline-none"
          >
            <option value="">{t("agent.categoryAll")}</option>
            {MEMORY_CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {categoryLabel(c)}
              </option>
            ))}
          </select>
        </div>

        <div className="flex-1 overflow-y-auto p-4">
          {isLoading ? (
            <p className="text-xs text-muted">{t("common.loading")}</p>
          ) : memories.length === 0 ? (
            <p className="py-6 text-center text-xs text-muted">
              {t("agent.memoryEmpty")}
            </p>
          ) : (
            <ul className="flex flex-col gap-3">
              {memories.map((mem) => (
                <li
                  key={mem.id}
                  className="rounded-xl border border-line bg-bg p-3"
                >
                  <div className="mb-1.5 flex items-center justify-between gap-2">
                    <span className="rounded-full bg-primary/10 px-2 py-0.5 text-[11px] font-medium text-primary">
                      {categoryLabel(mem.category)}
                    </span>
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        aria-label={t("agent.editMemory")}
                        onClick={() => {
                          setEditingId(mem.id);
                          setDraftContent(mem.content);
                          setDraftCategory(mem.category);
                          setFormErr(null);
                        }}
                        className="rounded p-1 text-muted hover:text-fg"
                      >
                        <Pencil aria-hidden className="size-3.5" />
                      </button>
                      <button
                        type="button"
                        aria-label={t("agent.deleteSession")}
                        disabled={deleteMemory.isPending}
                        onClick={() => void deleteMemory.mutateAsync(mem.id)}
                        className="rounded p-1 text-muted hover:text-red-600 disabled:opacity-50"
                      >
                        <Trash2 aria-hidden className="size-3.5" />
                      </button>
                    </div>
                  </div>
                  <div className="text-xs">
                    <SafeMarkdown content={mem.content} />
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>
    </div>
  );
}
