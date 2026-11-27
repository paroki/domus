import type { User } from "./api/types";

export const ROLE_UMAT_ADMIN = "umat:admin";
export const ROLE_SUPER_ADMIN = "user:super-admin";
export const NIL_UUID = "00000000-0000-0000-0000-000000000000";

export function isUmatAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_UMAT_ADMIN));
}

export function isSuperAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_SUPER_ADMIN));
}

export function canEditUmatReferensi(
  user?: User | null,
  parokiId?: string,
): boolean {
  if (parokiId) {
    return isUmatAdmin(user);
  }
  return isSuperAdmin(user);
}

export interface UmatReferensiCategoryMeta {
  slug: string;
  label: string;
  description: string;
}

export const UMAT_REFERENSI_CATEGORIES: readonly UmatReferensiCategoryMeta[] = [
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
    slug: "status_ekonomi",
    label: "Status Ekonomi",
    description: "Kategori kondisi ekonomi keluarga.",
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
    slug: "aktivitas_sosial",
    label: "Aktivitas Sosial",
    description: "Keterlibatan sosial dan kemasyarakatan.",
  },
] as const;

export function getUmatReferensiCategoryLabel(slug: string): string {
  return UMAT_REFERENSI_CATEGORIES.find((c) => c.slug === slug)?.label ?? slug;
}

export function slugifyReferensiCode(raw: string): string {
  return raw
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

export function optionalUuidForPatch(raw: string): string {
  const trimmed = raw.trim();
  return trimmed === "" ? NIL_UUID : trimmed;
}

export function optionalStringForPatch(raw: string): string {
  return raw.trim();
}

export function genderLabel(jenisKelamin?: string | null): string {
  if (jenisKelamin === "pria") return "Pria";
  if (jenisKelamin === "wanita") return "Wanita";
  return jenisKelamin ?? "-";
}

export const SAKRAMEN_JENIS_OPTIONS = [
  { value: "baptis", label: "Baptis" },
  { value: "komuni_pertama", label: "Komuni Pertama" },
  { value: "krisma", label: "Krisma" },
  { value: "perkawinan", label: "Perkawinan" },
  { value: "imamat", label: "Imamat" },
  { value: "perminyakan", label: "Perminyakan Orang Sakit" },
  { value: "kematian", label: "Kematian" },
] as const;

export type SakramenJenis = (typeof SAKRAMEN_JENIS_OPTIONS)[number]["value"];

export function sakramenJenisLabel(jenis?: string | null): string {
  const found = SAKRAMEN_JENIS_OPTIONS.find((o) => o.value === jenis);
  return found?.label ?? jenis ?? "-";
}

export function validateYYYYMMDD(raw: string, label: string): string | null {
  const trimmed = raw.trim();
  if (trimmed === "") return null;
  if (!/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) {
    return `${label} harus berformat YYYY-MM-DD.`;
  }
  const d = new Date(`${trimmed}T00:00:00Z`);
  if (Number.isNaN(d.getTime()) || d.toISOString().slice(0, 10) !== trimmed) {
    return `${label} bukan tanggal yang valid.`;
  }
  return null;
}

export function validatePIIInput(raw: string, label: string): string | null {
  const trimmed = raw.trim();
  if (trimmed === "") return null;
  const len = Array.from(trimmed).length;
  if (len < 4 || len > 32) {
    return `${label} harus terdiri dari 4 sampai 32 karakter.`;
  }
  for (const ch of trimmed) {
    const code = ch.charCodeAt(0);
    if (code < 0x20 || code === 0x7f) {
      return `${label} tidak boleh mengandung karakter kontrol.`;
    }
  }
  return null;
}

export type UmatAlmarhumFilter = "all" | "alive" | "deceased";
export type KeluargaAktifFilter = "all" | "true" | "false";

export function parseUmatDirectoryFilters(params: URLSearchParams) {
  const almarhumRaw = params.get("almarhum");
  const almarhum: UmatAlmarhumFilter =
    almarhumRaw === "alive" || almarhumRaw === "deceased" ? almarhumRaw : "all";
  const jkRaw = params.get("jenisKelamin");
  const jenisKelamin: "pria" | "wanita" | "" =
    jkRaw === "pria" || jkRaw === "wanita" ? jkRaw : "";
  return {
    q: params.get("q") ?? "",
    nik: params.get("nik") ?? "",
    wilayahId: params.get("wilayahId") ?? "",
    lingkunganId: params.get("lingkunganId") ?? "",
    jenisKelamin,
    agama: params.get("agama") ?? "",
    statusPernikahan: params.get("statusPernikahan") ?? "",
    almarhum,
  };
}

export function parseKeluargaDirectoryFilters(params: URLSearchParams) {
  const aktifRaw = params.get("aktif");
  const aktif: KeluargaAktifFilter =
    aktifRaw === "true" || aktifRaw === "false" ? aktifRaw : "all";
  return {
    q: params.get("q") ?? "",
    nomorKk: params.get("nomorKk") ?? "",
    wilayahId: params.get("wilayahId") ?? "",
    lingkunganId: params.get("lingkunganId") ?? "",
    aktif,
  };
}

export function countActiveAdvancedUmatFilters(
  filters: ReturnType<typeof parseUmatDirectoryFilters>,
): number {
  let count = 0;
  if (filters.nik.trim()) count += 1;
  if (filters.jenisKelamin) count += 1;
  if (filters.agama) count += 1;
  if (filters.statusPernikahan) count += 1;
  if (filters.almarhum !== "all") count += 1;
  return count;
}
