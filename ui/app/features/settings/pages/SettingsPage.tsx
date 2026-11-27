import { useTranslation } from "react-i18next";
import { useMe } from "~/shared/api/session";
import {
  InductiveHubGrid,
  InductivePageHeader,
  InductiveStepCard,
} from "~/shared/ui/iui";
import { getVisibleSettingsHubs } from "../model";

export function meta() {
  return [{ title: "Pengaturan — Domus" }];
}

export default function SettingsPage() {
  const { t } = useTranslation();
  const { data: me, isLoading: meLoading } = useMe();
  const hubs = getVisibleSettingsHubs(me?.roles);

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: t("settings.title") },
        ]}
        title={t("settings.title")}
        subtitle="Pilih area pengaturan yang ingin dikelola sesuai peran Anda."
      />

      {!meLoading && hubs.length === 0 ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Anda tidak memiliki izin untuk mengakses pengaturan.
        </div>
      ) : null}

      {hubs.length > 0 ? (
        <nav aria-label={t("settings.title")}>
          <InductiveHubGrid>
            {hubs.map((hub) => (
              <InductiveStepCard
                key={hub.id}
                to={hub.path}
                title={hub.title}
                description={hub.description}
                icon={hub.icon}
                plate={hub.plate}
              />
            ))}
          </InductiveHubGrid>
        </nav>
      ) : null}
    </div>
  );
}
