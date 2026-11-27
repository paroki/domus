import type { BadgeTone } from "~/shared/ui/Badge";

export const ROLE_PAROKI_ADMIN = "ref:paroki-admin";
export const ROLE_LITURGI_ADMIN = "liturgi:admin";

export function canManageLiturgiGlobal(roles?: string[] | null): boolean {
  if (!roles?.length) return false;
  return (
    roles.includes(ROLE_LITURGI_ADMIN) || roles.includes(ROLE_PAROKI_ADMIN)
  );
}

export const LITURGI_ROLE_SLUGS = [
  "lektor",
  "mazmur",
  "pembagi-komuni",
  "misdinar",
  "organis",
  "kor",
] as const;

export type LiturgiRoleSlug = (typeof LITURGI_ROLE_SLUGS)[number];

export interface LiturgiMinistryMeta {
  slug: LiturgiRoleSlug;
  label: string;
  shortLabel: string;
  description: string;
}

export const LITURGI_MINISTRIES: readonly LiturgiMinistryMeta[] = [
  {
    slug: "lektor",
    label: "Lektor",
    shortLabel: "Lektor",
    description: "Pembaca sabda dan doa umat pada perayaan Ekaristi.",
  },
  {
    slug: "mazmur",
    label: "Pemazmur",
    shortLabel: "Mazmur",
    description: "Penyanyi mazmur tanggapan dan bait pengantar Injil.",
  },
  {
    slug: "pembagi-komuni",
    label: "Pembagi Komuni",
    shortLabel: "Komuni",
    description: "Prodiakon dan petugas pembagi komuni kudus.",
  },
  {
    slug: "misdinar",
    label: "Misdinar",
    shortLabel: "Misdinar",
    description: "Putra-putri altar yang melayani di sekitar altar.",
  },
  {
    slug: "organis",
    label: "Organis",
    shortLabel: "Organis",
    description: "Pengiring musik liturgi dan organis gereja.",
  },
  {
    slug: "kor",
    label: "Kor",
    shortLabel: "Kor",
    description:
      "Kelompok paduan suara atau dirigen yang memimpin nyanyian umat.",
  },
];

export function getLiturgiMinistryMeta(
  slug?: string,
): LiturgiMinistryMeta | undefined {
  return LITURGI_MINISTRIES.find((m) => m.slug === slug);
}

export const DEFAULT_LITURGI_SONG_PARTS = [
  "Pembuka",
  "Kyrie",
  "Gloria",
  "Mazmur",
  "BPI",
  "Persembahan",
  "Sanctus",
  "Agnus Dei",
  "Komuni",
  "Madah Syukur",
  "Penutup",
] as const;

export function petugasStatusMeta(status?: string): {
  label: string;
  tone: BadgeTone;
} {
  switch (status) {
    case "confirmed":
      return { label: "Siap", tone: "success" };
    case "declined":
      return { label: "Berhalangan", tone: "danger" };
    case "assigned":
      return { label: "Ditugaskan", tone: "neutral" };
    default:
      return { label: status || "Ditugaskan", tone: "neutral" };
  }
}

export function isUserAssignedToSchedule(
  petugas: Array<{ userId?: string }> | undefined,
  userId?: string,
): boolean {
  if (!userId || !petugas?.length) return false;
  return petugas.some((p) => p.userId === userId);
}

export function groupPetugasByRole<T extends { role?: string }>(
  petugas: readonly T[] | undefined,
): Record<LiturgiRoleSlug, T[]> {
  const grouped: Record<LiturgiRoleSlug, T[]> = {
    lektor: [],
    mazmur: [],
    "pembagi-komuni": [],
    misdinar: [],
    organis: [],
    kor: [],
  };
  for (const item of petugas ?? []) {
    const slug = item.role as LiturgiRoleSlug | undefined;
    if (slug && slug in grouped) {
      grouped[slug].push(item);
    }
  }
  return grouped;
}

export const DEFAULT_PAROKI_TIMEZONE = "Asia/Makassar";

export function formatParokiTimezoneShort(tz?: string): string {
  switch (tz) {
    case "Asia/Jakarta":
      return "WIB";
    case "Asia/Makassar":
      return "WITA";
    case "Asia/Jayapura":
      return "WIT";
    default:
      return tz || "WITA";
  }
}

function getTimeZoneParts(date: Date, timeZone: string) {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  }).formatToParts(date);
  const get = (type: Intl.DateTimeFormatPartTypes) =>
    parts.find((p) => p.type === type)?.value ?? "00";
  const hour = get("hour") === "24" ? "00" : get("hour");
  return {
    year: get("year"),
    month: get("month"),
    day: get("day"),
    hour,
    minute: get("minute"),
    second: get("second"),
  };
}

export function toLocalDatetimeInputValue(
  iso?: string,
  timeZone = DEFAULT_PAROKI_TIMEZONE,
): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const p = getTimeZoneParts(d, timeZone);
  return `${p.year}-${p.month}-${p.day}T${p.hour}:${p.minute}`;
}

export function fromLocalDatetimeInputValue(
  localVal: string,
  timeZone = DEFAULT_PAROKI_TIMEZONE,
): string {
  if (!localVal) return "";
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?$/.exec(
    localVal,
  );
  if (!match) {
    const d = new Date(localVal);
    return Number.isNaN(d.getTime()) ? "" : d.toISOString();
  }
  const [, y, m, day, h, min, sec = "00"] = match;
  const utcGuess = Date.UTC(
    Number(y),
    Number(m) - 1,
    Number(day),
    Number(h),
    Number(min),
    Number(sec),
  );
  const p = getTimeZoneParts(new Date(utcGuess), timeZone);
  const tzAsUtc = Date.UTC(
    Number(p.year),
    Number(p.month) - 1,
    Number(p.day),
    Number(p.hour),
    Number(p.minute),
    Number(p.second),
  );
  const actualUtc = utcGuess - (tzAsUtc - utcGuess);
  return new Date(actualUtc).toISOString();
}

export function formatLiturgiRange(
  startAt?: string,
  endAt?: string,
  timeZone = DEFAULT_PAROKI_TIMEZONE,
): string {
  if (!startAt) return "";
  const start = new Date(startAt);
  if (Number.isNaN(start.getTime())) return "";
  const tzShort = formatParokiTimezoneShort(timeZone);
  const dateFmt = new Intl.DateTimeFormat("id-ID", {
    weekday: "short",
    day: "numeric",
    month: "short",
    year: "numeric",
    timeZone,
  });
  const timeFmt = new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    timeZone,
  });
  if (!endAt) {
    return `${dateFmt.format(start)}, ${timeFmt.format(start)} ${tzShort}`;
  }
  const end = new Date(endAt);
  if (Number.isNaN(end.getTime())) {
    return `${dateFmt.format(start)}, ${timeFmt.format(start)} ${tzShort}`;
  }
  const sp = getTimeZoneParts(start, timeZone);
  const ep = getTimeZoneParts(end, timeZone);
  const sameDay =
    sp.year === ep.year && sp.month === ep.month && sp.day === ep.day;
  if (sameDay) {
    return `${dateFmt.format(start)} · ${timeFmt.format(start)} – ${timeFmt.format(end)} ${tzShort}`;
  }
  return `${dateFmt.format(start)}, ${timeFmt.format(start)} – ${dateFmt.format(end)}, ${timeFmt.format(end)} ${tzShort}`;
}
