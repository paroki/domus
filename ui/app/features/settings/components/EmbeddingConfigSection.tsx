import type { FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { ApiError } from "~/shared/api/client";
import { useUpdateSettings } from "../api/actions";
import { useAgentProviders } from "../api/queries";
import type { SettingResponse } from "../api/types";
import {
  buildProviderOptionList,
  EMBEDDING_KEYS,
  getProviderModelsByCapability,
} from "../model";

interface EmbeddingConfigSectionProps {
  itemsByKey: Record<string, SettingResponse>;
  draft: Record<string, string>;
  onChange: (key: string, value: string) => void;
  onStatus: (msg: string | null, err: string | null) => void;
}

export function EmbeddingConfigSection({
  draft,
  onChange,
  onStatus,
}: EmbeddingConfigSectionProps) {
  const { t } = useTranslation();
  const updateSettings = useUpdateSettings();
  const { data: providers = [] } = useAgentProviders();

  const embeddingProviderOptions = buildProviderOptionList(providers, true);
  const embProvider = draft[EMBEDDING_KEYS.providerKey] ?? "";
  const embModel = draft[EMBEDDING_KEYS.modelKey] ?? "";
  const embModels =
    embProvider === "" || embProvider === "none"
      ? []
      : getProviderModelsByCapability(providers, embProvider, "embedding");
  const embHasCustom =
    embModel.trim() !== "" && !embModels.some((m) => m.id === embModel);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    onStatus(null, null);
    try {
      await updateSettings.mutateAsync({
        values: {
          [EMBEDDING_KEYS.providerKey]: draft[EMBEDDING_KEYS.providerKey] ?? "",
          [EMBEDDING_KEYS.modelKey]: draft[EMBEDDING_KEYS.modelKey] ?? "",
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
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5">
      <section className="rounded-xl border border-line bg-surface p-5 shadow-xs">
        <h2 className="mb-4 text-base font-semibold text-fg">
          {t("settings.agents.embeddingTitle")}
        </h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor={`setting-${EMBEDDING_KEYS.providerKey}`}
              className="text-sm font-medium text-fg"
            >
              {t("settings.agents.provider")}
            </label>
            <select
              id={`setting-${EMBEDDING_KEYS.providerKey}`}
              value={embProvider}
              onChange={(e) =>
                onChange(EMBEDDING_KEYS.providerKey, e.target.value)
              }
              className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg focus:border-fg/40 focus:outline-none"
            >
              <option value="">{t("settings.agents.selectProvider")}</option>
              {embeddingProviderOptions.map((opt) => (
                <option key={opt.id} value={opt.id}>
                  {opt.label}
                </option>
              ))}
            </select>
            {providers.length === 0 ? (
              <p className="text-[11px] text-muted">
                {t("settings.agents.noProvidersHint")}{" "}
                <Link
                  to="/settings/ai-provider"
                  className="font-medium text-fg underline underline-offset-2"
                >
                  {t("settings.agents.goToProviders")}
                </Link>
              </p>
            ) : null}
          </div>

          <div className="flex flex-col gap-1.5">
            <label
              htmlFor={`setting-${EMBEDDING_KEYS.modelKey}`}
              className="text-sm font-medium text-fg"
            >
              {t("settings.agents.model")}
            </label>
            {embModels.length > 0 &&
            embProvider !== "none" &&
            embProvider !== "" ? (
              <select
                id={`setting-${EMBEDDING_KEYS.modelKey}`}
                value={embModel}
                onChange={(e) =>
                  onChange(EMBEDDING_KEYS.modelKey, e.target.value)
                }
                className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg focus:border-fg/40 focus:outline-none"
              >
                {embHasCustom ? (
                  <option value={embModel}>{embModel}</option>
                ) : null}
                {embModels.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name && m.name !== m.id ? `${m.name} (${m.id})` : m.id}
                  </option>
                ))}
              </select>
            ) : (
              <input
                id={`setting-${EMBEDDING_KEYS.modelKey}`}
                type="text"
                disabled={embProvider === "none"}
                value={embModel}
                onChange={(e) =>
                  onChange(EMBEDDING_KEYS.modelKey, e.target.value)
                }
                placeholder="text-embedding-004"
                className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
              />
            )}
          </div>
        </div>
        <p className="mt-3 text-xs text-muted">
          {t("settings.agents.embeddingNote")}
        </p>
      </section>

      <div className="flex justify-end">
        <button
          type="submit"
          disabled={updateSettings.isPending}
          className="h-10 rounded-lg bg-fg px-5 text-sm font-semibold text-bg transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {updateSettings.isPending
            ? t("settings.saving")
            : t("settings.agents.saveEmbedding")}
        </button>
      </div>
    </form>
  );
}
