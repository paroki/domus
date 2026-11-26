import type { BadgeTone } from "~/shared/ui/Badge";

export type EventLifecycle = "upcoming" | "ongoing" | "completed" | "cancelled";
export type EventVisibility = "public" | "private";
export type EventStatus = "scheduled" | "cancelled";
export type RSVPStatus = "going" | "maybe" | "not_going";
export type PetugasStatus = "assigned" | "confirmed" | "declined";

export const DEFAULT_EVENT_ROLES: readonly string[] = [
  "Selebran",
  "Organis",
  "Dirigen",
  "Kor",
  "Pemazmur",
  "Lektor 1",
  "Lektor 2",
  "Prodiakon",
  "Misdinar",
  "Kolektan",
  "Tata Tertib",
  "Kebersihan Gereja",
  "MC",
  "Notulis",
];

export const DEFAULT_EVENT_SONG_PARTS: readonly string[] = [
  "Pembuka",
  "Tuhan Kasihanilah (Kyrie)",
  "Kemuliaan (Gloria)",
  "Mazmur Tanggapan",
  "Bait Pengantar Injil",
  "Persembahan",
  "Kudus (Sanctus)",
  "Anak Domba Allah (Agnus Dei)",
  "Komuni",
  "Madah Syukur",
  "Penutup",
];

export function petugasStatusMeta(status: string): {
  label: string;
  tone: BadgeTone;
} {
  switch (status) {
    case "assigned":
      return { label: "Ditugaskan", tone: "neutral" };
    case "confirmed":
      return { label: "Siap Bertugas", tone: "success" };
    case "declined":
      return { label: "Berhalangan", tone: "danger" };
    default:
      return { label: status, tone: "neutral" };
  }
}

export interface EventTypeOption {
  code: string;
  label: string;
}

export const DEFAULT_EVENT_TYPES: readonly EventTypeOption[] = [
  { code: "misa", label: "Misa / Ekaristi" },
  { code: "ibadat", label: "Ibadat" },
  { code: "rapat", label: "Rapat" },
  { code: "latihan", label: "Latihan" },
  { code: "rekoleksi", label: "Rekoleksi / Retret" },
  { code: "katekese", label: "Katekese / Pembinaan" },
  { code: "bakti-sosial", label: "Bakti Sosial" },
  { code: "lainnya", label: "Lainnya" },
];

export function mergeEventTypeOptions(
  lookups?: Array<{ code?: string; label?: string }>,
): EventTypeOption[] {
  const byCode = new Map<string, string>();
  for (const item of DEFAULT_EVENT_TYPES) {
    byCode.set(item.code, item.label);
  }
  for (const row of lookups ?? []) {
    const code = row.code?.trim().toLowerCase();
    const label = row.label?.trim();
    if (code && label) {
      byCode.set(code, label);
    }
  }
  return Array.from(byCode, ([code, label]) => ({ code, label }));
}

export function eventTypeLabel(
  code?: string,
  lookups?: Array<{ code?: string; label?: string }>,
): string {
  if (!code) return "";
  const normalized = code.trim().toLowerCase();
  const matched = (lookups ?? []).find(
    (row) => row.code?.trim().toLowerCase() === normalized && row.label?.trim(),
  );
  if (matched?.label) return matched.label.trim();
  const def = DEFAULT_EVENT_TYPES.find((item) => item.code === normalized);
  return def?.label ?? code;
}

export function computeEventLifecycle(
  event: { status?: string; startAt?: string; endAt?: string },
  nowMs = Date.now(),
): EventLifecycle {
  if (event.status === "cancelled") {
    return "cancelled";
  }
  const startMs = event.startAt ? new Date(event.startAt).getTime() : NaN;
  const endMs = event.endAt ? new Date(event.endAt).getTime() : NaN;
  if (!Number.isNaN(endMs) && nowMs > endMs) {
    return "completed";
  }
  if (
    !Number.isNaN(startMs) &&
    !Number.isNaN(endMs) &&
    nowMs >= startMs &&
    nowMs <= endMs
  ) {
    return "ongoing";
  }
  return "upcoming";
}

export function lifecycleMeta(lifecycle: EventLifecycle): {
  label: string;
  tone: BadgeTone;
} {
  switch (lifecycle) {
    case "upcoming":
      return { label: "Akan Datang", tone: "neutral" };
    case "ongoing":
      return { label: "Berlangsung", tone: "success" };
    case "completed":
      return { label: "Selesai", tone: "neutral" };
    case "cancelled":
      return { label: "Dibatalkan", tone: "danger" };
  }
}

export function visibilityMeta(visibility: string): {
  label: string;
  tone: BadgeTone;
} {
  switch (visibility) {
    case "public":
      return { label: "Publik Paroki", tone: "success" };
    case "private":
      return { label: "Internal Organisasi", tone: "warning" };
    default:
      return { label: visibility, tone: "neutral" };
  }
}

export function rsvpStatusMeta(status: string): {
  label: string;
  tone: BadgeTone;
} {
  switch (status) {
    case "going":
      return { label: "Hadir", tone: "success" };
    case "maybe":
      return { label: "Mungkin", tone: "warning" };
    case "not_going":
      return { label: "Tidak Hadir", tone: "danger" };
    default:
      return { label: status, tone: "neutral" };
  }
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

export function formatEventRange(
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
