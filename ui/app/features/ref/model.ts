import {
  Building2,
  CalendarRange,
  Church,
  Compass,
  type LucideIcon,
  MapPin,
  Network,
} from "lucide-react";

export const ROLE_SUPER_ADMIN = "user:super-admin";
export const ROLE_PAROKI_ADMIN = "ref:paroki-admin";

export const LOOKUP_CATEGORIES = [
  "status_paroki",
  "agama",
  "golongan_darah",
  "suku",
  "hubungan_keluarga",
  "status_pernikahan",
  "status_rumah",
  "pendidikan",
  "pekerjaan",
  "profesi",
  "disabilitas",
  "status_ekonomi",
  "aktivitas_sosial",
  "org_position",
] as const;

export type LookupCategory = (typeof LOOKUP_CATEGORIES)[number];

export interface LookupCategoryMeta {
  slug: LookupCategory;
  label: string;
  description: string;
}

export const LOOKUP_CATEGORY_META: LookupCategoryMeta[] = [
  {
    slug: "status_paroki",
    label: "Status Paroki",
    description: "Klasifikasi status kanonik paroki dan stasi.",
  },
  {
    slug: "agama",
    label: "Agama",
    description: "Daftar agama untuk data umat dan keluarga.",
  },
  {
    slug: "golongan_darah",
    label: "Golongan Darah",
    description: "Golongan darah umat (A, B, AB, O, rhesus).",
  },
  {
    slug: "suku",
    label: "Suku / Etnis",
    description: "Latar belakang suku atau etnis umat.",
  },
  {
    slug: "hubungan_keluarga",
    label: "Hubungan Keluarga",
    description: "Peran anggota dalam kartu keluarga paroki.",
  },
  {
    slug: "status_pernikahan",
    label: "Status Pernikahan",
    description: "Status perkawinan sipil maupun gerejani.",
  },
  {
    slug: "status_rumah",
    label: "Status Tempat Tinggal",
    description: "Status kepemilikan atau hunian rumah tangga.",
  },
  {
    slug: "pendidikan",
    label: "Pendidikan",
    description: "Jenjang pendidikan formal terakhir.",
  },
  {
    slug: "pekerjaan",
    label: "Pekerjaan",
    description: "Bidang atau jenis pekerjaan utama.",
  },
  {
    slug: "profesi",
    label: "Profesi / Keahlian",
    description: "Keahlian khusus untuk pelayanan paroki.",
  },
  {
    slug: "disabilitas",
    label: "Disabilitas",
    description: "Kebutuhan khusus atau disabilitas umat.",
  },
  {
    slug: "status_ekonomi",
    label: "Status Ekonomi",
    description: "Kategori kondisi ekonomi keluarga.",
  },
  {
    slug: "aktivitas_sosial",
    label: "Aktivitas Sosial",
    description: "Keterlibatan sosial dan kemasyarakatan.",
  },
  {
    slug: "org_position",
    label: "Jabatan Organisasi",
    description: "Jabatan pengurus organisasi dan dewan paroki.",
  },
];

export function getLookupCategoryLabel(slug: string): string {
  return LOOKUP_CATEGORY_META.find((c) => c.slug === slug)?.label ?? slug;
}

export function isSuperAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes(ROLE_SUPER_ADMIN));
}

export function isParokiAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes(ROLE_PAROKI_ADMIN));
}

export function hasAnyRefAdminRole(roles?: string[]): boolean {
  return isSuperAdmin(roles) || isParokiAdmin(roles);
}

export function slugifyCode(raw: string): string {
  return raw
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

export interface HubStepDef {
  id: string;
  path: string;
  title: string;
  description: string;
  icon: LucideIcon;
  plate: 1 | 2 | 3 | 4 | 5 | 6;
}

export const GEREJAWI_STEPS: HubStepDef[] = [
  {
    id: "keuskupan",
    path: "/settings/gerejawi/keuskupan",
    title: "Keuskupan",
    description: "Kelola data keuskupan sebagai induk wilayah gerejawi.",
    icon: Building2,
    plate: 1,
  },
  {
    id: "kevikepan",
    path: "/settings/gerejawi/kevikepan",
    title: "Kevikepan",
    description: "Kelola kevikepan di bawah masing-masing keuskupan.",
    icon: Network,
    plate: 2,
  },
  {
    id: "paroki",
    path: "/settings/gerejawi/paroki",
    title: "Paroki",
    description: "Kelola tenant paroki dan tunjuk Admin Paroki.",
    icon: Church,
    plate: 3,
  },
];

export const PAROKI_STEPS: HubStepDef[] = [
  {
    id: "profile",
    path: "/settings/paroki/profile",
    title: "Profil Paroki",
    description: "Identitas paroki aktif dan daftar admin atau anggota paroki.",
    icon: Church,
    plate: 1,
  },
  {
    id: "wilayah",
    path: "/settings/paroki/wilayah",
    title: "Wilayah & Stasi",
    description: "Kelola wilayah dan stasi pada paroki aktif.",
    icon: Compass,
    plate: 2,
  },
  {
    id: "lingkungan",
    path: "/settings/paroki/lingkungan",
    title: "Lingkungan",
    description: "Kelola lingkungan di bawah setiap wilayah/stasi.",
    icon: MapPin,
    plate: 3,
  },
  {
    id: "periods",
    path: "/settings/paroki/periods",
    title: "Periode",
    description: "Periode tahun buku dan tahun liturgi paroki.",
    icon: CalendarRange,
    plate: 5,
  },
];
