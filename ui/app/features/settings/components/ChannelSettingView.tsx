import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { Button } from "~/shared/ui/Button";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import { InductiveDangerZone } from "~/shared/ui/iui";
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
  const [editing, setEditing] = useState(false);
  const [confirmClearOpen, setConfirmClearOpen] = useState(false);

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
      setEditing(false);
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
      setConfirmClearOpen(false);
      setEditing(false);
      editor.setStatus(t("settings.saved"), null);
    } catch (err) {
      setConfirmClearOpen(false);
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
        <>
          <SettingGroupCard
            group={group}
            draft={editor.draft}
            editing={editing}
            busy={updateSettings.isPending}
            onStartEdit={() => {
              editor.setStatus(null, null);
              setEditing(true);
            }}
            onCancelEdit={() => {
              setEditing(false);
              for (const item of group.items) {
                if (item.key) {
                  editor.updateDraft(item.key, item.value ?? "");
                }
              }
            }}
            onChange={editor.updateDraft}
            onSubmit={(e) => void handleSubmit(e)}
          />

          <InductiveDangerZone
            title={t("settings.clearGroup")}
            description="Menghapus kredensial akan menonaktifkan integrasi kanal ini sampai dikonfigurasi ulang."
            action={
              <Button
                type="button"
                variant="danger"
                disabled={updateSettings.isPending || !group.configured}
                onClick={() => setConfirmClearOpen(true)}
                className="w-full sm:w-auto"
              >
                {t("settings.clearGroup")}
              </Button>
            }
          />

          <ConfirmDialog
            open={confirmClearOpen}
            title={`${t("settings.clearGroup")}?`}
            description={`Semua nilai konfigurasi pada ${group.name} akan dikosongkan.`}
            confirmLabel={t("settings.clearGroup")}
            pending={updateSettings.isPending}
            onConfirm={() => void handleClear(group)}
            onCancel={() => setConfirmClearOpen(false)}
          />
        </>
      )}
    </SettingsSubpageShell>
  );
}
