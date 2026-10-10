import { Typography } from "antd";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { flattenMenu, getModule, modulePath } from "~/shared/modules/registry";
import { useActiveModule } from "~/shared/modules/useActiveModule";
import type { Route } from "../../routes/+types/_app.$moduleId._index";

export function meta({ params }: Route.MetaArgs) {
  const key = getModule(params.moduleId)?.name;
  return [{ title: key ? `${i18n.t(key)} | Domus` : "Domus" }];
}

/** Beranda modul: ringkasan dan pintasan ke semua halaman di menu modul. */
export default function ModuleHomePage() {
  const { t } = useTranslation();
  const { activeModule } = useActiveModule();
  if (!activeModule) return null;

  const shortcuts = flattenMenu(activeModule.menu);

  return (
    <>
      <ControlPanel breadcrumbs={[{ title: t(activeModule.name) }]} />
      <section className="mx-auto w-full max-w-5xl px-4 py-5">
        <Typography.Paragraph style={{ color: "var(--domus-text-secondary)" }}>
          {t(activeModule.description)}
        </Typography.Paragraph>
        <ul className="grid list-none grid-cols-1 gap-3 p-0 sm:grid-cols-2 md:grid-cols-3">
          {shortcuts.map(({ slug, label, group }) => (
            <li key={slug}>
              <Link
                to={modulePath(activeModule.id, slug)}
                className="glass launcher-tile flex h-full flex-col gap-1 p-4"
              >
                <span style={{ fontWeight: 500 }}>{t(label)}</span>
                {group && (
                  <span
                    style={{ fontSize: 13, color: "var(--domus-text-muted)" }}
                  >
                    {t(group)}
                  </span>
                )}
              </Link>
            </li>
          ))}
        </ul>
      </section>
    </>
  );
}
