import { Monitor, Moon, Palette, Sun } from "lucide-react";
import { useTranslation } from "react-i18next";

import { type Mode, type Theme, useTheme } from "~/shared/theme/theme";

const nextMode: Record<Mode, Mode> = {
  system: "light",
  light: "dark",
  dark: "system",
};
const modeIcon = { system: Monitor, light: Sun, dark: Moon } as const;
const nextTheme: Record<Theme, Theme> = {
  green: "terracotta",
  terracotta: "green",
};

const btn =
  "flex size-10 items-center justify-center rounded-full text-fg hover:bg-fg/10";

export function ThemeToggle() {
  const { t } = useTranslation();
  const { theme, mode, setTheme, setMode } = useTheme();
  const ModeIcon = modeIcon[mode];

  return (
    <>
      <button
        type="button"
        className={btn}
        title={t(`theme.mode.${mode}`)}
        aria-label={t("theme.modeLabel", { mode: t(`theme.mode.${mode}`) })}
        onClick={() => setMode(nextMode[mode])}
      >
        <ModeIcon aria-hidden className="size-5" />
      </button>
      <button
        type="button"
        className={btn}
        title={t(`theme.color.${theme}`)}
        aria-label={t("theme.colorLabel", { theme: t(`theme.color.${theme}`) })}
        onClick={() => setTheme(nextTheme[theme])}
      >
        <Palette aria-hidden className="size-5" />
      </button>
    </>
  );
}
