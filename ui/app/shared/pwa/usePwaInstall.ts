import { useCallback, useEffect, useState } from "react";

export const PWA_BANNER_DISMISSED_KEY = "domus:pwa-banner-dismissed";

export interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

let deferredPrompt: BeforeInstallPromptEvent | null = null;
const promptListeners = new Set<() => void>();

function notifyPromptListeners() {
  for (const listener of promptListeners) {
    listener();
  }
}

export function setDeferredInstallPrompt(
  ev: BeforeInstallPromptEvent | null,
): void {
  deferredPrompt = ev;
  notifyPromptListeners();
}

export function getDeferredInstallPrompt(): BeforeInstallPromptEvent | null {
  return deferredPrompt;
}

if (typeof window !== "undefined") {
  window.addEventListener("beforeinstallprompt", (event) => {
    event.preventDefault();
    setDeferredInstallPrompt(event as BeforeInstallPromptEvent);
  });
  window.addEventListener("appinstalled", () => {
    setDeferredInstallPrompt(null);
  });
}

export function isStandaloneMode(
  win: Pick<Window, "matchMedia" | "navigator"> | undefined,
): boolean {
  if (!win) return false;
  const matchStandalone =
    typeof win.matchMedia === "function" &&
    win.matchMedia("(display-mode: standalone)").matches;
  const iosStandalone =
    (win.navigator as { standalone?: boolean } | undefined)?.standalone ===
    true;
  return Boolean(matchStandalone || iosStandalone);
}

export function isIosDevice(
  nav: Pick<Navigator, "userAgent" | "platform" | "maxTouchPoints"> | undefined,
): boolean {
  if (!nav) return false;
  return (
    /iphone|ipad|ipod/i.test(nav.userAgent) ||
    (nav.platform === "MacIntel" && nav.maxTouchPoints > 1)
  );
}

export function readBannerDismissed(): boolean {
  if (typeof window === "undefined") return false;
  try {
    return window.localStorage.getItem(PWA_BANNER_DISMISSED_KEY) === "1";
  } catch {
    return false;
  }
}

export function persistBannerDismissed(): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(PWA_BANNER_DISMISSED_KEY, "1");
  } catch {
    // Ignore storage errors in restricted browsing modes.
  }
}

export function usePwaInstall() {
  const [hasNativePrompt, setHasNativePrompt] = useState(
    () => deferredPrompt !== null,
  );
  const [isStandalone, setIsStandalone] = useState(() =>
    typeof window !== "undefined" ? isStandaloneMode(window) : false,
  );
  const [isIosInstallable, setIsIosInstallable] = useState(() => {
    if (typeof window === "undefined") return false;
    const standalone = isStandaloneMode(window);
    return !standalone && isIosDevice(window.navigator);
  });
  const [isDismissed, setIsDismissed] = useState(() => readBannerDismissed());

  useEffect(() => {
    const syncPrompt = () => setHasNativePrompt(deferredPrompt !== null);
    promptListeners.add(syncPrompt);
    syncPrompt();

    const standalone = isStandaloneMode(window);
    setIsStandalone(standalone);
    setIsIosInstallable(!standalone && isIosDevice(window.navigator));

    const mq =
      typeof window.matchMedia === "function"
        ? window.matchMedia("(display-mode: standalone)")
        : null;
    const onDisplayChange = () => {
      const nextStandalone = isStandaloneMode(window);
      setIsStandalone(nextStandalone);
      setIsIosInstallable(!nextStandalone && isIosDevice(window.navigator));
    };
    mq?.addEventListener?.("change", onDisplayChange);

    return () => {
      promptListeners.delete(syncPrompt);
      mq?.removeEventListener?.("change", onDisplayChange);
    };
  }, []);

  const promptInstall = useCallback(async () => {
    if (!deferredPrompt) return null;
    const promptEvent = deferredPrompt;
    await promptEvent.prompt();
    const choice = await promptEvent.userChoice;
    setDeferredInstallPrompt(null);
    return choice;
  }, []);

  const dismissBanner = useCallback(() => {
    persistBannerDismissed();
    setIsDismissed(true);
  }, []);

  return {
    isStandalone,
    canInstallNative: !isStandalone && hasNativePrompt,
    isIosInstallable,
    isDismissed,
    promptInstall,
    dismissBanner,
  };
}
