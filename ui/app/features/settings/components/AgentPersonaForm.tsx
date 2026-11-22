import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { ApiError } from "~/shared/api/client";
import { Button } from "~/shared/ui/Button";
import { useUpdateSettings } from "../api/actions";
import { useAgentProviders } from "../api/queries";
import type { SettingResponse } from "../api/types";
import {
  type AgentPersonaDef,
  buildProviderOptionList,
  getProviderModelsByCapability,
  REASONING_EFFORTS,
} from "../model";

interface AgentPersonaFormProps {
  persona: AgentPersonaDef;
  itemsByKey: Record<string, SettingResponse>;
  draft: Record<string, string>;
  onChange: (key: string, value: string) => void;
  onStatus: (msg: string | null, err: string | null) => void;
}

export function AgentPersonaForm({
  persona,
  draft,
  onChange,
  onStatus,
}: AgentPersonaFormProps) {
  const { t } = useTranslation();
  const updateSettings = useUpdateSettings();
  const { data: providers = [] } = useAgentProviders();
  const [customModelMode, setCustomModelMode] = useState(false);

  const providerOptions = buildProviderOptionList(providers, false);
  const prov = draft[persona.providerKey] ?? "";
  const modelVal = draft[persona.modelKey] ?? "";
  const effortVal = draft[persona.effortKey] ?? "low";
  const instrVal = draft[persona.instructionKey] ?? "";
  const chatModels = getProviderModelsByCapability(providers, prov, "chat");
  const useCustom = customModelMode || chatModels.length === 0;
  const modelAbsentFromCatalog =
    modelVal.trim() !== "" && !chatModels.some((m) => m.id === modelVal);

  function handleProviderChange(nextProvider: string) {
    onChange(persona.providerKey, nextProvider);
    const nextChatModels = getProviderModelsByCapability(
      providers,
      nextProvider,
      "chat",
    );
    if (
      nextChatModels.length > 0 &&
      !nextChatModels.some((m) => m.id === modelVal)
    ) {
      const firstId = nextChatModels[0]?.id;
      if (firstId) {
        onChange(persona.modelKey, firstId);
      }
    }
  }

  async function handleResetInstruction() {
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
    try {
      await updateSettings.mutateAsync({
        values: {
          [persona.providerKey]: draft[persona.providerKey] ?? "",
          [persona.modelKey]: draft[persona.modelKey] ?? "",
          [persona.effortKey]: draft[persona.effortKey] ?? "",
          [persona.instructionKey]: draft[persona.instructionKey] ?? "",
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
        <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="text-base font-semibold text-fg">{persona.name}</h2>
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
              onChange={(e) => handleProviderChange(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-fg/40 focus:outline-none"
            >
              <option value="">{t("settings.agents.selectProvider")}</option>
              {providerOptions.map((opt) => (
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
                  onClick={() => setCustomModelMode((prev) => !prev)}
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
                className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
              />
            ) : (
              <select
                id={`setting-${persona.modelKey}`}
                value={modelVal}
                onChange={(e) => onChange(persona.modelKey, e.target.value)}
                className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-fg/40 focus:outline-none"
              >
                {modelAbsentFromCatalog ? (
                  <option value={modelVal}>{modelVal}</option>
                ) : null}
                {chatModels.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name && m.name !== m.id ? `${m.name} (${m.id})` : m.id}
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
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-fg/40 focus:outline-none"
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
              onClick={() => void handleResetInstruction()}
              className="rounded-lg border border-line px-2.5 py-1 text-xs font-medium text-muted transition hover:bg-fg/6 hover:text-fg disabled:cursor-not-allowed disabled:opacity-50"
            >
              {t("settings.agents.resetInstruction")}
            </button>
          </div>
          <textarea
            id={`setting-${persona.instructionKey}`}
            rows={10}
            value={instrVal}
            onChange={(e) => onChange(persona.instructionKey, e.target.value)}
            className="rounded-lg border border-line bg-canvas p-3 font-mono text-xs/relaxed text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
          />
        </div>
      </section>

      <div className="flex justify-end">
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
  );
}
