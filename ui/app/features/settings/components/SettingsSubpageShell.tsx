import { type ReactNode, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMe } from "~/shared/api/session";
import { InductivePageHeader } from "~/shared/ui/iui";
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
  title: string;
  subtitle?: string;
  action?: ReactNode;
  meLoading: boolean;
  canManage: boolean;
  settingsLoading: boolean;
  isEmpty: boolean;
  errorMsg: string | null;
  savedMsg: string | null;
  children: ReactNode;
}

export function SettingsSubpageShell({
  title,
  subtitle,
  action,
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
    <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: t("settings.title"), to: "/settings" },
          { label: "Sistem & Integrasi", to: "/settings/system" },
          { label: title },
        ]}
        title={title}
        subtitle={subtitle ?? ""}
        action={canManage ? action : undefined}
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
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
              className="rounded-xl border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
            >
              {errorMsg}
            </div>
          ) : null}

          {savedMsg ? (
            <div
              role="status"
              className="rounded-xl border border-success-fg/30 bg-success-bg px-4 py-3 text-sm text-success-fg"
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
