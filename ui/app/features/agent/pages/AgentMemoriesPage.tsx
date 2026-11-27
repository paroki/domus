import { Brain, Pencil, Plus, Search, Trash2 } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { useMe } from "~/shared/api/session";
import { buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import {
  InductiveEmptyState,
  InductiveFilterBar,
  InductivePageHeader,
} from "~/shared/ui/iui";
import { useDeleteAgentMemory } from "../api/actions";
import { useAgentMemories } from "../api/queries";
import type { MemoryResponse } from "../api/types";
import { SafeMarkdown } from "../components/SafeMarkdown";
import { MEMORY_CATEGORIES, type MemoryCategory } from "../model";

export function meta() {
  return [{ title: "Memori AI — Domus" }];
}

export default function AgentMemoriesPage() {
  const { t } = useTranslation();
  const { data: me } = useMe();
  const hasParish = Boolean(me?.activeParokiId);

  const [query, setQuery] = useState("");
  const [filterCat, setFilterCat] = useState<string>("");
  const [deleteTarget, setDeleteTarget] = useState<MemoryResponse | null>(null);

  const { data: memories = [], isLoading } = useAgentMemories(
    query,
    filterCat,
    hasParish,
  );
  const deleteMemory = useDeleteAgentMemory();

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

  const filterChips: { value: string; label: string }[] = [
    { value: "", label: t("agent.categoryAll") },
    { value: "fact", label: t("agent.categoryFact") },
    { value: "preference", label: t("agent.categoryPreference") },
    { value: "context", label: t("agent.categoryContext") },
  ];

  // Ensure all MEMORY_CATEGORIES remain represented
  void MEMORY_CATEGORIES;

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: t("agent.title"), to: "/agent" },
          { label: t("agent.memoriesBreadcrumb") },
        ]}
        title={t("agent.memoryDrawerTitle")}
        subtitle={t("agent.memoriesSubtitle")}
        action={
          <Link
            to="/agent/memories/new"
            className={`${buttonClass("primary")} min-h-11 w-full sm:w-auto`}
          >
            <Plus aria-hidden className="size-4" />
            {t("agent.memoryAdd")}
          </Link>
        }
      />

      {me && !hasParish ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          {t("agent.noParishDescription")}
        </div>
      ) : (
        <>
          <InductiveFilterBar className="flex-col items-stretch sm:flex-row sm:items-center">
            <div className="relative flex-1">
              <Search
                aria-hidden
                className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted"
              />
              <input
                type="search"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder={t("agent.memorySearchPlaceholder")}
                className="h-11 w-full rounded-lg border border-line bg-canvas pr-3 pl-9 text-base text-fg placeholder:text-muted/70 focus:border-primary focus:outline-none sm:text-sm"
              />
            </div>
            <div
              role="group"
              aria-label={t("agent.memoryCategoryLabel")}
              className="flex flex-wrap items-center gap-2"
            >
              {filterChips.map((chip) => {
                const active = filterCat === chip.value;
                return (
                  <button
                    key={chip.value || "all"}
                    type="button"
                    aria-pressed={active}
                    onClick={() => setFilterCat(chip.value)}
                    className={`inline-flex min-h-10 items-center rounded-lg border px-3 text-xs font-medium transition-colors ${
                      active
                        ? "border-primary bg-primary text-on-primary"
                        : "border-line bg-canvas text-muted hover:bg-fg/6 hover:text-fg"
                    }`}
                  >
                    {chip.label}
                  </button>
                );
              })}
            </div>
          </InductiveFilterBar>

          {isLoading ? (
            <p className="text-sm text-muted">{t("common.loading")}</p>
          ) : memories.length === 0 ? (
            <InductiveEmptyState
              icon={Brain}
              title={t("agent.memoryEmpty")}
              description={t("agent.memoriesSubtitle")}
              action={
                <Link
                  to="/agent/memories/new"
                  className={`${buttonClass("primary")} min-h-11 w-full sm:w-auto`}
                >
                  <Plus aria-hidden className="size-4" />
                  {t("agent.memoryAdd")}
                </Link>
              }
            />
          ) : (
            <ul className="flex flex-col gap-3">
              {memories.map((mem) => (
                <li key={mem.id}>
                  <Card className="flex flex-col gap-3 p-4 sm:p-6">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <span className="rounded-full bg-primary/10 px-2.5 py-1 text-xs font-medium text-primary">
                        {categoryLabel(mem.category)}
                      </span>
                      <div className="flex items-center gap-2">
                        <Link
                          to={`/agent/memories/${mem.id}/edit`}
                          className="inline-flex min-h-10 items-center gap-1.5 rounded-lg border border-line bg-canvas px-3 text-xs font-medium text-fg no-underline transition-colors hover:bg-fg/6"
                        >
                          <Pencil aria-hidden className="size-3.5" />
                          {t("agent.editMemory")}
                        </Link>
                        <button
                          type="button"
                          disabled={deleteMemory.isPending}
                          onClick={() => setDeleteTarget(mem)}
                          className="inline-flex min-h-10 items-center gap-1.5 rounded-lg border border-line bg-canvas px-3 text-xs font-medium text-danger transition-colors hover:bg-danger/10 disabled:opacity-50"
                        >
                          <Trash2 aria-hidden className="size-3.5" />
                          {t("agent.deleteSession")}
                        </button>
                      </div>
                    </div>
                    <div className="text-sm text-fg">
                      <SafeMarkdown content={mem.content} />
                    </div>
                  </Card>
                </li>
              ))}
            </ul>
          )}
        </>
      )}

      <ConfirmDialog
        open={Boolean(deleteTarget)}
        title={t("agent.confirmDeleteMemory")}
        description={t("agent.confirmDeleteMemoryDescription")}
        confirmLabel={t("agent.deleteSession")}
        cancelLabel={t("agent.cancel")}
        pending={deleteMemory.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          const id = deleteTarget.id;
          void deleteMemory.mutateAsync(id).then(() => {
            setDeleteTarget(null);
          });
        }}
        onCancel={() => setDeleteTarget(null)}
      />
    </div>
  );
}
