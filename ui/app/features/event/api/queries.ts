import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useQuery,
} from "@tanstack/react-query";
import { api, unwrap, unwrapMeta } from "~/shared/api/client";
import { eventKeys } from "./keys";
import type { SearchEventRSVPsRequest, SearchEventsRequest } from "./types";

export function eventsInfiniteOptions(
  params: Omit<SearchEventsRequest, "cursor"> = {},
  enabled = true,
) {
  return infiniteQueryOptions({
    queryKey: eventKeys.events(params),
    initialPageParam: undefined as string | undefined,
    enabled,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/events/search", {
          body: { ...params, cursor: pageParam, limit: params.limit ?? 20 },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useEvents(
  params: Omit<SearchEventsRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(eventsInfiniteOptions(params, enabled));
}

export function eventDetailOptions(eventId: string, enabled = true) {
  return queryOptions({
    queryKey: eventKeys.event(eventId),
    enabled: Boolean(eventId) && enabled,
    queryFn: () =>
      unwrap(api.GET("/events/{id}", { params: { path: { id: eventId } } })),
  });
}

export function useEvent(eventId: string, enabled = true) {
  return useQuery(eventDetailOptions(eventId, enabled));
}

export function eventRSVPSummaryOptions(eventId: string, enabled = true) {
  return queryOptions({
    queryKey: eventKeys.rsvpSummary(eventId),
    enabled: Boolean(eventId) && enabled,
    queryFn: () =>
      unwrap(
        api.GET("/events/{id}/rsvps/summary", {
          params: { path: { id: eventId } },
        }),
      ),
  });
}

export function useEventRSVPSummary(eventId: string, enabled = true) {
  return useQuery(eventRSVPSummaryOptions(eventId, enabled));
}

export function eventRSVPsInfiniteOptions(
  eventId: string,
  params: Omit<SearchEventRSVPsRequest, "cursor"> = {},
  enabled = true,
) {
  return infiniteQueryOptions({
    queryKey: eventKeys.rsvps(eventId, params),
    initialPageParam: undefined as string | undefined,
    enabled: Boolean(eventId) && enabled,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/events/{id}/rsvps/search", {
          params: { path: { id: eventId } },
          body: { ...params, cursor: pageParam, limit: params.limit ?? 30 },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useEventRSVPs(
  eventId: string,
  params: Omit<SearchEventRSVPsRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(eventRSVPsInfiniteOptions(eventId, params, enabled));
}

export function useEventOrgs(enabled = true) {
  return useQuery({
    queryKey: eventKeys.orgs(),
    enabled,
    queryFn: () => unwrap(api.POST("/orgs/search", { body: { limit: 100 } })),
  });
}

export function useActiveParoki(enabled = true) {
  return useQuery({
    queryKey: eventKeys.paroki(),
    enabled,
    queryFn: () => unwrap(api.GET("/ref/paroki/current")),
  });
}
