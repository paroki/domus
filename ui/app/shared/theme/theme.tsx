import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

export type Theme = "green" | "terracotta";
export type Mode = "system" | "light" | "dark";

const THEME_KEY = "domus.theme";
const MODE_KEY = "domus.mode";
const DARK_QUERY = "(prefers-color-scheme: dark)";

/**
 * Runs inline in <head> before first paint so the page never flashes the wrong
 * theme. Keep in sync with `apply` below.
 */
export const themeInitScript = `(function(){try{var d=document.documentElement,t=localStorage.getItem("${THEME_KEY}"),m=localStorage.getItem("${MODE_KEY}")||"system",k=m==="dark"||(m!=="light"&&matchMedia("${DARK_QUERY}").matches);d.dataset.theme=t==="terracotta"?"terracotta":"green";d.dataset.mode=k?"dark":"light"}catch(e){}})()`;

function read<T extends string>(key: string, allowed: readonly T[], fallback: T): T {
  try {
    const v = localStorage.getItem(key);
    return allowed.includes(v as T) ? (v as T) : fallback;
  } catch {
    return fallback;
  }
}

function write(key: string, value: string) {
  try {
    localStorage.setItem(key, value);
  } catch {
    // storage unavailable (private mode); the choice just won't persist.
  }
}

function apply(theme: Theme, mode: Mode) {
  const dark = mode === "dark" || (mode === "system" && matchMedia(DARK_QUERY).matches);
  const el = document.documentElement;
  el.dataset.theme = theme;
  el.dataset.mode = dark ? "dark" : "light";
}

type ThemeContextValue = {
  theme: Theme;
  mode: Mode;
  setTheme: (theme: Theme) => void;
  setMode: (mode: Mode) => void;
};

const ThemeContext = createContext<ThemeContextValue | null>(null);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>("green");
  const [mode, setModeState] = useState<Mode>("system");

  // Sync from storage after mount (SSR-safe: the server shell has no storage).
  useEffect(() => {
    setThemeState(read<Theme>(THEME_KEY, ["green", "terracotta"], "green"));
    setModeState(read<Mode>(MODE_KEY, ["system", "light", "dark"], "system"));
  }, []);

  // Re-apply on change, and follow the OS while mode is "system".
  useEffect(() => {
    apply(theme, mode);
    if (mode !== "system") return;
    const mq = matchMedia(DARK_QUERY);
    const onChange = () => apply(theme, mode);
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, [theme, mode]);

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t);
    write(THEME_KEY, t);
  }, []);
  const setMode = useCallback((m: Mode) => {
    setModeState(m);
    write(MODE_KEY, m);
  }, []);

  const value = useMemo(
    () => ({ theme, mode, setTheme, setMode }),
    [theme, mode, setTheme, setMode],
  );
  return <ThemeContext value={value}>{children}</ThemeContext>;
}

export function useTheme() {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useTheme must be used inside <ThemeProvider>");
  return ctx;
}
