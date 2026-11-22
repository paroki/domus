import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { useMe } from "~/shared/api/session";
import { plateBg } from "~/shared/ui/plate";
import { isSuperAdmin, SETTINGS_LAUNCHERS } from "../model";

export function meta() {
  return [{ title: "Pengaturan — Domus" }];
}

export default function SettingsPage() {
  const { t } = useTranslation();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  return (
    <div className="flex flex-1 flex-col items-center gap-12 px-6 pt-16 pb-12 sm:pt-[72px]">
      <div className="flex flex-col gap-2 text-center">
        <h1 className="text-[28px] font-bold text-fg">{t("settings.title")}</h1>
        <p className="text-[15px] text-muted">{t("settings.subtitle")}</p>
      </div>

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="w-full max-w-xl rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-800 dark:text-amber-200"
        >
          {t("settings.forbiddenSuperAdmin")}
        </div>
      ) : null}

      {canManage ? (
        <nav
          aria-label={t("settings.title")}
          className="flex max-w-190 flex-wrap justify-center gap-4"
        >
          {SETTINGS_LAUNCHERS.map((item) => {
            const Icon = item.icon;
            return (
              <Link
                key={item.id}
                to={item.path}
                className="flex w-34 flex-col items-center gap-3 rounded-[18px] px-2 py-4 text-center text-fg no-underline hover:bg-fg/8"
              >
                <span
                  className={`flex size-18 items-center justify-center rounded-[20px] text-white ${plateBg[item.plate]}`}
                >
                  <Icon aria-hidden className="size-8" strokeWidth={1.75} />
                </span>
                <span className="text-sm font-semibold">
                  {t(`settings.launcher.${item.id}`)}
                </span>
              </Link>
            );
          })}
        </nav>
      ) : null}
    </div>
  );
}
