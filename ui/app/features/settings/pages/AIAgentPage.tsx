import { Cpu } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import { InductiveStepCard } from "~/shared/ui/iui";
import { useUpdateSettings } from "../api/actions";
import { useAgentProviders } from "../api/queries";
import {
  SettingsSubpageShell,
  useSettingsEditor,
} from "../components/SettingsSubpageShell";
import {
  AGENT_MODEL_KEYS,
  AGENT_PROMPT_KEYS,
  buildProviderModelOptions,
  parseModelSelection,
} from "../model";

export function meta() {
  return [{ title: "AI Agent — Domus" }];
}

export default function AIAgentPage() {
  const { t } = useTranslation();
  const editor = useSettingsEditor();
  const { data: providers = [] } = useAgentProviders(editor.canManage);
  const updateSettings = useUpdateSettings();
  const [editingModels, setEditingModels] = useState(false);
  const [editingPromptKey, setEditingPromptKey] = useState<string | null>(null);
  const [resetTarget, setResetTarget] = useState<{
    key: string;
    label: string;
  } | null>(null);
  const [customMode, setCustomMode] = useState<Record<string, boolean>>({});

  async function saveKeys(keys: readonly string[]) {
    editor.setStatus(null, null);
    const values: Record<string, string> = {};
    for (const key of keys) {
      values[key] = editor.draft[key] ?? "";
    }
    try {
      await updateSettings.mutateAsync({ values });
      editor.setStatus(t("settings.saved"), null);
      return true;
    } catch (err) {
      editor.setStatus(
        null,
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
      return false;
    }
  }

  async function handleSaveModels(e: FormEvent) {
    e.preventDefault();
    const ok = await saveKeys(AGENT_MODEL_KEYS);
    if (ok) {
      setEditingModels(false);
    }
  }

  async function handleSavePrompt(e: FormEvent, key: string) {
    e.preventDefault();
    const ok = await saveKeys([key]);
    if (ok) {
      setEditingPromptKey(null);
    }
  }

  async function handleResetPrompt(key: string, label: string) {
    editor.setStatus(null, null);
    try {
      await updateSettings.mutateAsync({ values: { [key]: "" } });
      setResetTarget(null);
      if (editingPromptKey === key) {
        setEditingPromptKey(null);
      }
      editor.setStatus(
        t("settings.agents.resetDone", { name: label || key }),
        null,
      );
    } catch (err) {
      setResetTarget(null);
      editor.setStatus(
        null,
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
    }
  }

  const chatOptions = buildProviderModelOptions(providers, "chat");
  const embeddingOptions = buildProviderModelOptions(providers, "embedding");

  function resolveModelDisplay(key: string, value: string): string {
    if (!value) {
      if (key === "agent.model.gabriel") {
        return "Belum dipilih";
      }
      if (key === "agent.model.embedding") {
        return t("settings.agents.embeddingDisabled");
      }
      return t("settings.agents.followGabriel");
    }
    const match = [...chatOptions, ...embeddingOptions].find(
      (opt) => opt.value === value,
    );
    if (match) return match.label;
    const parsed = parseModelSelection(value);
    const provider = providers.find((p) => p.id === parsed.providerId);
    if (provider && parsed.modelId) {
      return `${provider.name ?? provider.provider} — ${parsed.modelId}`;
    }
    return value;
  }

  return (
    <SettingsSubpageShell
      title={t("settings.agents.title")}
      subtitle={t("settings.agents.subtitle")}
      meLoading={editor.meLoading}
      canManage={editor.canManage}
      settingsLoading={editor.settingsLoading}
      isEmpty={editor.items.length === 0}
      errorMsg={editor.errorMsg}
      savedMsg={editor.savedMsg}
    >
      {providers.length === 0 ? (
        <InductiveStepCard
          to="/settings/system/ai-provider"
          title={t("settings.agents.goToProviders")}
          description={t("settings.agents.noProvidersHint")}
          icon={Cpu}
          plate={3}
        />
      ) : null}

      <section className="rounded-xl border border-line bg-surface p-4 sm:p-6">
        <div className="mb-4 flex flex-col justify-between gap-3 sm:flex-row sm:items-start">
          <div>
            <h2 className="text-base font-semibold text-fg">
              {t("settings.agents.modelRoutingTitle")}
            </h2>
            <p className="mt-1 text-xs text-muted">
              {t("settings.agents.embeddingNote")}
            </p>
          </div>
          <div className="flex w-full shrink-0 flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
            {providers.length > 0 ? (
              <Link
                to="/settings/system/ai-provider"
                className="text-xs font-medium text-primary hover:underline sm:mr-2"
              >
                {t("settings.agents.goToProviders")}
              </Link>
            ) : null}
            {!editingModels ? (
              <Button
                type="button"
                variant="outline"
                disabled={updateSettings.isPending}
                onClick={() => {
                  editor.setStatus(null, null);
                  setEditingModels(true);
                }}
                className="w-full sm:w-auto"
              >
                Ubah
              </Button>
            ) : null}
          </div>
        </div>

        {editingModels ? (
          <form onSubmit={handleSaveModels} className="flex flex-col gap-5">
            <div className="grid gap-4 sm:grid-cols-2">
              {AGENT_MODEL_KEYS.map((key) => {
                const item = editor.itemsByKey[key];
                if (!item) return null;
                const isEmbedding = key === "agent.model.embedding";
                const options = isEmbedding ? embeddingOptions : chatOptions;
                const currentValue = editor.draft[key] ?? "";
                const parsed = parseModelSelection(currentValue);
                const inCatalog =
                  currentValue === "" ||
                  options.some((opt) => opt.value === currentValue);
                const isCustom = customMode[key] ?? !inCatalog;
                const emptyLabel =
                  key === "agent.model.gabriel"
                    ? t("settings.agents.selectGabrielModel")
                    : isEmbedding
                      ? t("settings.agents.embeddingDisabled")
                      : t("settings.agents.followGabriel");

                return (
                  <div key={key} className="flex flex-col gap-1.5">
                    <div className="flex items-center justify-between gap-2">
                      <label
                        htmlFor={`setting-${key}`}
                        className="text-sm font-medium text-fg"
                      >
                        {item.label}
                      </label>
                      {providers.length > 0 ? (
                        <button
                          type="button"
                          onClick={() =>
                            setCustomMode((prev) => ({
                              ...prev,
                              [key]: !isCustom,
                            }))
                          }
                          className="text-xs font-medium text-primary hover:underline"
                        >
                          {isCustom
                            ? t("settings.agents.selectModel")
                            : t("settings.agents.customModel")}
                        </button>
                      ) : null}
                    </div>

                    {isCustom ? (
                      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                        <select
                          aria-label={`${item.label} ${t("settings.agents.provider")}`}
                          value={parsed.providerId}
                          disabled={updateSettings.isPending}
                          onChange={(e) => {
                            const pid = e.target.value;
                            editor.updateDraft(
                              key,
                              pid ? `${pid}:${parsed.modelId}` : "",
                            );
                          }}
                          className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-fg/40 focus:outline-none sm:text-sm"
                        >
                          <option value="">
                            {t("settings.agents.selectProvider")}
                          </option>
                          {providers.map((p) => (
                            <option key={p.id} value={p.id}>
                              {p.name ?? p.provider}
                            </option>
                          ))}
                        </select>
                        <input
                          id={`setting-${key}`}
                          type="text"
                          value={parsed.modelId}
                          placeholder={t("settings.agents.modelIdPlaceholder")}
                          disabled={updateSettings.isPending}
                          onChange={(e) => {
                            const mid = e.target.value;
                            const pid =
                              parsed.providerId || providers[0]?.id || "";
                            editor.updateDraft(
                              key,
                              pid && mid ? `${pid}:${mid}` : "",
                            );
                          }}
                          className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg placeholder:text-muted/70 focus:border-fg/40 focus:outline-none sm:text-sm"
                        />
                      </div>
                    ) : (
                      <select
                        id={`setting-${key}`}
                        value={currentValue}
                        disabled={updateSettings.isPending}
                        onChange={(e) =>
                          editor.updateDraft(key, e.target.value)
                        }
                        className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-fg/40 focus:outline-none sm:text-sm"
                      >
                        <option value="">{emptyLabel}</option>
                        {options.map((opt) => (
                          <option key={opt.value} value={opt.value}>
                            {opt.label}
                          </option>
                        ))}
                      </select>
                    )}

                    {item.description ? (
                      <p className="text-xs text-muted">{item.description}</p>
                    ) : null}
                  </div>
                );
              })}
            </div>

            <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <Button
                type="button"
                variant="outline"
                disabled={updateSettings.isPending}
                onClick={() => {
                  setEditingModels(false);
                  for (const key of AGENT_MODEL_KEYS) {
                    const item = editor.itemsByKey[key];
                    if (item) {
                      editor.updateDraft(key, item.value ?? "");
                    }
                  }
                }}
                className="w-full sm:w-auto"
              >
                {t("settings.providers.cancel")}
              </Button>
              <Button
                type="submit"
                disabled={updateSettings.isPending}
                className="w-full sm:w-auto"
              >
                {updateSettings.isPending
                  ? t("settings.saving")
                  : t("settings.agents.save")}
              </Button>
            </div>
          </form>
        ) : (
          <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {AGENT_MODEL_KEYS.map((key) => {
              const item = editor.itemsByKey[key];
              if (!item) return null;
              const currentValue = item.value ?? "";
              return (
                <div key={key} className="flex flex-col gap-1">
                  <dt className="text-xs font-medium text-muted">
                    {item.label}
                  </dt>
                  <dd className="text-sm font-medium text-fg">
                    {resolveModelDisplay(key, currentValue)}
                  </dd>
                </div>
              );
            })}
          </dl>
        )}
      </section>

      <section className="rounded-xl border border-line bg-surface p-4 sm:p-6">
        <h2 className="mb-4 text-base font-semibold text-fg">
          {t("settings.agents.promptsTitle")}
        </h2>

        <div className="flex flex-col gap-4">
          {AGENT_PROMPT_KEYS.map((key) => {
            const item = editor.itemsByKey[key];
            if (!item) return null;
            const value = editor.draft[key] ?? item.value ?? "";
            const isEditingThis = editingPromptKey === key;

            return (
              <div
                key={key}
                data-agent-prompt={key}
                className="flex flex-col gap-3 rounded-xl border border-line bg-canvas/50 p-4"
              >
                <div className="flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-semibold text-fg">
                      {item.label}
                    </span>
                    <Badge tone={item.overridden ? "warning" : "neutral"}>
                      {item.overridden
                        ? t("settings.agents.overridden")
                        : "Default"}
                    </Badge>
                  </div>

                  {!isEditingThis ? (
                    <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row">
                      {item.overridden ? (
                        <Button
                          type="button"
                          variant="outline"
                          disabled={updateSettings.isPending}
                          onClick={() =>
                            setResetTarget({
                              key,
                              label: item.label ?? key,
                            })
                          }
                          className="w-full sm:w-auto"
                        >
                          {t("settings.agents.resetInstruction")}
                        </Button>
                      ) : null}
                      <Button
                        type="button"
                        variant="outline"
                        disabled={updateSettings.isPending}
                        onClick={() => {
                          editor.setStatus(null, null);
                          setEditingPromptKey(key);
                        }}
                        className="w-full sm:w-auto"
                      >
                        Ubah Prompt
                      </Button>
                    </div>
                  ) : null}
                </div>

                {isEditingThis ? (
                  <form
                    onSubmit={(e) => void handleSavePrompt(e, key)}
                    className="flex flex-col gap-3"
                  >
                    <textarea
                      id={`setting-${key}`}
                      rows={8}
                      value={value}
                      disabled={updateSettings.isPending}
                      onChange={(e) => editor.updateDraft(key, e.target.value)}
                      className="w-full rounded-lg border border-line bg-canvas px-3 py-2 font-mono text-base text-fg focus:border-fg/40 focus:outline-none sm:text-xs"
                    />
                    {item.description ? (
                      <p className="text-xs text-muted">{item.description}</p>
                    ) : null}
                    <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
                      <Button
                        type="button"
                        variant="outline"
                        disabled={updateSettings.isPending}
                        onClick={() => {
                          setEditingPromptKey(null);
                          editor.updateDraft(key, item.value ?? "");
                        }}
                        className="w-full sm:w-auto"
                      >
                        {t("settings.providers.cancel")}
                      </Button>
                      <Button
                        type="submit"
                        disabled={updateSettings.isPending}
                        className="w-full sm:w-auto"
                      >
                        {updateSettings.isPending
                          ? t("settings.saving")
                          : t("settings.agents.savePrompts")}
                      </Button>
                    </div>
                  </form>
                ) : (
                  <p className="line-clamp-3 font-mono text-xs text-muted">
                    {item.value || "Belum ada instruksi."}
                  </p>
                )}
              </div>
            );
          })}
        </div>
      </section>

      <ConfirmDialog
        open={Boolean(resetTarget)}
        title="Reset Prompt ke Default?"
        description={`Instruksi kustom untuk ${resetTarget?.label ?? ""} akan dikembalikan ke bawaan sistem.`}
        confirmLabel={t("settings.agents.resetInstruction")}
        pending={updateSettings.isPending}
        onConfirm={() => {
          if (resetTarget) {
            void handleResetPrompt(resetTarget.key, resetTarget.label);
          }
        }}
        onCancel={() => setResetTarget(null)}
      />
    </SettingsSubpageShell>
  );
}
