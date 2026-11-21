import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { useMe } from "~/shared/api/session";
import { useUpdateSettings } from "../api/actions";
import { useSettings } from "../api/queries";
import {
  buildClearValues,
  groupSettings,
  isSuperAdmin,
  type SettingGroup,
} from "../model";

export function meta() {
  return [{ title: "Pengaturan — Domus" }];
}

function SettingGroupCard({
  group,
  draft,
  busy,
  onChange,
  onClear,
}: {
  group: SettingGroup;
  draft: Record<string, string>;
  busy: boolean;
  onChange: (key: string, value: string) => void;
  onClear: (group: SettingGroup) => void;
}) {
  const { t } = useTranslation();

  return (
    <section className="rounded-xl border border-line bg-surface p-5 shadow-xs">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <h2 className="text-base font-semibold text-fg">{group.name}</h2>
          <span
            className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
              group.configured
                ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300"
                : "bg-fg/8 text-muted"
            }`}
          >
            {group.configured
              ? t("settings.active")
              : t("settings.unconfigured")}
          </span>
        </div>

        <button
          type="button"
          disabled={busy || !group.configured}
          onClick={() => onClear(group)}
          className="rounded-lg border border-line px-3 py-1.5 text-xs font-medium text-muted transition hover:bg-fg/6 hover:text-fg disabled:cursor-not-allowed disabled:opacity-50"
        >
          {t("settings.clearGroup")}
        </button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        {group.items.map((item) => {
          const key = item.key ?? "";
          const inputId = `setting-${key}`;
          const value = draft[key] ?? "";
          return (
            <div key={key} className="flex flex-col gap-1.5">
              <label htmlFor={inputId} className="text-sm font-medium text-fg">
                {item.label}
              </label>
              <input
                id={inputId}
                type={item.secret ? "password" : "text"}
                value={value}
                onChange={(e) => onChange(key, e.target.value)}
                placeholder={item.placeholder ?? ""}
                autoComplete="off"
                className="h-10 rounded-lg border border-line bg-bg px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
              />
              {item.description ? (
                <p className="text-xs text-muted">{item.description}</p>
              ) : null}
            </div>
          );
        })}
      </div>
    </section>
  );
}

export default function SettingsPage() {
  const { t } = useTranslation();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);
  const { data: items, isLoading: settingsLoading } = useSettings(canManage);
  const updateSettings = useUpdateSettings();

  const [draft, setDraft] = useState<Record<string, string>>({});
  const [savedMsg, setSavedMsg] = useState<string | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (!items) return;
    const next: Record<string, string> = {};
    for (const item of items) {
      if (item.key) {
        next[item.key] = item.value ?? "";
      }
    }
    setDraft(next);
  }, [items]);

  const groups = groupSettings(items ?? []);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSavedMsg(null);
    setErrorMsg(null);
    try {
      await updateSettings.mutateAsync({ values: draft });
      setSavedMsg(t("settings.saved"));
    } catch (err) {
      setErrorMsg(
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
    }
  }

  async function handleClearGroup(group: SettingGroup) {
    setSavedMsg(null);
    setErrorMsg(null);
    try {
      await updateSettings.mutateAsync({
        values: buildClearValues(group.items),
      });
      setSavedMsg(t("settings.saved"));
    } catch (err) {
      setErrorMsg(
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-6 py-8">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold text-fg">{t("settings.title")}</h1>
        <p className="text-sm text-muted">{t("settings.subtitle")}</p>
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

      {canManage && !settingsLoading && groups.length === 0 ? (
        <p className="text-sm text-muted">{t("settings.empty")}</p>
      ) : null}

      {canManage && groups.length > 0 ? (
        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
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

          {groups.map((group) => (
            <SettingGroupCard
              key={group.name}
              group={group}
              draft={draft}
              busy={updateSettings.isPending}
              onChange={(key, value) =>
                setDraft((prev) => ({ ...prev, [key]: value }))
              }
              onClear={handleClearGroup}
            />
          ))}

          <div className="flex justify-end">
            <button
              type="submit"
              disabled={updateSettings.isPending}
              className="h-10 rounded-lg bg-fg px-5 text-sm font-semibold text-bg transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {updateSettings.isPending
                ? t("settings.saving")
                : t("settings.save")}
            </button>
          </div>
        </form>
      ) : null}
    </div>
  );
}
