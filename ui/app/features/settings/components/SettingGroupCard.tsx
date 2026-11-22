import { useTranslation } from "react-i18next";
import type { SettingGroup } from "../model";

interface SettingGroupCardProps {
  group: SettingGroup;
  draft: Record<string, string>;
  busy: boolean;
  onChange: (key: string, value: string) => void;
  onClear: (group: SettingGroup) => void;
}

export function SettingGroupCard({
  group,
  draft,
  busy,
  onChange,
  onClear,
}: SettingGroupCardProps) {
  const { t } = useTranslation();

  return (
    <section className="rounded-xl border border-line bg-surface p-5 shadow-xs">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <h2 className="text-base font-semibold text-fg">{group.name}</h2>
          <span
            className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
              group.configured
                ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300"
                : "bg-fg/8 text-muted"
            }`}
          >
            {group.configured
              ? t("settings.active")
              : t("settings.unconfigured")}
          </span>
        </div>

        <button
          type="button"
          disabled={busy || !group.configured}
          onClick={() => onClear(group)}
          className="rounded-lg border border-line px-3 py-1.5 text-xs font-medium text-muted transition hover:bg-fg/6 hover:text-fg disabled:cursor-not-allowed disabled:opacity-50"
        >
          {t("settings.clearGroup")}
        </button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        {group.items.map((item) => {
          const key = item.key ?? "";
          const inputId = `setting-${key}`;
          const value = draft[key] ?? "";
          return (
            <div key={key} className="flex flex-col gap-1.5">
              <label htmlFor={inputId} className="text-sm font-medium text-fg">
                {item.label}
              </label>
              <input
                id={inputId}
                type={item.secret ? "password" : "text"}
                value={value}
                onChange={(e) => onChange(key, e.target.value)}
                placeholder={item.placeholder ?? ""}
                autoComplete="off"
                className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none"
              />
              {item.description ? (
                <p className="text-xs text-muted">{item.description}</p>
              ) : null}
            </div>
          );
        })}
      </div>
    </section>
  );
}
