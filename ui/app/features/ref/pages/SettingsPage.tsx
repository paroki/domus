import { useTranslation } from "react-i18next";
import { Link, useSearchParams } from "react-router";
import { plateBg } from "~/shared/ui/plate";
import { useCurrentIdentity } from "../api/queries";
import { getVisibleSettingsItems } from "../model";

export function meta() {
  return [{ title: "Referensi — Domus" }];
}

export default function SettingsPage() {
  const { t } = useTranslation();
  const [params, setParams] = useSearchParams();
  const { data: me } = useCurrentIdentity();
  const q = (params.get("q") ?? "").trim();
  const items = getVisibleSettingsItems(me?.roles ?? [], q);

  return (
    <div className="flex flex-1 flex-col items-center gap-10 px-6 pt-12 pb-12 sm:pt-16">
      <div className="flex flex-col gap-2 text-center">
        <h1 className="text-[28px] font-bold text-fg">
          {t("ref.settingsTitle")}
        </h1>
        <p className="text-[15px] text-muted">{t("ref.settingsSubtitle")}</p>
      </div>

      <div className="w-full max-w-sm">
        <input
          type="search"
          value={q}
          onChange={(e) => {
            const next = e.target.value;
            setParams(next ? { q: next } : {}, { replace: true });
          }}
          placeholder={t("ref.searchPlaceholder")}
          aria-label={t("ref.searchPlaceholder")}
          className="h-10 w-full rounded-full border border-line bg-surface px-4 text-sm text-fg placeholder:text-muted"
        />
      </div>

      {items.length === 0 ? (
        <p className="text-[15px] text-muted">{t("ref.settingsEmpty")}</p>
      ) : (
        <nav
          aria-label={t("ref.settingsTitle")}
          className="flex max-w-190 flex-wrap justify-center gap-4"
        >
          {items.map((item) => {
            const Icon = item.icon;
            return (
              <Link
                key={item.id}
                to={item.path}
                className="flex w-40 flex-col items-center gap-3 rounded-[18px] px-3 py-4 text-center text-fg no-underline hover:bg-fg/8"
              >
                <span
                  className={`flex size-18 items-center justify-center rounded-[20px] text-white ${plateBg[item.plate]}`}
                >
                  <Icon aria-hidden className="size-8" strokeWidth={1.75} />
                </span>
                <span className="text-sm font-semibold">{item.label}</span>
              </Link>
            );
          })}
        </nav>
      )}
    </div>
  );
}
