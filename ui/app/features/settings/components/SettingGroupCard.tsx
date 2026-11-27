import type { FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import type { SettingGroup } from "../model";

interface SettingGroupCardProps {
  group: SettingGroup;
  draft: Record<string, string>;
  editing: boolean;
  busy: boolean;
  onStartEdit: () => void;
  onCancelEdit: () => void;
  onChange: (key: string, value: string) => void;
  onSubmit: (e: FormEvent) => void;
}

export function SettingGroupCard({
  group,
  draft,
  editing,
  busy,
  onStartEdit,
  onCancelEdit,
  onChange,
  onSubmit,
}: SettingGroupCardProps) {
  const { t } = useTranslation();

  return (
    <section className="rounded-xl border border-line bg-surface p-4 sm:p-6">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-2.5">
          <h2 className="text-base font-semibold text-fg">{group.name}</h2>
          <Badge tone={group.configured ? "success" : "neutral"}>
            {group.configured
              ? t("settings.active")
              : t("settings.unconfigured")}
          </Badge>
        </div>

        {!editing ? (
          <Button
            type="button"
            variant="outline"
            disabled={busy}
            onClick={onStartEdit}
            className="w-full sm:w-auto"
          >
            Ubah
          </Button>
        ) : null}
      </div>

      {editing ? (
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <div className="grid gap-4 sm:grid-cols-2">
            {group.items.map((item) => {
              const key = item.key ?? "";
              const inputId = `setting-${key}`;
              const value = draft[key] ?? "";
              return (
                <div key={key} className="flex flex-col gap-1.5">
                  <label
                    htmlFor={inputId}
                    className="text-sm font-medium text-fg"
                  >
                    {item.label}
                  </label>
                  <input
                    id={inputId}
                    type={item.secret ? "password" : "text"}
                    value={value}
                    onChange={(e) => onChange(key, e.target.value)}
                    placeholder={item.placeholder ?? ""}
                    autoComplete="off"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg placeholder:text-muted focus:border-fg/40 focus:outline-none sm:text-sm"
                  />
                  {item.description ? (
                    <p className="text-xs text-muted">{item.description}</p>
                  ) : null}
                </div>
              );
            })}
          </div>

          <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="outline"
              disabled={busy}
              onClick={onCancelEdit}
              className="w-full sm:w-auto"
            >
              {t("settings.providers.cancel")}
            </Button>
            <Button type="submit" disabled={busy} className="w-full sm:w-auto">
              {busy ? t("settings.saving") : t("settings.save")}
            </Button>
          </div>
        </form>
      ) : (
        <dl className="grid gap-4 sm:grid-cols-2">
          {group.items.map((item) => {
            const key = item.key ?? "";
            const displayValue = item.secret
              ? item.configured
                ? "••••••••"
                : "Belum diisi"
              : item.value || "Belum diisi";
            return (
              <div key={key} className="flex flex-col gap-1">
                <dt className="text-xs font-medium text-muted">{item.label}</dt>
                <dd className="break-all font-mono text-sm text-fg">
                  {displayValue}
                </dd>
                {item.description ? (
                  <p className="text-xs text-muted">{item.description}</p>
                ) : null}
              </div>
            );
          })}
        </dl>
      )}
    </section>
  );
}
