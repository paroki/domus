import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { ApiError } from "~/shared/api/client";
import { useMe } from "~/shared/api/session";
import { Button, buttonClass } from "~/shared/ui/Button";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import {
  InductiveDangerZone,
  InductiveFormSection,
  InductivePageHeader,
} from "~/shared/ui/iui";
import {
  useCreateAgentProvider,
  useDeleteAgentProvider,
  useUpdateAgentProvider,
} from "../api/actions";
import { useAgentProviders } from "../api/queries";
import { isSuperAdmin, LLM_PROVIDERS, type LLMProviderType } from "../model";

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

const inputCls =
  "h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none sm:text-sm";

function errMessage(err: unknown): string {
  return err instanceof ApiError || err instanceof Error
    ? err.message
    : "Terjadi kesalahan.";
}

export function meta() {
  return [{ title: "Konfigurasi AI Provider — Domus" }];
}

export default function AIProviderFormPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const isEdit = Boolean(id);

  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);
  const { data: providers = [], isLoading: providersLoading } =
    useAgentProviders(canManage && isEdit);

  const provider = isEdit
    ? providers.find((item) => item.id === id)
    : undefined;

  const createProvider = useCreateAgentProvider();
  const updateProvider = useUpdateAgentProvider();
  const deleteProvider = useDeleteAgentProvider();

  const [form, setForm] = useState<ProviderFormState>(EMPTY_FORM);
  const [showApiKeyInput, setShowApiKeyInput] = useState(!isEdit);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);

  useEffect(() => {
    if (!isEdit) {
      setForm(EMPTY_FORM);
      setShowApiKeyInput(true);
      return;
    }
    if (!provider) return;
    setForm({
      name: provider.name ?? "",
      description: provider.description ?? "",
      provider: (provider.provider as LLMProviderType) ?? "gemini",
      apiKey: "",
      baseUrl: provider.baseUrl ?? "",
    });
    setShowApiKeyInput(false);
  }, [isEdit, provider]);

  const busy =
    createProvider.isPending ||
    updateProvider.isPending ||
    deleteProvider.isPending;

  const selectedDef = LLM_PROVIDERS.find((p) => p.id === form.provider);
  const pageTitle = isEdit
    ? (provider?.name ?? t("settings.providers.edit"))
    : t("settings.providers.add");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      if (isEdit && id) {
        await updateProvider.mutateAsync({
          providerId: id,
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
      void navigate("/settings/system/ai-provider");
    } catch (err) {
      setErrorMsg(errMessage(err));
    }
  }

  async function handleDelete() {
    if (!id) return;
    setErrorMsg(null);
    try {
      await deleteProvider.mutateAsync(id);
      setConfirmDeleteOpen(false);
      void navigate("/settings/system/ai-provider");
    } catch (err) {
      setConfirmDeleteOpen(false);
      setErrorMsg(errMessage(err));
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: t("settings.title"), to: "/settings" },
          { label: "Sistem & Integrasi", to: "/settings/system" },
          {
            label: t("settings.providers.title"),
            to: "/settings/system/ai-provider",
          },
          { label: pageTitle },
        ]}
        title={pageTitle}
        subtitle={t("settings.providers.subtitle")}
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          {t("settings.forbiddenSuperAdmin")}
        </div>
      ) : null}

      {canManage && isEdit && providersLoading ? (
        <p className="text-sm text-muted">{t("common.loading")}</p>
      ) : null}

      {canManage && isEdit && !providersLoading && !provider ? (
        <div
          role="alert"
          className="flex flex-col gap-3 rounded-xl border border-danger-fg/30 bg-danger-bg p-4 text-sm text-danger-fg sm:flex-row sm:items-center sm:justify-between"
        >
          <span>Konfigurasi penyedia LLM tidak ditemukan.</span>
          <Link
            to="/settings/system/ai-provider"
            className={`${buttonClass("outline")} w-full sm:w-auto`}
          >
            Kembali ke AI Provider
          </Link>
        </div>
      ) : null}

      {canManage && (!isEdit || Boolean(provider)) ? (
        <div className="flex flex-col gap-5">
          {errorMsg ? (
            <div
              role="alert"
              className="rounded-xl border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
            >
              {errorMsg}
            </div>
          ) : null}

          <form onSubmit={handleSubmit} className="flex flex-col gap-5">
            <InductiveFormSection
              title={
                isEdit
                  ? t("settings.providers.edit")
                  : t("settings.providers.add")
              }
              description="Atur kredensial API dan endpoint penyedia LLM."
            >
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
                    className={inputCls}
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
                    className={inputCls}
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
                    className={inputCls}
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
                    {isEdit && showApiKeyInput ? (
                      <button
                        type="button"
                        onClick={() => {
                          setShowApiKeyInput(false);
                          setForm({ ...form, apiKey: "" });
                        }}
                        className="text-xs font-medium text-muted underline-offset-2 hover:text-fg hover:underline"
                      >
                        {t("settings.providers.cancel")}
                      </button>
                    ) : null}
                  </div>
                  {isEdit && !showApiKeyInput ? (
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => setShowApiKeyInput(true)}
                      className="w-full sm:w-fit"
                    >
                      {t("settings.providers.updateApiKey")}
                    </Button>
                  ) : (
                    <input
                      id="provider-apikey"
                      type="password"
                      required={!isEdit}
                      value={form.apiKey}
                      onChange={(e) =>
                        setForm({ ...form, apiKey: e.target.value })
                      }
                      placeholder="sk-..."
                      autoComplete="off"
                      className={inputCls}
                    />
                  )}
                  {isEdit ? (
                    <p className="text-xs text-muted">
                      {t("settings.providers.apiKeyKeepHint")}
                    </p>
                  ) : null}
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
                    onChange={(e) =>
                      setForm({ ...form, baseUrl: e.target.value })
                    }
                    placeholder={selectedDef?.defaultBaseURL ?? ""}
                    autoComplete="off"
                    className={inputCls}
                  />
                </div>
              </div>

              <div className="mt-2 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
                <Link
                  to="/settings/system/ai-provider"
                  className={`${buttonClass("outline")} w-full sm:w-auto`}
                >
                  {t("settings.providers.cancel")}
                </Link>
                <Button
                  type="submit"
                  disabled={busy}
                  className="w-full sm:w-auto"
                >
                  {t("settings.providers.save")}
                </Button>
              </div>
            </InductiveFormSection>
          </form>

          {isEdit && provider ? (
            <InductiveDangerZone
              title="Hapus Konfigurasi"
              description="Menghapus penyedia ini akan gagal jika masih digunakan oleh agen AI atau embedding."
              action={
                <Button
                  type="button"
                  variant="danger"
                  disabled={busy}
                  onClick={() => setConfirmDeleteOpen(true)}
                  className="w-full sm:w-auto"
                >
                  {t("settings.providers.delete")}
                </Button>
              }
            />
          ) : null}
        </div>
      ) : null}

      <ConfirmDialog
        open={confirmDeleteOpen}
        title="Hapus Konfigurasi AI Provider?"
        description={`Konfigurasi "${provider?.name ?? ""}" akan dihapus permanen.`}
        confirmLabel={t("settings.providers.delete")}
        pending={deleteProvider.isPending}
        onConfirm={() => void handleDelete()}
        onCancel={() => setConfirmDeleteOpen(false)}
      />
    </div>
  );
}
