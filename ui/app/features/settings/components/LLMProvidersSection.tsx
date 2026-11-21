import { useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { useSyncProviderModels, useUpdateSettings } from "../api/actions";
import type { SettingResponse } from "../api/types";
import {
  isProviderConfigured,
  LLM_PROVIDERS,
  type LLMProviderDef,
  parseProviderModels,
} from "../model";

interface LLMProvidersSectionProps {
  itemsByKey: Record<string, SettingResponse>;
  draft: Record<string, string>;
  onChange: (key: string, value: string) => void;
  onStatus: (msg: string | null, err: string | null) => void;
}

export function LLMProvidersSection({
  itemsByKey,
  draft,
  onChange,
  onStatus,
}: LLMProvidersSectionProps) {
  const { t } = useTranslation();
  const updateSettings = useUpdateSettings();
  const syncModels = useSyncProviderModels();
  const [activeProvider, setActiveProvider] = useState<string | null>(null);

  const busy = updateSettings.isPending || syncModels.isPending;

  async function handleSaveProvider(p: LLMProviderDef) {
    onStatus(null, null);
    setActiveProvider(p.id);
    try {
      await updateSettings.mutateAsync({
        values: {
          [p.apiKeyKey]: draft[p.apiKeyKey] ?? "",
          [p.baseURLKey]: draft[p.baseURLKey] ?? "",
        },
      });
      onStatus(t("settings.saved"), null);
    } catch (err) {
      onStatus(
        null,
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
    } finally {
      setActiveProvider(null);
    }
  }

  async function handleSyncProvider(p: LLMProviderDef) {
    onStatus(null, null);
    setActiveProvider(p.id);
    try {
      const apiKeyDirty =
        (draft[p.apiKeyKey] ?? "") !== (itemsByKey[p.apiKeyKey]?.value ?? "");
      const baseURLDirty =
        (draft[p.baseURLKey] ?? "") !== (itemsByKey[p.baseURLKey]?.value ?? "");

      if (apiKeyDirty || baseURLDirty) {
        await updateSettings.mutateAsync({
          values: {
            [p.apiKeyKey]: draft[p.apiKeyKey] ?? "",
            [p.baseURLKey]: draft[p.baseURLKey] ?? "",
          },
        });
      }
      await syncModels.mutateAsync(p.id);
      onStatus(t("settings.providers.synced", { provider: p.label }), null);
    } catch (err) {
      onStatus(
        null,
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
    } finally {
      setActiveProvider(null);
    }
  }

  async function handleClearProvider(p: LLMProviderDef) {
    onStatus(null, null);
    setActiveProvider(p.id);
    try {
      await updateSettings.mutateAsync({
        values: {
          [p.apiKeyKey]: "",
          [p.baseURLKey]: "",
          [p.modelsKey]: "",
        },
      });
      onStatus(t("settings.saved"), null);
    } catch (err) {
      onStatus(
        null,
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
    } finally {
      setActiveProvider(null);
    }
  }

  return (
    <div className="flex flex-col gap-5">
      {LLM_PROVIDERS.map((p) => {
        const configured = isProviderConfigured(itemsByKey, draft, p.id);
        const models = parseProviderModels(
          draft[p.modelsKey] ?? itemsByKey[p.modelsKey]?.value,
        );
        const isThisSyncing = syncModels.isPending && activeProvider === p.id;

        return (
          <section
            key={p.id}
            className="rounded-xl border border-line bg-surface p-5 shadow-xs"
          >
            <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
              <div className="flex flex-wrap items-center gap-2.5">
                <h2 className="text-base font-semibold text-fg">{p.label}</h2>
                <span
                  className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    configured
                      ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300"
                      : "bg-fg/8 text-muted"
                  }`}
                >
                  {configured
                    ? t("settings.active")
                    : t("settings.unconfigured")}
                </span>
                <span className="inline-flex items-center rounded-full bg-fg/6 px-2.5 py-0.5 text-xs text-muted">
                  {t("settings.providers.modelCount", { count: models.length })}
                </span>
              </div>

              <button
                type="button"
                disabled={busy}
                onClick={() => void handleClearProvider(p)}
                className="rounded-lg border border-line px-3 py-1.5 text-xs font-medium text-muted transition hover:bg-fg/6 hover:text-fg disabled:cursor-not-allowed disabled:opacity-50"
              >
                {t("settings.providers.clear")}
              </button>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor={`setting-${p.apiKeyKey}`}
                  className="text-sm font-medium text-fg"
                >
                  {t("settings.providers.apiKey")}
                </label>
                <input
                  id={`setting-${p.apiKeyKey}`}
                  type="password"
                  value={draft[p.apiKeyKey] ?? ""}
                  onChange={(e) => onChange(p.apiKeyKey, e.target.value)}
                  placeholder="sk-..."
                  autoComplete="off"
                  className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor={`setting-${p.baseURLKey}`}
                  className="text-sm font-medium text-fg"
                >
                  {t("settings.providers.baseUrl")}
                </label>
                <input
                  id={`setting-${p.baseURLKey}`}
                  type="url"
                  value={draft[p.baseURLKey] ?? ""}
                  onChange={(e) => onChange(p.baseURLKey, e.target.value)}
                  placeholder={p.defaultBaseURL}
                  autoComplete="off"
                  className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
                />
              </div>
            </div>

            <div className="mt-4 flex flex-wrap items-center gap-2.5">
              <button
                type="button"
                disabled={busy}
                onClick={() => void handleSaveProvider(p)}
                className="h-9 rounded-lg bg-fg px-4 text-xs font-semibold text-bg transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {t("settings.providers.save")}
              </button>
              <button
                type="button"
                disabled={busy}
                onClick={() => void handleSyncProvider(p)}
                className="h-9 rounded-lg border border-line bg-bg px-4 text-xs font-medium text-fg transition hover:bg-fg/6 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {isThisSyncing
                  ? t("settings.providers.syncing")
                  : t("settings.providers.sync")}
              </button>
            </div>

            <details className="mt-4 rounded-lg border border-line/70 bg-bg/50 px-3.5 py-2.5">
              <summary className="cursor-pointer text-xs font-medium text-muted select-none hover:text-fg">
                {t("settings.providers.catalogToggle", {
                  count: models.length,
                })}
              </summary>
              {models.length === 0 ? (
                <p className="mt-2.5 text-xs text-muted">
                  {t("settings.providers.catalogEmpty")}
                </p>
              ) : (
                <ul className="mt-2.5 flex max-h-56 flex-col gap-1.5 overflow-y-auto pr-1">
                  {models.map((m) => (
                    <li
                      key={m.id}
                      className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-line/50 bg-surface px-2.5 py-1.5 text-xs"
                    >
                      <div className="flex flex-col">
                        <span className="font-mono font-medium text-fg">
                          {m.id}
                        </span>
                        {m.name && m.name !== m.id ? (
                          <span className="text-[11px] text-muted">
                            {m.name}
                          </span>
                        ) : null}
                      </div>
                      <div className="flex flex-wrap gap-1">
                        {(m.capabilities && m.capabilities.length > 0
                          ? m.capabilities
                          : ["chat"]
                        ).map((cap) => (
                          <span
                            key={cap}
                            className="rounded-full bg-fg/8 px-2 py-0.5 text-[10px] font-medium text-muted"
                          >
                            {cap}
                          </span>
                        ))}
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </details>
          </section>
        );
      })}
    </div>
  );
}
