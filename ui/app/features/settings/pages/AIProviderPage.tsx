import { Plus } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { buttonClass } from "~/shared/ui/Button";
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
      title={t("settings.providers.title")}
      subtitle={t("settings.providers.subtitle")}
      action={
        <Link
          to="/settings/system/ai-provider/new"
          className={buttonClass("primary")}
        >
          <Plus aria-hidden className="size-4" />
          {t("settings.providers.add")}
        </Link>
      }
      meLoading={editor.meLoading}
      canManage={editor.canManage}
      settingsLoading={editor.settingsLoading}
      isEmpty={false}
      errorMsg={editor.errorMsg}
      savedMsg={editor.savedMsg}
    >
      <LLMProvidersSection onStatus={editor.setStatus} />
    </SettingsSubpageShell>
  );
}
