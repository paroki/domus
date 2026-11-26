import type { User } from "./api/types";

export const ROLE_UMAT_ADMIN = "umat:admin";
export const NIL_UUID = "00000000-0000-0000-0000-000000000000";

export function isUmatAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.includes(ROLE_UMAT_ADMIN));
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
