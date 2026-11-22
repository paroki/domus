import type { FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { useUpdateSettings } from "../api/actions";
import {
  buildClearValues,
  findChannelGroup,
  type SettingGroup,
} from "../model";
import { SettingGroupCard } from "./SettingGroupCard";
import {
  SettingsSubpageShell,
  useSettingsEditor,
} from "./SettingsSubpageShell";

interface ChannelSettingViewProps {
  groupName: string;
  title: string;
  subtitle: string;
}

export function ChannelSettingView({
  groupName,
  title,
  subtitle,
}: ChannelSettingViewProps) {
  const { t } = useTranslation();
  const editor = useSettingsEditor();
  const updateSettings = useUpdateSettings();
  const group = findChannelGroup(editor.items, groupName);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!group) return;
    editor.setStatus(null, null);
    const values: Record<string, string> = {};
    for (const item of group.items) {
      if (item.key) {
        values[item.key] = editor.draft[item.key] ?? "";
      }
    }
    try {
      await updateSettings.mutateAsync({ values });
      editor.setStatus(t("settings.saved"), null);
    } catch (err) {
      editor.setStatus(
        null,
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
    }
  }

  async function handleClear(target: SettingGroup) {
    editor.setStatus(null, null);
    try {
      await updateSettings.mutateAsync({
        values: buildClearValues(target.items),
      });
      editor.setStatus(t("settings.saved"), null);
    } catch (err) {
      editor.setStatus(
        null,
        err instanceof ApiError || err instanceof Error
          ? err.message
          : "Terjadi kesalahan.",
      );
    }
  }

  return (
    <SettingsSubpageShell
      backTo="/settings"
      backLabel={t("settings.backToSettings")}
      title={title}
      subtitle={subtitle}
      meLoading={editor.meLoading}
      canManage={editor.canManage}
      settingsLoading={editor.settingsLoading}
      isEmpty={editor.items.length === 0}
      errorMsg={editor.errorMsg}
      savedMsg={editor.savedMsg}
    >
      {!group ? (
        <p className="text-sm text-muted">
          {t("settings.channels.notConfiguredGroup")}
        </p>
      ) : (
        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          <SettingGroupCard
            group={group}
            draft={editor.draft}
            busy={updateSettings.isPending}
            onChange={editor.updateDraft}
            onClear={handleClear}
          />

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
      )}
    </SettingsSubpageShell>
  );
}
