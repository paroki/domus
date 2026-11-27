import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button, buttonClass } from "~/shared/ui/Button";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useCreateAgentMemory, useUpdateAgentMemory } from "../api/actions";
import { useAgentMemories } from "../api/queries";
import {
  MAX_MEMORY_CHARS,
  MEMORY_CATEGORIES,
  type MemoryCategory,
  validateMemoryContent,
} from "../model";

export function meta() {
  return [{ title: "Formulir Memori AI — Domus" }];
}

export default function AgentMemoryFormPage() {
  const { memoryId } = useParams<{ memoryId: string }>();
  const isEdit = Boolean(memoryId);
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: me } = useMe();
  const hasParish = Boolean(me?.activeParokiId);

  const { data: memories = [], isLoading } = useAgentMemories(
    "",
    "",
    isEdit && hasParish,
  );
  const targetMemory = isEdit
    ? memories.find((m) => m.id === memoryId)
    : undefined;

  const createMemory = useCreateAgentMemory();
  const updateMemory = useUpdateAgentMemory();

  const [category, setCategory] = useState<MemoryCategory>("fact");
  const [content, setContent] = useState("");
  const [formErr, setFormErr] = useState<string | null>(null);

  useEffect(() => {
    if (!targetMemory) return;
    setCategory(targetMemory.category);
    setContent(targetMemory.content);
  }, [targetMemory]);

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
    const err = validateMemoryContent(content);
    if (err) {
      setFormErr(err);
      return;
    }
    setFormErr(null);
    const trimmed = content.trim();
    try {
      if (isEdit && memoryId) {
        await updateMemory.mutateAsync({
          memoryId,
          body: { content: trimmed, category },
        });
      } else {
        await createMemory.mutateAsync({
          content: trimmed,
          category,
        });
      }
      navigate("/agent/memories");
    } catch (submitErr) {
      setFormErr(
        submitErr instanceof Error
          ? submitErr.message
          : t("agent.memoryNotFound"),
      );
    }
  }

  const pending = createMemory.isPending || updateMemory.isPending;
  const pageTitle = isEdit ? t("agent.editMemory") : t("agent.memoryAdd");

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: t("agent.title"), to: "/agent" },
          { label: t("agent.memoriesBreadcrumb"), to: "/agent/memories" },
          { label: pageTitle },
        ]}
        title={pageTitle}
        subtitle={
          isEdit
            ? t("agent.memoryFormSubtitleEdit")
            : t("agent.memoryFormSubtitleAdd")
        }
      />

      {isEdit && isLoading ? (
        <p className="text-sm text-muted">{t("common.loading")}</p>
      ) : isEdit && !targetMemory ? (
        <div
          role="alert"
          className="flex flex-col items-start gap-3 rounded-xl border border-danger/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
        >
          <span>{t("agent.memoryNotFound")}</span>
          <Link to="/agent/memories" className={buttonClass("outline")}>
            {t("agent.backToMemories")}
          </Link>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="flex flex-col gap-6">
          <InductiveFormSection title={t("agent.memoryFormSectionTitle")}>
            <div className="flex flex-col gap-1.5">
              <label
                htmlFor="memory-category"
                className="text-xs font-semibold text-muted"
              >
                {t("agent.memoryCategoryLabel")}
              </label>
              <select
                id="memory-category"
                value={category}
                onChange={(e) => setCategory(e.target.value as MemoryCategory)}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
              >
                {MEMORY_CATEGORIES.map((c) => (
                  <option key={c} value={c}>
                    {categoryLabel(c)}
                  </option>
                ))}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <label
                htmlFor="memory-content"
                className="text-xs font-semibold text-muted"
              >
                {t("agent.memoryContentLabel")}
              </label>
              <textarea
                id="memory-content"
                rows={5}
                value={content}
                maxLength={MAX_MEMORY_CHARS}
                onChange={(e) => {
                  setContent(e.target.value);
                  if (formErr) setFormErr(null);
                }}
                placeholder={t("agent.memoryContentPlaceholder")}
                className="w-full rounded-lg border border-line bg-canvas px-3 py-2.5 text-base text-fg placeholder:text-muted/70 focus:border-primary focus:outline-none sm:text-sm"
              />
              <div className="flex items-center justify-between gap-2 text-xs">
                {formErr ? (
                  <span role="alert" className="text-danger">
                    {formErr}
                  </span>
                ) : (
                  <span />
                )}
                <span className="tabular-nums text-muted">
                  {content.length} / {MAX_MEMORY_CHARS}
                </span>
              </div>
            </div>

            <div className="flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:justify-end">
              <Link
                to="/agent/memories"
                className={`${buttonClass("outline")} min-h-11 w-full sm:w-auto`}
              >
                {t("agent.cancel")}
              </Link>
              <Button
                type="submit"
                disabled={pending}
                className="h-11 w-full sm:w-auto"
              >
                {isEdit ? t("agent.saveMemory") : t("agent.memoryAdd")}
              </Button>
            </div>
          </InductiveFormSection>
        </form>
      )}
    </div>
  );
}
