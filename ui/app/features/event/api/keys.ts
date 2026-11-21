import type { SearchEventRSVPsRequest, SearchEventsRequest } from "./types";

export const eventKeys = {
  all: ["event"] as const,
  events: (params?: Omit<SearchEventsRequest, "cursor">) =>
    [...eventKeys.all, "list", params ?? {}] as const,
  event: (eventId: string) => [...eventKeys.all, "detail", eventId] as const,
  rsvpSummary: (eventId: string) =>
    [...eventKeys.all, "rsvp-summary", eventId] as const,
  rsvps: (eventId: string, params?: Omit<SearchEventRSVPsRequest, "cursor">) =>
    [...eventKeys.all, "rsvps", eventId, params ?? {}] as const,
  orgs: () => [...eventKeys.all, "orgs"] as const,
};
