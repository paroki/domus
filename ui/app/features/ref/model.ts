import {
  Building2,
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
  "event_type",
  "liturgi_song_part",
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
  {
    slug: "event_type",
    label: "Jenis Kegiatan",
    description:
      "Klasifikasi jenis kegiatan paroki dan organisasi (Misa, Ibadat, Rapat, Latihan, dll.).",
  },
  {
    slug: "liturgi_song_part",
    label: "Bagian Lagu Liturgi",
    description:
      "Bagian nyanyian dalam tata perayaan liturgi Misa (Pembuka, Kyrie, Gloria, Mazmur, BPI, Persembahan, Sanctus, Agnus Dei, Komuni, Madah Syukur, Penutup).",
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
];

export const SIPIL_STEPS: HubStepDef[] = [
  {
    id: "kota",
    path: "/settings/sipil/kota",
    title: "Kota / Kabupaten",
    description: "Data referensi kota dan kabupaten tingkat II.",
    icon: Building2,
    plate: 1,
  },
  {
    id: "kecamatan",
    path: "/settings/sipil/kecamatan",
    title: "Kecamatan",
    description: "Data referensi kecamatan di bawah kota/kabupaten.",
    icon: MapPin,
    plate: 2,
  },
  {
    id: "kelurahan",
    path: "/settings/sipil/kelurahan",
    title: "Kelurahan / Desa",
    description: "Data referensi kelurahan atau desa beserta kode pos.",
    icon: Compass,
    plate: 3,
  },
];

export const NIL_UUID = "00000000-0000-0000-0000-000000000000";
export const DEFAULT_PAROKI_TIMEZONE = "Asia/Makassar";

export const PAROKI_TIMEZONES = [
  { value: "Asia/Jakarta", label: "Asia/Jakarta (WIB)", short: "WIB" },
  { value: "Asia/Makassar", label: "Asia/Makassar (WITA)", short: "WITA" },
  { value: "Asia/Jayapura", label: "Asia/Jayapura (WIT)", short: "WIT" },
] as const;

export function formatParokiTimezone(tz?: string): string {
  if (!tz) return "WITA";
  return PAROKI_TIMEZONES.find((item) => item.value === tz)?.short ?? tz;
}

export function validateCoordinatePair(
  latStr: string,
  lngStr: string,
): {
  valid: boolean;
  latitude?: number;
  longitude?: number;
  error?: string;
} {
  const latTrim = latStr.trim();
  const lngTrim = lngStr.trim();
  if (!latTrim && !lngTrim) {
    return { valid: true };
  }
  if (!latTrim || !lngTrim) {
    return {
      valid: false,
      error: "Latitude dan longitude harus diisi berpasangan.",
    };
  }
  const latitude = Number(latTrim);
  const longitude = Number(lngTrim);
  if (!Number.isFinite(latitude) || latitude < -90 || latitude > 90) {
    return { valid: false, error: "Latitude harus antara -90 dan 90." };
  }
  if (!Number.isFinite(longitude) || longitude < -180 || longitude > 180) {
    return { valid: false, error: "Longitude harus antara -180 dan 180." };
  }
  return { valid: true, latitude, longitude };
}
