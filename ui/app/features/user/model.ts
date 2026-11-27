import type {
  Organization,
  OrgMember,
  RoleDescriptor,
  User,
} from "./api/types";

export const ROLE_SUPER_ADMIN = "user:super-admin";
export const ROLE_CANTUS_ADMIN = "cantus:admin";
export const ROLE_USER = "user:user";
export const ROLE_PAROKI_ADMIN = "ref:paroki-admin";
export const ROLE_UMAT_ADMIN = "umat:admin";
export const ROLE_LITURGI_ADMIN = "liturgi:admin";
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
    code: ROLE_CANTUS_ADMIN,
    name: "Admin Nyanyian",
    scope: "global",
    description:
      "Mengelola inventaris lagu, mazmur, dan BPI global lintas-paroki.",
  },
  {
    code: ROLE_USER,
    name: "Pengguna",
    scope: "global",
    description:
      "Pengguna terdaftar aplikasi yang dapat bergabung ke paroki dan organisasi.",
  },
  {
    code: ROLE_PAROKI_ADMIN,
    name: "Admin Paroki",
    scope: "paroki",
    description:
      "Mengelola profil paroki, wilayah & lingkungan, kamus referensi, periode, dan organisasi paroki.",
  },
  {
    code: ROLE_UMAT_ADMIN,
    name: "Admin Data Umat",
    scope: "paroki",
    description:
      "Mengelola buku induk kartu keluarga, profil individu umat, dan riwayat sakramen paroki.",
  },
  {
    code: ROLE_LITURGI_ADMIN,
    name: "Admin Liturgi",
    scope: "paroki",
    description:
      "Mengelola jadwal perayaan Misa paroki, pemetaan 6 bidang petugas liturgi, penugasan petugas, dan daftar lagu perayaan.",
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

export type UserDirectoryScope = "global" | "org";
export type UserDetailTab = "profile" | "paroki" | "org";

export function getAvailableUserDetailTabs(
  roles?: string[] | null,
): UserDetailTab[] {
  if (!roles?.length) return [];
  const tabs: UserDetailTab[] = [];
  if (roles.includes(ROLE_SUPER_ADMIN)) {
    tabs.push("profile", "paroki", "org");
    return tabs;
  }
  if (roles.includes(ROLE_PAROKI_ADMIN)) {
    tabs.push("paroki", "org");
    return tabs;
  }
  if (roles.includes(ROLE_UMAT_ADMIN)) {
    tabs.push("paroki");
  }
  if (roles.includes(ROLE_ORG_ADMIN)) {
    tabs.push("org");
  }
  return tabs;
}

export function parseUserDetailTab(
  raw: string | null,
  allowed: UserDetailTab[] = ["profile", "paroki", "org"],
): UserDetailTab {
  if (
    (raw === "profile" || raw === "paroki" || raw === "org") &&
    allowed.includes(raw)
  ) {
    return raw;
  }
  return allowed[0] ?? "profile";
}

export type DirectoryOrgUser = {
  id: string;
  name: string;
  email: string;
  phone?: string;
  orgs: Array<{
    orgId: string;
    orgName: string;
    role: string;
    status: string;
  }>;
};

export function deduplicateOrgMembers(
  groups: Array<{ org: Organization; members: OrgMember[] }>,
): DirectoryOrgUser[] {
  const byId = new Map<string, DirectoryOrgUser>();
  for (const { org, members } of groups) {
    for (const m of members) {
      if (!m.userId) continue;
      const existing = byId.get(m.userId);
      const entry = {
        orgId: org.id ?? "",
        orgName: org.name ?? "",
        role: m.role ?? "member",
        status: m.status ?? "active",
      };
      if (existing) {
        existing.orgs.push(entry);
      } else {
        byId.set(m.userId, {
          id: m.userId,
          name: m.name ?? "",
          email: m.email ?? "",
          phone: m.phone,
          orgs: [entry],
        });
      }
    }
  }
  return Array.from(byId.values()).sort((a, b) =>
    (a.name || a.email).localeCompare(b.name || b.email, "id"),
  );
}

export function isSuperAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_SUPER_ADMIN));
}

export function isCantusAdmin(user?: User | null): boolean {
  return Boolean(
    user?.isCantusAdmin || user?.roles?.includes(ROLE_CANTUS_ADMIN),
  );
}

export function isParokiAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_PAROKI_ADMIN));
}

export function isUmatAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_UMAT_ADMIN));
}

export function isLiturgiAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_LITURGI_ADMIN));
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
  if (
    roles.includes(ROLE_PAROKI_ADMIN) ||
    roles.includes(ROLE_UMAT_ADMIN) ||
    roles.includes(ROLE_ORG_ADMIN)
  ) {
    scopes.push("org");
  }
  return scopes;
}

export function canManageUsers(user?: User | null): boolean {
  return getAvailableUserScopes(user?.roles).length > 0;
}

export function getInitials(
  name?: string | null,
  email?: string | null,
): string {
  const source = name?.trim() || email?.trim() || "";
  if (!source) return "?";
  const parts = source.split(/\s+/).filter(Boolean);
  if (parts.length >= 2) {
    return `${parts[0]?.[0] ?? ""}${parts[1]?.[0] ?? ""}`.toUpperCase();
  }
  return source.slice(0, 2).toUpperCase();
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

export const USER_GENDERS = ["pria", "wanita"] as const;
export type UserGender = (typeof USER_GENDERS)[number];

export const ME_ROUTES = {
  index: "/me",
  verify: "/me/verify",
  tokenNew: "/me/tokens/new",
  linkWhatsapp: "/me/accounts/whatsapp",
  linkTelegram: "/me/accounts/telegram",
} as const;

export function getBirthDateBounds(now: Date = new Date()): {
  min: string;
  max: string;
} {
  const min = new Date(
    Date.UTC(now.getUTCFullYear() - 120, now.getUTCMonth(), now.getUTCDate()),
  );
  const max = new Date(
    Date.UTC(now.getUTCFullYear() - 13, now.getUTCMonth(), now.getUTCDate()),
  );
  return {
    min: min.toISOString().slice(0, 10),
    max: max.toISOString().slice(0, 10),
  };
}
