import { useTranslation } from "react-i18next";
import { useMe } from "~/shared/api/session";
import {
  InductiveHubGrid,
  InductivePageHeader,
  InductiveStepCard,
} from "~/shared/ui/iui";
import { isSuperAdmin, SETTINGS_LAUNCHERS } from "../model";

export function meta() {
  return [{ title: "Sistem & Integrasi — Domus" }];
}

export default function SystemHubPage() {
  const { t } = useTranslation();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: t("settings.title"), to: "/settings" },
          { label: "Sistem & Integrasi" },
        ]}
        title="Sistem & Integrasi"
        subtitle={t("settings.subtitle")}
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          {t("settings.forbiddenSuperAdmin")}
        </div>
      ) : null}

      {canManage ? (
        <nav aria-label="Sistem & Integrasi">
          <InductiveHubGrid>
            {SETTINGS_LAUNCHERS.map((item, idx) => (
              <InductiveStepCard
                key={item.id}
                to={item.path}
                step={idx + 1}
                title={t(`settings.launcher.${item.id}`)}
                description="Konfigurasi integrasi dan parameter sistem."
                icon={item.icon}
                plate={item.plate}
              />
            ))}
          </InductiveHubGrid>
        </nav>
      ) : null}
    </div>
  );
}
