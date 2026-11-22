import { useTranslation } from "react-i18next";
import { LLMProvidersSection } from "../components/LLMProvidersSection";
import {
  SettingsSubpageShell,
  useSettingsEditor,
} from "../components/SettingsSubpageShell";

export function meta() {
  return [{ title: "AI Provider — Domus" }];
}

export default function AIProviderPage() {
  const { t } = useTranslation();
  const editor = useSettingsEditor();

  return (
    <SettingsSubpageShell
      backTo="/settings"
      backLabel={t("settings.backToSettings")}
      title={t("settings.providers.title")}
      subtitle={t("settings.providers.subtitle")}
      meLoading={editor.meLoading}
      canManage={editor.canManage}
      settingsLoading={editor.settingsLoading}
      isEmpty={editor.items.length === 0}
      errorMsg={editor.errorMsg}
      savedMsg={editor.savedMsg}
    >
      <LLMProvidersSection onStatus={editor.setStatus} />
    </SettingsSubpageShell>
  );
}
