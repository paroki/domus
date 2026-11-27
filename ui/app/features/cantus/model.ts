export const ROLE_CANTUS_ADMIN = "cantus:admin";

export const CANTUS_JENIS_OPTIONS = [
  { value: "lagu", label: "Lagu Liturgi" },
  { value: "mazmur", label: "Mazmur & BPI" },
  { value: "profan", label: "Lagu Non-Resmi (Profan)" },
] as const;

export const CANTUS_SUMBER_SUGGESTIONS = [
  "PS",
  "MB",
  "KA",
  "Mazmur",
  "Lainnya",
] as const;

export const CANTUS_KATEGORI_SUGGESTIONS = [
  "Pembuka",
  "Tuhan Kasihanilah",
  "Kemuliaan",
  "Mazmur Tanggapan",
  "Bait Pengantar Injil",
  "Persembahan",
  "Kudus",
  "Anak Domba Allah",
  "Komuni",
  "Madah Syukur",
  "Penutup",
  "Adven",
  "Natal",
  "Prapaskah",
  "Paskah",
  "Maria",
  "Umum",
] as const;

export const CANTUS_SIKLUS_OPTIONS = ["A", "B", "C", "I", "II"] as const;

export const CANTUS_MEDIA_TYPES = ["image", "pdf", "audio", "video"] as const;

export type CantusActiveFilter = "active" | "all" | "inactive";

export interface CantusDirectoryFilters {
  q: string;
  jenis: string;
  sumber: string;
  kategori: string;
  tag: string;
  nadaDasar: string;
  active: CantusActiveFilter;
}

export function isCantusAdmin(
  me?: { isCantusAdmin?: boolean; roles?: string[] | null } | null,
): boolean {
  return Boolean(me?.isCantusAdmin || me?.roles?.includes(ROLE_CANTUS_ADMIN));
}

export function cantusJenisLabel(jenis?: string | null): string {
  switch (jenis) {
    case "lagu":
      return "Lagu Liturgi";
    case "mazmur":
      return "Mazmur & BPI";
    case "profan":
      return "Profan";
    default:
      return jenis || "-";
  }
}

export function slugifyCantus(
  sumber?: string | null,
  nomor?: string | null,
  judul?: string | null,
): string {
  const prefix = (nomor?.trim() || sumber?.trim() || "").trim();
  const title = (judul?.trim() || "").trim();
  const raw = [prefix, title].filter(Boolean).join(" ").toLowerCase();
  const normalized = raw
    .replace(/[^a-z0-9_-]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .replace(/-{2,}/g, "-");
  if (!normalized) return "";
  return normalized.slice(0, 64).replace(/-+$/g, "");
}

export interface MazmurMetaForm {
  perayaan?: string;
  siklusTahun?: string;
  masa?: string;
  refAlkitab?: string;
  ulangan?: string;
  bait?: string;
  bpi?: string;
}

export function parseMazmurMeta(
  meta?: Record<string, unknown> | null,
): MazmurMetaForm {
  if (!meta) return {};
  return {
    perayaan: typeof meta.perayaan === "string" ? meta.perayaan : undefined,
    siklusTahun:
      typeof meta.siklus_tahun === "string" ? meta.siklus_tahun : undefined,
    masa: typeof meta.masa === "string" ? meta.masa : undefined,
    refAlkitab:
      typeof meta.ref_alkitab === "string" ? meta.ref_alkitab : undefined,
    ulangan: typeof meta.ulangan === "string" ? meta.ulangan : undefined,
    bait: typeof meta.bait === "string" ? meta.bait : undefined,
    bpi: typeof meta.bpi === "string" ? meta.bpi : undefined,
  };
}

export function buildMazmurMeta(
  form: MazmurMetaForm,
  existingMeta?: Record<string, unknown> | null,
): Record<string, unknown> {
  const next: Record<string, unknown> = { ...(existingMeta ?? {}) };
  const setOrDelete = (key: string, val?: string) => {
    const trimmed = val?.trim();
    if (trimmed) next[key] = trimmed;
    else delete next[key];
  };
  setOrDelete("perayaan", form.perayaan);
  setOrDelete("siklus_tahun", form.siklusTahun);
  setOrDelete("masa", form.masa);
  setOrDelete("ref_alkitab", form.refAlkitab);
  setOrDelete("ulangan", form.ulangan);
  setOrDelete("bait", form.bait);
  setOrDelete("bpi", form.bpi);
  return next;
}

export interface LaguMetaForm {
  birama?: string;
  tempo?: string;
  notAwal?: string;
}

export function parseLaguMeta(
  meta?: Record<string, unknown> | null,
): LaguMetaForm {
  if (!meta) return {};
  return {
    birama: typeof meta.birama === "string" ? meta.birama : undefined,
    tempo: typeof meta.tempo === "string" ? meta.tempo : undefined,
    notAwal: typeof meta.not_awal === "string" ? meta.not_awal : undefined,
  };
}

export function buildLaguMeta(
  form: LaguMetaForm,
  existingMeta?: Record<string, unknown> | null,
): Record<string, unknown> {
  const next: Record<string, unknown> = { ...(existingMeta ?? {}) };
  const setOrDelete = (key: string, val?: string) => {
    const trimmed = val?.trim();
    if (trimmed) next[key] = trimmed;
    else delete next[key];
  };
  setOrDelete("birama", form.birama);
  setOrDelete("tempo", form.tempo);
  setOrDelete("not_awal", form.notAwal);
  return next;
}

export function parseCantusDirectoryFilters(
  params: URLSearchParams,
): CantusDirectoryFilters {
  const rawActive = params.get("active");
  const active: CantusActiveFilter =
    rawActive === "all" || rawActive === "false"
      ? rawActive === "false"
        ? "inactive"
        : "all"
      : "active";
  return {
    q: params.get("q") ?? "",
    jenis: params.get("jenis") ?? "",
    sumber: params.get("sumber") ?? "",
    kategori: params.get("kategori") ?? "",
    tag: params.get("tag") ?? "",
    nadaDasar: params.get("nadaDasar") ?? "",
    active,
  };
}
