import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { useUpdateSettings } from "../api/actions";
import type { SettingResponse } from "../api/types";
import {
  AGENT_PERSONAS,
  type AgentPersonaDef,
  buildProviderOptionList,
  EMBEDDING_KEYS,
  filterModelsByCapability,
  parseProviderModels,
  REASONING_EFFORTS,
} from "../model";

interface AgentConfigSectionProps {
  itemsByKey: Record<string, SettingResponse>;
  draft: Record<string, string>;
  onChange: (key: string, value: string) => void;
  onStatus: (msg: string | null, err: string | null) => void;
}

function getProviderChatModels(
  itemsByKey: Record<string, SettingResponse>,
  draft: Record<string, string>,
  provider: string,
) {
  const modelsKey = `agent.provider.${provider}.models`;
  const raw = draft[modelsKey] ?? itemsByKey[modelsKey]?.value;
  return filterModelsByCapability(parseProviderModels(raw), "chat");
}

function getProviderEmbeddingModels(
  itemsByKey: Record<string, SettingResponse>,
  draft: Record<string, string>,
  provider: string,
) {
  const modelsKey = `agent.provider.${provider}.models`;
  const raw = draft[modelsKey] ?? itemsByKey[modelsKey]?.value;
  return filterModelsByCapability(parseProviderModels(raw), "embedding");
}

