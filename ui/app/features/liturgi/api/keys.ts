import type { SearchLiturgiSchedulesRequest } from "./types";

export const liturgiKeys = {
  all: ["liturgi"] as const,
  ministries: () => ["liturgi", "ministries"] as const,
  schedules: (params: Omit<SearchLiturgiSchedulesRequest, "cursor"> = {}) =>
    ["liturgi", "schedules", params] as const,
  schedule: (eventId: string) => ["liturgi", "schedule", eventId] as const,
  paroki: (parokiId: string) => ["liturgi", "paroki", parokiId] as const,
  orgs: () => ["liturgi", "orgs"] as const,
  parokiMembers: (parokiId: string, q = "") =>
    ["liturgi", "parokiMembers", parokiId, q] as const,
  songPartLookups: () => ["liturgi", "lookups", "liturgi_song_part"] as const,
  cantusSearch: (q = "") => ["liturgi", "cantusSearch", q] as const,
};
