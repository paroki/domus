import { useTranslation } from "react-i18next";
import { useParams } from "react-router";
import { AgentPersonaForm } from "../components/AgentPersonaForm";
import {
  SettingsSubpageShell,
  useSettingsEditor,
} from "../components/SettingsSubpageShell";
import { findAgentPersona } from "../model";

export function meta() {
  return [{ title: "Konfigurasi Agen — Domus" }];
}

export default function AIAgentDetailPage() {
  const { t } = useTranslation();
  const { agentId } = useParams();
  const persona = findAgentPersona(agentId);
  const editor = useSettingsEditor();

  return (
    <SettingsSubpageShell
      backTo="/settings/ai-agent"
      backLabel={t("settings.backToAgents")}
      title={persona ? persona.name : t("settings.agents.title")}
      subtitle={persona?.role}
      meLoading={editor.meLoading}
      canManage={editor.canManage}
      settingsLoading={editor.settingsLoading}
      isEmpty={Boolean(persona) && editor.items.length === 0}
      errorMsg={editor.errorMsg}
      savedMsg={editor.savedMsg}
    >
      {!persona ? (
        <p className="text-sm text-muted">{t("settings.agents.notFound")}</p>
      ) : (
        <AgentPersonaForm
          persona={persona}
          itemsByKey={editor.itemsByKey}
          draft={editor.draft}
          onChange={editor.updateDraft}
          onStatus={editor.setStatus}
        />
      )}
    </SettingsSubpageShell>
  );
}
