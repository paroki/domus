import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { plateBg } from "~/shared/ui/plate";
import { EmbeddingConfigSection } from "../components/EmbeddingConfigSection";
import {
  SettingsSubpageShell,
  useSettingsEditor,
} from "../components/SettingsSubpageShell";
import { AGENT_PERSONAS } from "../model";

export function meta() {
  return [{ title: "AI Agent — Domus" }];
}

export default function AIAgentPage() {
  const { t } = useTranslation();
  const editor = useSettingsEditor();

  return (
    <SettingsSubpageShell
      backTo="/settings"
      backLabel={t("settings.backToSettings")}
      title={t("settings.agents.title")}
      subtitle={t("settings.agents.subtitle")}
      meLoading={editor.meLoading}
      canManage={editor.canManage}
      settingsLoading={editor.settingsLoading}
      isEmpty={editor.items.length === 0}
      errorMsg={editor.errorMsg}
      savedMsg={editor.savedMsg}
    >
      <nav
        aria-label={t("settings.agents.title")}
        className="flex flex-wrap justify-center gap-4 py-4"
      >
        {AGENT_PERSONAS.map((persona) => {
          const Icon = persona.icon;
          return (
            <Link
              key={persona.id}
              to={`/settings/ai-agent/${persona.id}`}
              className="flex w-34 flex-col items-center gap-3 rounded-[18px] px-2 py-4 text-center text-fg no-underline hover:bg-fg/8"
            >
              <span
                className={`flex size-18 items-center justify-center rounded-[20px] text-white ${plateBg[persona.plate]}`}
              >
                <Icon aria-hidden className="size-8" strokeWidth={1.75} />
              </span>
              <span className="text-sm font-semibold">{persona.name}</span>
            </Link>
          );
        })}
      </nav>

      <EmbeddingConfigSection
        itemsByKey={editor.itemsByKey}
        draft={editor.draft}
        onChange={editor.updateDraft}
        onStatus={editor.setStatus}
      />
    </SettingsSubpageShell>
  );
}
