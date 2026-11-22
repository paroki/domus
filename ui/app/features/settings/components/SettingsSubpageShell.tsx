import { ArrowLeft } from "lucide-react";
import { type ReactNode, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { useMe } from "~/shared/api/session";
import { useSettings } from "../api/queries";
import type { SettingResponse } from "../api/types";
import { buildSettingsDraft, indexSettingsByKey, isSuperAdmin } from "../model";

export function useSettingsEditor() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);
  const { data: items, isLoading: settingsLoading } = useSettings(canManage);

  const [draft, setDraft] = useState<Record<string, string>>({});
  const [savedMsg, setSavedMsg] = useState<string | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (!items) return;
    setDraft(buildSettingsDraft(items));
  }, [items]);

  const itemsByKey = useMemo<Record<string, SettingResponse>>(
    () => indexSettingsByKey(items),
    [items],
  );

  function updateDraft(key: string, value: string) {
    setDraft((prev) => ({ ...prev, [key]: value }));
  }

  function setStatus(msg: string | null, err: string | null) {
    setSavedMsg(msg);
    setErrorMsg(err);
  }

  return {
    meLoading,
    canManage,
    settingsLoading,
    items: items ?? [],
    itemsByKey,
    draft,
    updateDraft,
    savedMsg,
    errorMsg,
    setStatus,
  };
}

interface SettingsSubpageShellProps {
  backTo: string;
  backLabel: string;
  title: string;
  subtitle?: string;
  meLoading: boolean;
  canManage: boolean;
  settingsLoading: boolean;
  isEmpty: boolean;
  errorMsg: string | null;
  savedMsg: string | null;
  children: ReactNode;
}

export function SettingsSubpageShell({
  backTo,
  backLabel,
  title,
  subtitle,
  meLoading,
  canManage,
  settingsLoading,
  isEmpty,
  errorMsg,
  savedMsg,
  children,
}: SettingsSubpageShellProps) {
  const { t } = useTranslation();

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-6 py-8">
      <div>
        <Link
          to={backTo}
          className="inline-flex items-center gap-1.5 text-sm font-medium text-muted no-underline hover:text-fg"
        >
          <ArrowLeft aria-hidden className="size-4" />
          {backLabel}
        </Link>
        <h1 className="mt-2 text-2xl font-bold text-fg">{title}</h1>
        {subtitle ? <p className="text-sm text-muted">{subtitle}</p> : null}
      </div>

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-800 dark:text-amber-200"
        >
          {t("settings.forbiddenSuperAdmin")}
        </div>
      ) : null}

      {canManage && settingsLoading ? (
        <p className="text-sm text-muted">{t("common.loading")}</p>
      ) : null}

      {canManage && !settingsLoading && isEmpty ? (
        <p className="text-sm text-muted">{t("settings.empty")}</p>
      ) : null}

      {canManage && !settingsLoading && !isEmpty ? (
        <div className="flex flex-col gap-5">
          {errorMsg ? (
            <div
              role="alert"
              className="rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-700 dark:text-red-300"
            >
              {errorMsg}
            </div>
          ) : null}

          {savedMsg ? (
            <div
              role="status"
              className="rounded-xl border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-300"
            >
              {savedMsg}
            </div>
          ) : null}

          {children}
        </div>
      ) : null}
    </div>
  );
}
