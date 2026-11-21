import type { User } from "./api/types";

export const ROLE_SUPER_ADMIN = "user:super-admin";

export function isSuperAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_SUPER_ADMIN));
}

export function formatUserAgent(ua: string): string {
  if (!ua) return "Perangkat tidak dikenal";

  let browser = "Browser";
  if (/Edg\//i.test(ua)) browser = "Edge";
  else if (/Firefox\//i.test(ua)) browser = "Firefox";
  else if (/Chrome\//i.test(ua) && !/Edg\//i.test(ua)) browser = "Chrome";
  else if (/Safari\//i.test(ua) && !/Chrome\//i.test(ua)) browser = "Safari";
  else if (/curl/i.test(ua)) browser = "curl";

  let os = "Lainnya";
  if (/Android/i.test(ua)) os = "Android";
  else if (/iPhone|iPad|iPod/i.test(ua)) os = "iOS";
  else if (/Windows/i.test(ua)) os = "Windows";
  else if (/Mac OS X|Macintosh/i.test(ua)) os = "macOS";
  else if (/Linux/i.test(ua)) os = "Linux";

  return `${browser} · ${os}`;
}

export function providerLabel(provider: string): string {
  switch (provider) {
    case "email":
      return "Email OTP";
    case "google":
      return "Google";
    case "github":
      return "GitHub";
    case "telegram":
      return "Telegram";
    case "whatsapp":
      return "WhatsApp";
    default:
      return provider;
  }
}

export function sanitizeRedirect(raw: string | null): string {
  if (!raw?.startsWith("/") || raw.startsWith("//")) {
    return "/";
  }
  return raw;
}

export type OnboardingStep = "biodata" | "whatsapp" | "telegram";

export function buildOnboardingSteps(otpChannels: string[]): OnboardingStep[] {
  const steps: OnboardingStep[] = ["biodata"];
  if (otpChannels.includes("whatsapp")) {
    steps.push("whatsapp");
  }
  if (otpChannels.includes("telegram")) {
    steps.push("telegram");
  }
  return steps;
}

export function shouldRedirectToOnboarding(
  onboarded: boolean | undefined,
  pathname: string,
): boolean {
  return onboarded === false && pathname !== "/onboarding";
}
