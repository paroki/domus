import type { RoleDescriptor, User } from "./api/types";

export const ROLE_SUPER_ADMIN = "user:super-admin";
export const ROLE_USER = "user:user";
export const ROLE_PARISH_ADMIN = "ref:parish-admin";
export const ROLE_ORG_ADMIN = "org:admin";
export const ROLE_ORG_MEMBER = "org:member";

export const ROLE_DESCRIPTORS: RoleDescriptor[] = [
  {
    code: ROLE_SUPER_ADMIN,
    name: "Super Admin",
    scope: "global",
    description:
      "Mengelola pengguna global, keuskupan, tenant paroki, dan kamus referensi global.",
  },
  {
    code: ROLE_USER,
    name: "Pengguna",
    scope: "global",
    description:
      "Pengguna terdaftar aplikasi yang dapat bergabung ke paroki dan organisasi.",
  },
  {
    code: ROLE_PARISH_ADMIN,
    name: "Admin Paroki",
    scope: "parish",
    description:
      "Mengelola profil paroki, wilayah & lingkungan, kamus referensi, periode, dan organisasi paroki.",
  },
  {
    code: ROLE_ORG_ADMIN,
    name: "Admin Organisasi",
    scope: "org",
    description:
      "Mengelola profil organisasi, pengurus & anggota, tautan undangan, dan kegiatan organisasi.",
  },
  {
    code: ROLE_ORG_MEMBER,
    name: "Anggota Organisasi",
    scope: "org",
    description:
      "Melihat direktori internal organisasi, kegiatan internal, dan konfirmasi kehadiran (RSVP).",
  },
];

export type UserDirectoryScope = "global" | "parish" | "org";

export function isSuperAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_SUPER_ADMIN));
}

export function isParishAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_PARISH_ADMIN));
}

export function isOrgAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_ORG_ADMIN));
}

export function getAvailableUserScopes(
  roles?: string[] | null,
): UserDirectoryScope[] {
  if (!roles?.length) return [];
  const scopes: UserDirectoryScope[] = [];
  if (roles.includes(ROLE_SUPER_ADMIN)) {
    scopes.push("global");
  }
  if (roles.includes(ROLE_SUPER_ADMIN) || roles.includes(ROLE_PARISH_ADMIN)) {
    scopes.push("parish");
  }
  if (roles.includes(ROLE_PARISH_ADMIN) || roles.includes(ROLE_ORG_ADMIN)) {
    scopes.push("org");
  }
  return scopes;
}

export function canManageUsers(user?: User | null): boolean {
  return getAvailableUserScopes(user?.roles).length > 0;
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
