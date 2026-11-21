import {
  BookOpen,
  Building2,
  CalendarRange,
  Church,
  type LucideIcon,
  MapPin,
} from "lucide-react";
import type { AppDef } from "~/shared/app-def";

export const ROLE_SUPER_ADMIN = "user:super-admin";
export const ROLE_PARISH_ADMIN = "ref:parish-admin";

export const LOOKUP_CATEGORIES = [
  "sacrament",
  "marriage_status",
  "education",
  "occupation",
  "relation",
  "expense_category",
  "income_category",
] as const;

export function isSuperAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes(ROLE_SUPER_ADMIN));
}

export function isParishAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes(ROLE_PARISH_ADMIN));
}

export function hasAnyRefAdminRole(roles?: string[]): boolean {
  return isSuperAdmin(roles) || isParishAdmin(roles);
}

export function slugifyCode(raw: string): string {
  return raw
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

export interface SettingsItem {
  id: "profile" | "territory" | "lookups" | "periods" | "parishes";
  path: string;
  label: string;
  description: string;
  icon: LucideIcon;
  plate: AppDef["plate"];
  roles: string[];
}

export const SETTINGS_ITEMS: SettingsItem[] = [
  {
    id: "profile",
    path: "/ref/profile",
    label: "Profil Paroki",
    description: "Identitas paroki aktif dan daftar admin atau anggota paroki.",
    icon: Church,
    plate: 1,
    roles: [ROLE_PARISH_ADMIN],
  },
  {
    id: "territory",
    path: "/ref/territory",
    label: "Wilayah & Lingkungan",
    description: "Kelola wilayah, stasi, dan lingkungan paroki.",
    icon: MapPin,
    plate: 2,
    roles: [ROLE_PARISH_ADMIN],
  },
  {
    id: "lookups",
    path: "/ref/lookups",
    label: "Kamus Referensi",
    description: "Daftar nilai referensi global maupun khusus paroki.",
    icon: BookOpen,
    plate: 3,
    roles: [ROLE_PARISH_ADMIN, ROLE_SUPER_ADMIN],
  },
  {
    id: "periods",
    path: "/ref/periods",
    label: "Periode",
    description: "Periode tahun buku dan tahun liturgi paroki.",
    icon: CalendarRange,
    plate: 5,
    roles: [ROLE_PARISH_ADMIN],
  },
  {
    id: "parishes",
    path: "/ref/parishes",
    label: "Keuskupan & Paroki",
    description: "Kelola keuskupan, tenant paroki, dan admin paroki.",
    icon: Building2,
    plate: 6,
    roles: [ROLE_SUPER_ADMIN],
  },
];

export function getVisibleSettingsItems(
  roles: string[] = [],
  q = "",
): SettingsItem[] {
  const needle = q.trim().toLowerCase();
  return SETTINGS_ITEMS.filter((item) => {
    if (!item.roles.some((r) => roles.includes(r))) {
      return false;
    }
    if (!needle) return true;
    return (
      item.label.toLowerCase().includes(needle) ||
      item.description.toLowerCase().includes(needle)
    );
  });
}
