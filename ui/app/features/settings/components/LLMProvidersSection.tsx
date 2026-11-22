import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import {
  useCreateAgentProvider,
  useDeleteAgentProvider,
  useSyncProviderModels,
  useUpdateAgentProvider,
} from "../api/actions";
import { useAgentProviders } from "../api/queries";
import type { AgentProviderResponse } from "../api/types";
import {
  getProviderLabel,
  LLM_PROVIDERS,
  type LLMProviderType,
} from "../model";

interface LLMProvidersSectionProps {
  onStatus: (msg: string | null, err: string | null) => void;
}

interface ProviderFormState {
  name: string;
  description: string;
  provider: LLMProviderType;
  apiKey: string;
  baseUrl: string;
}

const EMPTY_FORM: ProviderFormState = {
  name: "",
  description: "",
  provider: "gemini",
  apiKey: "",
  baseUrl: "",
};

function errMessage(err: unknown): string {
  return err instanceof ApiError || err instanceof Error
    ? err.message
    : "Terjadi kesalahan.";
}

export function LLMProvidersSection({ onStatus }: LLMProvidersSectionProps) {
  const { t } = useTranslation();
  const { data: providers = [], isLoading } = useAgentProviders();
  const createProvider = useCreateAgentProvider();
  const updateProvider = useUpdateAgentProvider();
  const deleteProvider = useDeleteAgentProvider();
  const syncModels = useSyncProviderModels();

  const [editingId, setEditingId] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [showApiKeyInput, setShowApiKeyInput] = useState(false);
  const [form, setForm] = useState<ProviderFormState>(EMPTY_FORM);
  const [activeId, setActiveId] = useState<string | null>(null);

  const busy =
    createProvider.isPending ||
    updateProvider.isPending ||
    deleteProvider.isPending ||
    syncModels.isPending;

  const selectedDef = LLM_PROVIDERS.find((p) => p.id === form.provider);

  function startCreate() {
    onStatus(null, null);
    setEditingId(null);
    setShowApiKeyInput(true);
    setForm(EMPTY_FORM);
    setShowForm(true);
  }

  function startEdit(p: AgentProviderResponse) {
    onStatus(null, null);
    setEditingId(p.id ?? null);
    setShowApiKeyInput(false);
    setForm({
      name: p.name ?? "",
      description: p.description ?? "",
      provider: (p.provider as LLMProviderType) ?? "gemini",
      apiKey: "",
      baseUrl: p.baseUrl ?? "",
    });
    setShowForm(true);
  }

  function cancelForm() {
    setShowForm(false);
    setEditingId(null);
    setShowApiKeyInput(false);
    setForm(EMPTY_FORM);
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    onStatus(null, null);
    try {
      if (editingId) {
        await updateProvider.mutateAsync({
          providerId: editingId,
          body: {
            name: form.name,
            description: form.description,
            provider: form.provider,
            apiKey: form.apiKey,
            baseUrl: form.baseUrl,
          },
        });
      } else {
        await createProvider.mutateAsync({
          name: form.name,
          description: form.description,
          provider: form.provider,
          apiKey: form.apiKey,
          baseUrl: form.baseUrl,
        });
      }
      cancelForm();
      onStatus(t("settings.saved"), null);
    } catch (err) {
      onStatus(null, errMessage(err));
    }
  }

  async function handleSync(p: AgentProviderResponse) {
    if (!p.id) return;
    onStatus(null, null);
    setActiveId(p.id);
    try {
      await syncModels.mutateAsync(p.id);
      onStatus(
        t("settings.providers.synced", { provider: p.name ?? p.provider }),
        null,
      );
    } catch (err) {
      onStatus(null, errMessage(err));
    } finally {
      setActiveId(null);
    }
  }

  async function handleDelete(p: AgentProviderResponse) {
    if (!p.id) return;
    onStatus(null, null);
    setActiveId(p.id);
    try {
      await deleteProvider.mutateAsync(p.id);
      if (editingId === p.id) {
        cancelForm();
      }
      onStatus(t("settings.providers.deleted"), null);
    } catch (err) {
      onStatus(null, errMessage(err));
    } finally {
      setActiveId(null);
    }
  }

  return (
    <div className="flex flex-col gap-5">
      <div className="flex justify-end">
        <button
          type="button"
          disabled={busy}
          onClick={startCreate}
          className="h-9 rounded-lg bg-fg px-4 text-xs font-semibold text-bg transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {t("settings.providers.add")}
        </button>
      </div>

      {showForm ? (
        <form
          onSubmit={handleSubmit}
          className="rounded-xl border border-line bg-surface p-5 shadow-xs"
        >
          <h2 className="mb-4 text-base font-semibold text-fg">
            {editingId
              ? t("settings.providers.edit")
              : t("settings.providers.add")}
          </h2>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1.5">
              <label
                htmlFor="provider-name"
                className="text-sm font-medium text-fg"
              >
                {t("settings.providers.name")}
              </label>
              <input
                id="provider-name"
                type="text"
                required
                maxLength={100}
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="Gemini 1"
                className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label
                htmlFor="provider-type"
                className="text-sm font-medium text-fg"
              >
                {t("settings.providers.providerType")}
              </label>
              <select
                id="provider-type"
                value={form.provider}
                onChange={(e) =>
                  setForm({
                    ...form,
                    provider: e.target.value as LLMProviderType,
                  })
                }
                className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg focus:border-fg/40 focus:outline-none"
              >
                {LLM_PROVIDERS.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="flex flex-col gap-1.5 sm:col-span-2">
              <label
                htmlFor="provider-desc"
                className="text-sm font-medium text-fg"
              >
                {t("settings.providers.description")}
              </label>
              <input
                id="provider-desc"
                type="text"
                maxLength={500}
                value={form.description}
                onChange={(e) =>
                  setForm({ ...form, description: e.target.value })
                }
                placeholder="API key utama paroki..."
                className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <div className="flex items-center justify-between gap-2">
                <label
                  htmlFor="provider-apikey"
                  className="text-sm font-medium text-fg"
                >
                  {t("settings.providers.apiKey")}
                </label>
                {editingId && showApiKeyInput ? (
                  <button
                    type="button"
                    onClick={() => {
                      setShowApiKeyInput(false);
                      setForm({ ...form, apiKey: "" });
                    }}
                    className="text-[11px] font-medium text-muted underline-offset-2 hover:text-fg hover:underline"
                  >
                    {t("settings.providers.cancel")}
                  </button>
                ) : null}
              </div>
              {editingId && !showApiKeyInput ? (
                <button
                  type="button"
                  onClick={() => setShowApiKeyInput(true)}
                  className="flex h-10 w-fit items-center rounded-lg border border-line bg-bg px-3 text-xs font-medium text-fg transition hover:bg-fg/6"
                >
                  {t("settings.providers.updateApiKey")}
                </button>
              ) : (
                <input
                  id="provider-apikey"
                  type="password"
                  required
                  value={form.apiKey}
                  onChange={(e) => setForm({ ...form, apiKey: e.target.value })}
                  placeholder="sk-..."
                  autoComplete="off"
                  className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
                />
              )}
            </div>

            <div className="flex flex-col gap-1.5">
              <label
                htmlFor="provider-baseurl"
                className="text-sm font-medium text-fg"
              >
                {t("settings.providers.baseUrl")}
              </label>
              <input
                id="provider-baseurl"
                type="url"
                value={form.baseUrl}
                onChange={(e) => setForm({ ...form, baseUrl: e.target.value })}
                placeholder={selectedDef?.defaultBaseURL ?? ""}
                autoComplete="off"
                className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
              />
            </div>
          </div>

          <div className="mt-4 flex flex-wrap items-center justify-end gap-2.5">
            <button
              type="button"
              disabled={busy}
              onClick={cancelForm}
              className="h-9 rounded-lg border border-line px-4 text-xs font-medium text-muted transition hover:bg-fg/6 hover:text-fg disabled:cursor-not-allowed disabled:opacity-50"
            >
              {t("settings.providers.cancel")}
            </button>
            <button
              type="submit"
              disabled={busy}
              className="h-9 rounded-lg bg-fg px-4 text-xs font-semibold text-bg transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {t("settings.providers.save")}
            </button>
          </div>
        </form>
      ) : null}

      {isLoading ? (
        <div className="h-32 animate-pulse rounded-xl border border-line bg-surface" />
      ) : providers.length === 0 ? (
        <p className="rounded-xl border border-line bg-surface p-6 text-center text-sm text-muted">
          {t("settings.providers.empty")}
        </p>
      ) : (
        providers.map((p) => {
          const models = p.models ?? [];
          const isThisSyncing = syncModels.isPending && activeId === p.id;

          return (
            <section
              key={p.id}
              className="rounded-xl border border-line bg-surface p-5 shadow-xs"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h2 className="text-base font-semibold text-fg">
                      {p.name}
                    </h2>
                    <span className="inline-flex items-center rounded-full bg-fg/8 px-2.5 py-0.5 text-xs font-medium text-fg">
                      {getProviderLabel(p.provider)}
                    </span>
                    <span className="inline-flex items-center rounded-full bg-emerald-500/15 px-2.5 py-0.5 text-xs font-medium text-emerald-700 dark:text-emerald-300">
                      {p.apiKeyMasked || t("settings.active")}
                    </span>
                    <span className="inline-flex items-center rounded-full bg-fg/6 px-2.5 py-0.5 text-xs text-muted">
                      {t("settings.providers.modelCount", {
                        count: models.length,
                      })}
                    </span>
                  </div>
                  {p.description ? (
                    <p className="mt-1 text-xs text-muted">{p.description}</p>
                  ) : null}
                  {p.baseUrl ? (
                    <p className="mt-0.5 font-mono text-[11px] text-muted">
                      {p.baseUrl}
                    </p>
                  ) : null}
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  <button
                    type="button"
                    disabled={busy}
                    onClick={() => void handleSync(p)}
                    className="rounded-lg border border-line px-3 py-1.5 text-xs font-medium text-fg transition hover:bg-fg/6 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {isThisSyncing
                      ? t("settings.providers.syncing")
                      : t("settings.providers.sync")}
                  </button>
                  <button
                    type="button"
                    disabled={busy}
                    onClick={() => startEdit(p)}
                    className="rounded-lg border border-line px-3 py-1.5 text-xs font-medium text-fg transition hover:bg-fg/6 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {t("settings.providers.edit")}
                  </button>
                  <button
                    type="button"
                    disabled={busy}
                    onClick={() => void handleDelete(p)}
                    className="rounded-lg border border-red-500/30 px-3 py-1.5 text-xs font-medium text-red-600 transition hover:bg-red-500/10 disabled:cursor-not-allowed disabled:opacity-50 dark:text-red-400"
                  >
                    {t("settings.providers.delete")}
                  </button>
                </div>
              </div>

              {models.length > 0 ? (
                <details className="mt-3">
                  <summary className="cursor-pointer text-xs font-medium text-muted hover:text-fg">
                    {t("settings.providers.catalogToggle", {
                      count: models.length,
                    })}
                  </summary>
                  <div className="mt-2 flex flex-wrap gap-1.5">
                    {models.map((m) => (
                      <span
                        key={m.id}
                        className="rounded-md bg-fg/6 px-2 py-0.5 font-mono text-[11px] text-muted"
                      >
                        {m.id}
                      </span>
                    ))}
                  </div>
                </details>
              ) : null}
            </section>
          );
        })
      )}
    </div>
  );
}
