import { Bot, Cpu, Plus } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { InductiveEmptyState, InductiveStepCard } from "~/shared/ui/iui";
import { useSyncProviderModels } from "../api/actions";
import { useAgentProviders } from "../api/queries";
import type { AgentProviderResponse } from "../api/types";
import { getProviderLabel } from "../model";

interface LLMProvidersSectionProps {
  onStatus: (msg: string | null, err: string | null) => void;
}

function errMessage(err: unknown): string {
  return err instanceof ApiError || err instanceof Error
    ? err.message
    : "Terjadi kesalahan.";
}

export function LLMProvidersSection({ onStatus }: LLMProvidersSectionProps) {
  const { t } = useTranslation();
  const { data: providers = [], isLoading } = useAgentProviders();
  const syncModels = useSyncProviderModels();
  const [activeId, setActiveId] = useState<string | null>(null);

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

  return (
    <div className="flex flex-col gap-5">
      {isLoading ? (
        <div className="h-32 animate-pulse rounded-xl border border-line bg-surface" />
      ) : providers.length === 0 ? (
        <InductiveEmptyState
          icon={Cpu}
          title="Belum ada AI Provider"
          description={t("settings.providers.empty")}
          action={
            <Link
              to="/settings/system/ai-provider/new"
              className={buttonClass("primary")}
            >
              <Plus aria-hidden className="size-4" />
              {t("settings.providers.add")}
            </Link>
          }
        />
      ) : (
        providers.map((p) => {
          const models = p.models ?? [];
          const isThisSyncing = syncModels.isPending && activeId === p.id;

          return (
            <section
              key={p.id}
              className="rounded-xl border border-line bg-surface p-4 sm:p-6"
            >
              <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <h2 className="text-base font-semibold text-fg">
                      {p.name}
                    </h2>
                    <Badge tone="neutral">{getProviderLabel(p.provider)}</Badge>
                    <Badge tone="success">
                      {p.apiKeyMasked || t("settings.active")}
                    </Badge>
                    <Badge tone="neutral">
                      {t("settings.providers.modelCount", {
                        count: models.length,
                      })}
                    </Badge>
                  </div>
                  {p.description ? (
                    <p className="mt-1.5 text-sm text-muted">{p.description}</p>
                  ) : null}
                  {p.baseUrl ? (
                    <p className="mt-1 break-all font-mono text-xs text-muted">
                      {p.baseUrl}
                    </p>
                  ) : null}
                </div>

                <div className="flex w-full shrink-0 flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
                  <Button
                    type="button"
                    variant="outline"
                    disabled={syncModels.isPending}
                    onClick={() => void handleSync(p)}
                    className="w-full sm:w-auto"
                  >
                    {isThisSyncing
                      ? t("settings.providers.syncing")
                      : t("settings.providers.sync")}
                  </Button>
                  <Link
                    to={`/settings/system/ai-provider/${p.id}`}
                    className={`${buttonClass("outline")} w-full sm:w-auto`}
                  >
                    Kelola
                  </Link>
                </div>
              </div>

              {models.length > 0 ? (
                <details className="mt-4">
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

      <InductiveStepCard
        to="/settings/system/ai-agent"
        title={t("settings.launcher.aiAgent")}
        description="Atur routing model dan persona untuk setiap agen AI setelah penyedia tersedia."
        icon={Bot}
        plate={4}
      />
    </div>
  );
}