export function AgentConfigSection({
  itemsByKey,
  draft,
  onChange,
  onStatus,
}: AgentConfigSectionProps) {
  const { t } = useTranslation();
  const updateSettings = useUpdateSettings();
  const [customModelMode, setCustomModelMode] = useState<
    Record<string, boolean>
  >({});

  const providerOptions = buildProviderOptionList(itemsByKey, draft, false);
  const embeddingProviderOptions = buildProviderOptionList(
    itemsByKey,
    draft,
    true,
  );

  function handleAgentProviderChange(
    persona: AgentPersonaDef,
    nextProvider: string,
  ) {
    onChange(persona.providerKey, nextProvider);
    const nextChatModels = getProviderChatModels(
      itemsByKey,
      draft,
      nextProvider,
    );
    const currentModel = draft[persona.modelKey] ?? "";
    if (
      nextChatModels.length > 0 &&
      !nextChatModels.some((m) => m.id === currentModel)
    ) {
      const firstId = nextChatModels[0]?.id;
      if (firstId) {
        onChange(persona.modelKey, firstId);
      }
    }
  }

  async function handleResetInstruction(persona: AgentPersonaDef) {
    onStatus(null, null);
    try {
      await updateSettings.mutateAsync({
        values: { [persona.instructionKey]: "" },
      });
      onStatus(t("settings.agents.resetDone", { name: persona.name }), null);
    } catch (err) {
      onStatus(
        null,
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
    }
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    onStatus(null, null);
    const values: Record<string, string> = {};
    for (const p of AGENT_PERSONAS) {
      values[p.providerKey] = draft[p.providerKey] ?? "";
      values[p.modelKey] = draft[p.modelKey] ?? "";
      values[p.effortKey] = draft[p.effortKey] ?? "";
      values[p.instructionKey] = draft[p.instructionKey] ?? "";
    }
    values[EMBEDDING_KEYS.providerKey] =
      draft[EMBEDDING_KEYS.providerKey] ?? "";
    values[EMBEDDING_KEYS.modelKey] = draft[EMBEDDING_KEYS.modelKey] ?? "";

    try {
      await updateSettings.mutateAsync({ values });
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

  const embProvider = draft[EMBEDDING_KEYS.providerKey] ?? "gemini";
  const embModel = draft[EMBEDDING_KEYS.modelKey] ?? "";
  const embModels =
    embProvider === "none"
      ? []
      : getProviderEmbeddingModels(itemsByKey, draft, embProvider);
  const embHasCustom =
    embModel.trim() !== "" && !embModels.some((m) => m.id === embModel);

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5">
      {AGENT_PERSONAS.map((persona) => {
        const prov = draft[persona.providerKey] ?? "gemini";
        const modelVal = draft[persona.modelKey] ?? "";
        const effortVal = draft[persona.effortKey] ?? "low";
        const instrVal = draft[persona.instructionKey] ?? "";
        const chatModels = getProviderChatModels(itemsByKey, draft, prov);
        const useCustom =
          Boolean(customModelMode[persona.id]) || chatModels.length === 0;
        const modelAbsentFromCatalog =
          modelVal.trim() !== "" && !chatModels.some((m) => m.id === modelVal);

        return (
          <section
            key={persona.id}
            className="rounded-xl border border-line bg-surface p-5 shadow-xs"
          >
            <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
              <div>
                <h2 className="text-base font-semibold text-fg">
                  {persona.name}
                </h2>
                <p className="text-xs text-muted">{persona.role}</p>
              </div>
              <div className="flex flex-wrap gap-1.5">
                {persona.tools.map((tool) => (
                  <span
                    key={tool}
                    className="rounded-full bg-fg/8 px-2.5 py-0.5 font-mono text-[11px] text-muted"
                  >
                    {tool}
                  </span>
                ))}
              </div>
            </div>

            <div className="grid gap-4 sm:grid-cols-3">
              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor={`setting-${persona.providerKey}`}
                  className="text-sm font-medium text-fg"
                >
                  {t("settings.agents.provider")}
                </label>
                <select
                  id={`setting-${persona.providerKey}`}
                  value={prov}
                  onChange={(e) =>
                    handleAgentProviderChange(persona, e.target.value)
                  }
                  className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg focus:border-fg/40 focus:outline-none"
                >
                  {providerOptions.map((opt) => (
                    <option key={opt.id} value={opt.id}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex flex-col gap-1.5">
                <div className="flex items-center justify-between gap-2">
                  <label
                    htmlFor={`setting-${persona.modelKey}`}
                    className="text-sm font-medium text-fg"
                  >
                    {t("settings.agents.model")}
                  </label>
                  {chatModels.length > 0 ? (
                    <button
                      type="button"
                      onClick={() =>
                        setCustomModelMode((prev) => ({
                          ...prev,
                          [persona.id]: !prev[persona.id],
                        }))
                      }
                      className="text-[11px] font-medium text-muted underline-offset-2 hover:text-fg hover:underline"
                    >
                      {useCustom
                        ? t("settings.agents.selectModel")
                        : t("settings.agents.customModel")}
                    </button>
                  ) : null}
                </div>
                {useCustom ? (
                  <input
                    id={`setting-${persona.modelKey}`}
                    type="text"
                    value={modelVal}
                    onChange={(e) => onChange(persona.modelKey, e.target.value)}
                    placeholder="gemini-2.5-flash"
                    className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
                  />
                ) : (
                  <select
                    id={`setting-${persona.modelKey}`}
                    value={modelVal}
                    onChange={(e) => onChange(persona.modelKey, e.target.value)}
                    className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg focus:border-fg/40 focus:outline-none"
                  >
                    {modelAbsentFromCatalog ? (
                      <option value={modelVal}>{modelVal}</option>
                    ) : null}
                    {chatModels.map((m) => (
                      <option key={m.id} value={m.id}>
                        {m.name && m.name !== m.id
                          ? `${m.name} (${m.id})`
                          : m.id}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor={`setting-${persona.effortKey}`}
                  className="text-sm font-medium text-fg"
                >
                  {t("settings.agents.effort")}
                </label>
                <select
                  id={`setting-${persona.effortKey}`}
                  value={effortVal}
                  onChange={(e) => onChange(persona.effortKey, e.target.value)}
                  className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg focus:border-fg/40 focus:outline-none"
                >
                  {REASONING_EFFORTS.map((eff) => (
                    <option key={eff.value} value={eff.value}>
                      {eff.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div className="mt-4 flex flex-col gap-1.5">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <label
                  htmlFor={`setting-${persona.instructionKey}`}
                  className="text-sm font-medium text-fg"
                >
                  {t("settings.agents.instruction")}
                </label>
                <button
                  type="button"
                  disabled={updateSettings.isPending}
                  onClick={() => void handleResetInstruction(persona)}
                  className="rounded-lg border border-line px-2.5 py-1 text-xs font-medium text-muted transition hover:bg-fg/6 hover:text-fg disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {t("settings.agents.resetInstruction")}
                </button>
              </div>
              <textarea
                id={`setting-${persona.instructionKey}`}
                rows={10}
                value={instrVal}
                onChange={(e) =>
                  onChange(persona.instructionKey, e.target.value)
                }
                className="rounded-lg border border-line bg-bg p-3 font-mono text-xs/relaxed text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
              />
            </div>
          </section>
        );
      })}

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
              {embeddingProviderOptions.map((opt) => (
                <option key={opt.id} value={opt.id}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>

          <div className="flex flex-col gap-1.5">
            <label
              htmlFor={`setting-${EMBEDDING_KEYS.modelKey}`}
              className="text-sm font-medium text-fg"
            >
              {t("settings.agents.model")}
            </label>
            {embModels.length > 0 && embProvider !== "none" ? (
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
            : t("settings.agents.save")}
        </button>
      </div>
    </form>
  );
}
