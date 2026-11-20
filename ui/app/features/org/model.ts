import type { BadgeTone } from "~/shared/ui/Badge";

export const ROLE_PARISH_ADMIN = "ref:parish-admin";

export type OrgKind =
  | "dpp"
  | "bgkp"
  | "section"
  | "categorical"
  | "zone"
  | "ward"
  | "community";

export const ORG_KINDS: { value: OrgKind; label: string }[] = [
  { value: "dpp", label: "DPP" },
  { value: "bgkp", label: "BGKP" },
  { value: "section", label: "Bidang / Seksi" },
  { value: "categorical", label: "Kategorial" },
  { value: "zone", label: "Wilayah / Stasi" },
  { value: "ward", label: "Lingkungan" },
  { value: "community", label: "Komunitas" },
];

export const DEFAULT_POSITION_SUGGESTIONS = [
  "Ketua",
  "Wakil Ketua",
  "Sekretaris",
  "Bendahara",
  "Koordinator",
  "Anggota",
] as const;

export const INVITE_TTL_OPTIONS: { label: string; seconds: number }[] = [
  { label: "24 jam", seconds: 86400 },
  { label: "3 hari", seconds: 259200 },
  { label: "7 hari", seconds: 604800 },
];

export function isParishAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes(ROLE_PARISH_ADMIN));
}

export function slugifyCode(raw: string): string {
  return raw
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

export function kindLabel(kind: string): string {
  return ORG_KINDS.find((k) => k.value === kind)?.label ?? kind;
}

export function providerLabel(provider: string): string {
  switch (provider) {
    case "google":
      return "Google";
    case "github":
      return "GitHub";
    default:
      return provider;
  }
}

export function inviteStatusMeta(status: string): {
  label: string;
  tone: BadgeTone;
} {
  switch (status) {
    case "active":
      return { label: "Aktif", tone: "success" };
    case "expired":
      return { label: "Kedaluwarsa", tone: "warning" };
    case "revoked":
      return { label: "Dicabut", tone: "danger" };
    case "exhausted":
      return { label: "Kuota Habis", tone: "neutral" };
    default:
      return { label: status, tone: "neutral" };
  }
}

export function computeInviteStatus(invite: {
  revokedAt?: string | null;
  expiresAt: string;
  maxUses?: number | null;
  usedCount: number;
}): "active" | "expired" | "revoked" | "exhausted" {
  if (invite.revokedAt) return "revoked";
  if (new Date(invite.expiresAt).getTime() <= Date.now()) return "expired";
  if (invite.maxUses != null && invite.usedCount >= invite.maxUses) {
    return "exhausted";
  }
  return "active";
}

export function buildInviteJoinUrl(origin: string, code: string): string {
  const base = origin.replace(/\/+$/, "");
  return `${base}/org/join/${encodeURIComponent(code)}`;
}
