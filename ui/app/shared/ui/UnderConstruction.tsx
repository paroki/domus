import { Construction } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import type { AppDef } from "../app-def";
import { Badge } from "./Badge";
import { buttonClass } from "./Button";
import { plateBg } from "./plate";

/** Placeholder page for a feature that is registered but not built yet. */
export function UnderConstruction({ app }: { app: AppDef }) {
  const { t } = useTranslation();
  const Icon = app.icon;
  const name = t(`apps.${app.id}.name`);

  return (
    <section
      aria-labelledby="uc-title"
      className="mx-auto flex w-full max-w-xl flex-col items-center gap-6 px-6 py-16 text-center sm:py-24"
    >
      <div className="relative">
        {/* Soft hazard-stripe halo behind the icon plate. */}
        <div
          aria-hidden
          className="absolute -inset-4 rounded-[28px] opacity-30"
          style={{
            backgroundImage:
              "repeating-linear-gradient(-45deg, var(--accent) 0 8px, transparent 8px 20px)",
          }}
        />
        <div
          className={`relative flex size-24 items-center justify-center rounded-[20px] text-white shadow-lg ${plateBg[app.plate]}`}
        >
          <Icon aria-hidden className="size-11" strokeWidth={1.75} />
          <span className="absolute -right-3 -bottom-3 flex size-10 items-center justify-center rounded-full border-4 border-canvas bg-accent text-canvas">
            <Construction aria-hidden className="size-5" strokeWidth={2} />
          </span>
        </div>
      </div>

      <Badge tone="warning">{t("underConstruction.badge")}</Badge>

      <div className="flex flex-col gap-2">
        <h1 id="uc-title" className="text-[28px] leading-tight font-bold">
          {t("underConstruction.title", { name })}
        </h1>
        <p className="text-[15px] text-muted">
          {t(`apps.${app.id}.description`)}
        </p>
        <p className="text-[15px] text-muted">
          {t("underConstruction.description")}
        </p>
      </div>

      <Link to="/" className={buttonClass("outline")}>
        {t("underConstruction.back")}
      </Link>
    </section>
  );
}
