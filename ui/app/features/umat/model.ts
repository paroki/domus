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
