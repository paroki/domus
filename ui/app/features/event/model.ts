import type { BadgeTone } from "~/shared/ui/Badge";

export type EventLifecycle = "upcoming" | "ongoing" | "completed" | "cancelled";
export type EventVisibility = "public" | "private";
export type EventStatus = "scheduled" | "cancelled";
export type RSVPStatus = "going" | "maybe" | "not_going";

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

export function toLocalDatetimeInputValue(iso?: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function fromLocalDatetimeInputValue(localVal: string): string {
  if (!localVal) return "";
  const d = new Date(localVal);
  if (Number.isNaN(d.getTime())) return "";
  return d.toISOString();
}

export function formatEventRange(startAt?: string, endAt?: string): string {
  if (!startAt) return "";
  const start = new Date(startAt);
  if (Number.isNaN(start.getTime())) return "";
  const dateFmt = new Intl.DateTimeFormat("id-ID", {
    weekday: "short",
    day: "numeric",
    month: "short",
    year: "numeric",
  });
  const timeFmt = new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
  });
  if (!endAt) {
    return `${dateFmt.format(start)}, ${timeFmt.format(start)}`;
  }
  const end = new Date(endAt);
  if (Number.isNaN(end.getTime())) {
    return `${dateFmt.format(start)}, ${timeFmt.format(start)}`;
  }
  const sameDay =
    start.getFullYear() === end.getFullYear() &&
    start.getMonth() === end.getMonth() &&
    start.getDate() === end.getDate();
  if (sameDay) {
    return `${dateFmt.format(start)} · ${timeFmt.format(start)} – ${timeFmt.format(end)}`;
  }
  return `${dateFmt.format(start)}, ${timeFmt.format(start)} – ${dateFmt.format(end)}, ${timeFmt.format(end)}`;
}
