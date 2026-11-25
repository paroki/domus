import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { ApiError } from "~/shared/api/client";
import { Button } from "~/shared/ui/Button";
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
    } catch (err) {
      editor.setStatus(
        null,
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
    }
  }

  async function handleSaveModels(e: FormEvent) {
    e.preventDefault();
    await saveKeys(AGENT_MODEL_KEYS);
  }

  async function handleSavePrompts(e: FormEvent) {
    e.preventDefault();
    await saveKeys(AGENT_PROMPT_KEYS);
  }

  async function handleResetPrompt(key: string, label: string) {
    editor.setStatus(null, null);
    try {
      await updateSettings.mutateAsync({ values: { [key]: "" } });
      editor.setStatus(
        t("settings.agents.resetDone", { name: label || key }),
        null,
      );
    } catch (err) {
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

  return (
    <SettingsSubpageShell
      backTo="/settings/system"
      backLabel={t("settings.backToSettings")}
      title={t("settings.agents.title")}
      subtitle={t("settings.agents.subtitle")}
      meLoading={editor.meLoading}
      canManage={editor.canManage}
      settingsLoading={editor.settingsLoading}
      isEmpty={editor.items.length === 0}
      errorMsg={editor.errorMsg}
      savedMsg={editor.savedMsg}
    >
      <form
        onSubmit={handleSaveModels}
        className="rounded-xl border border-line bg-surface p-5 shadow-xs"
      >
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-base font-semibold text-fg">
              {t("settings.agents.modelRoutingTitle")}
            </h2>
            <p className="mt-1 text-xs text-muted">
              {t("settings.agents.embeddingNote")}
            </p>
          </div>
          {providers.length === 0 ? (
            <Link
              to="/settings/system/ai-provider"
              className="text-xs font-medium text-accent hover:underline"
            >
              {t("settings.agents.goToProviders")}
            </Link>
          ) : null}
        </div>

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
                      className="text-xs text-accent hover:underline"
                    >
                      {isCustom
                        ? t("settings.agents.selectModel")
                        : t("settings.agents.customModel")}
                    </button>
                  ) : null}
                </div>

                {isCustom ? (
                  <div className="grid grid-cols-2 gap-2">
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
                      className="rounded-lg border border-line bg-canvas px-3 py-2 text-sm text-fg focus:border-accent focus:outline-none"
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
                        const pid = parsed.providerId || providers[0]?.id || "";
                        editor.updateDraft(
                          key,
                          pid && mid ? `${pid}:${mid}` : "",
                        );
                      }}
                      className="rounded-lg border border-line bg-canvas px-3 py-2 text-sm text-fg placeholder:text-muted/70 focus:border-accent focus:outline-none"
                    />
                  </div>
                ) : (
                  <select
                    id={`setting-${key}`}
                    value={currentValue}
                    disabled={updateSettings.isPending}
                    onChange={(e) => editor.updateDraft(key, e.target.value)}
                    className="rounded-lg border border-line bg-canvas px-3 py-2 text-sm text-fg focus:border-accent focus:outline-none"
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

        <div className="mt-5 flex justify-end">
          <Button
            type="submit"
            disabled={updateSettings.isPending}
            className="h-10"
          >
            {updateSettings.isPending
              ? t("settings.saving")
              : t("settings.agents.save")}
          </Button>
        </div>
      </form>

      <form
        onSubmit={handleSavePrompts}
        className="rounded-xl border border-line bg-surface p-5 shadow-xs"
      >
        <h2 className="mb-4 text-base font-semibold text-fg">
          {t("settings.agents.promptsTitle")}
        </h2>

        <div className="flex flex-col gap-5">
          {AGENT_PROMPT_KEYS.map((key) => {
            const item = editor.itemsByKey[key];
            if (!item) return null;
            const value = editor.draft[key] ?? item.value ?? "";
            return (
              <div key={key} className="flex flex-col gap-1.5">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <label
                      htmlFor={`setting-${key}`}
                      className="text-sm font-medium text-fg"
                    >
                      {item.label}
                    </label>
                    {item.overridden ? (
                      <span className="inline-flex items-center rounded-full bg-warning-bg px-2 py-0.5 text-xs font-medium text-warning-fg">
                        {t("settings.agents.overridden")}
                      </span>
                    ) : null}
                  </div>
                  {item.overridden ? (
                    <button
                      type="button"
                      disabled={updateSettings.isPending}
                      onClick={() => handleResetPrompt(key, item.label ?? key)}
                      className="rounded-lg border border-line px-2.5 py-1 text-xs font-medium text-muted transition hover:bg-fg/6 hover:text-fg disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      {t("settings.agents.resetInstruction")}
                    </button>
                  ) : null}
                </div>

                <textarea
                  id={`setting-${key}`}
                  rows={8}
                  value={value}
                  disabled={updateSettings.isPending}
                  onChange={(e) => editor.updateDraft(key, e.target.value)}
                  className="w-full rounded-lg border border-line bg-canvas px-3 py-2 font-mono text-xs text-fg focus:border-accent focus:outline-none"
                />

                {item.description ? (
                  <p className="text-xs text-muted">{item.description}</p>
                ) : null}
              </div>
            );
          })}
        </div>

        <div className="mt-5 flex justify-end">
          <Button
            type="submit"
            disabled={updateSettings.isPending}
            className="h-10"
          >
            {updateSettings.isPending
              ? t("settings.saving")
              : t("settings.agents.savePrompts")}
          </Button>
        </div>
      </form>
    </SettingsSubpageShell>
  );
}
