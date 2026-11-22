import { useTranslation } from "react-i18next";
import { Link, useSearchParams } from "react-router";

import { apps } from "~/apps";
import { AgentWorkspace } from "~/features/agent/components/AgentWorkspace";
import { useMe } from "~/features/user/api/queries";
import i18n from "~/shared/i18n";
import { plateBg } from "~/shared/ui/plate";

import type { Route } from "./+types/HomePage";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Domus" },
    { name: "description", content: i18n.t("home.subtitle") },
  ];
}

export default function HomePage() {
  const { t } = useTranslation();
  const [params] = useSearchParams();
  const { data: me } = useMe();
  const tab = params.get("tab") === "menu" ? "menu" : "gabriel";

  if (tab === "gabriel") {
    return <AgentWorkspace />;
  }

  const q = (params.get("q") ?? "").trim();
  const userRoles = me?.roles ?? [];

  const visible = apps.filter((app) => {
    if (app.roles?.length && !app.roles.some((r) => userRoles.includes(r))) {
      return false;
    }
    return t(`apps.${app.id}.name`).toLowerCase().includes(q.toLowerCase());
  });

  return (
    <div className="flex flex-1 flex-col items-center gap-12 px-6 pt-16 pb-12 sm:pt-[72px]">
      <div className="flex flex-col gap-2 text-center">
        <h1 className="text-[28px] font-bold">{t("home.title")}</h1>
        <p className="text-[15px] text-muted">{t("home.subtitle")}</p>
      </div>

      {visible.length === 0 ? (
        <p className="text-[15px] text-muted">{t("home.empty", { q })}</p>
      ) : (
        <nav
          aria-label={t("home.menuLabel")}
          className="flex max-w-190 flex-wrap justify-center gap-4"
        >
          {visible.map((app) => {
            const Icon = app.icon;
            return (
              <Link
                key={app.id}
                to={app.path}
                className="flex w-34 flex-col items-center gap-3 rounded-[18px] px-2 py-4 text-fg no-underline hover:bg-fg/8"
              >
                <span
                  className={`flex size-18 items-center justify-center rounded-[20px] text-white ${plateBg[app.plate]}`}
                >
                  <Icon aria-hidden className="size-8" strokeWidth={1.75} />
                </span>
                <span className="text-sm font-semibold">
                  {t(`apps.${app.id}.name`)}
                </span>
              </Link>
            );
          })}
        </nav>
      )}
    </div>
  );
}
