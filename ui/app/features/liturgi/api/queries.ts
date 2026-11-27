import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useQuery,
} from "@tanstack/react-query";
import { api, unwrap, unwrapMeta } from "~/shared/api/client";
import { useMe } from "~/shared/api/session";
import { liturgiKeys } from "./keys";
import type { SearchLiturgiSchedulesRequest } from "./types";

export function liturgiMinistriesOptions(enabled = true) {
  return queryOptions({
    queryKey: liturgiKeys.ministries(),
    enabled,
    queryFn: () => unwrap(api.GET("/liturgi/ministries")),
  });
}

export function useLiturgiMinistries(enabled = true) {
  return useQuery(liturgiMinistriesOptions(enabled));
}

export function liturgiSchedulesInfiniteOptions(
  params: Omit<SearchLiturgiSchedulesRequest, "cursor"> = {},
  enabled = true,
) {
  return infiniteQueryOptions({
    queryKey: liturgiKeys.schedules(params),
    initialPageParam: undefined as string | undefined,
    enabled,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/liturgi/schedules/search", {
          body: { ...params, cursor: pageParam, limit: params.limit ?? 20 },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useLiturgiSchedules(
  params: Omit<SearchLiturgiSchedulesRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(liturgiSchedulesInfiniteOptions(params, enabled));
}

export function liturgiScheduleDetailOptions(eventId: string, enabled = true) {
  return queryOptions({
    queryKey: liturgiKeys.schedule(eventId),
    enabled: Boolean(eventId) && enabled,
    queryFn: () =>
      unwrap(
        api.GET("/liturgi/schedules/{eventId}", {
          params: { path: { eventId } },
        }),
      ),
  });
}

export function useLiturgiSchedule(eventId: string, enabled = true) {
  return useQuery(liturgiScheduleDetailOptions(eventId, enabled));
}

export function useActiveParoki() {
  const { data: me } = useMe();
  const parokiId = me?.activeParokiId ?? "";
  return useQuery({
    queryKey: liturgiKeys.paroki(parokiId),
    enabled: Boolean(parokiId),
    queryFn: () =>
      unwrap(
        api.GET("/ref/paroki/{id}", {
          params: { path: { id: parokiId } },
        }),
      ),
  });
}

export function useLiturgiOrgs(enabled = true) {
  return useQuery({
    queryKey: liturgiKeys.orgs(),
    enabled,
    queryFn: () => unwrap(api.POST("/orgs/search", { body: { limit: 100 } })),
  });
}

export function useLiturgiParokiMembers(
  parokiId: string,
  q = "",
  enabled = true,
) {
  return useQuery({
    queryKey: liturgiKeys.parokiMembers(parokiId, q),
    enabled: Boolean(parokiId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/paroki/{id}/members/search", {
          params: { path: { id: parokiId } },
          body: {
            q: q.trim() || undefined,
            limit: 100,
          },
        }),
      ),
  });
}

export function useLiturgiSongPartLookups(enabled = true) {
  return useQuery({
    queryKey: liturgiKeys.songPartLookups(),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/lookups/search", {
          body: {
            filter: { category: "liturgi_song_part", active: true },
            limit: 50,
          },
        }),
      ),
  });
}

export function useLiturgiCantusSearch(q = "", enabled = true) {
  const trimmed = q.trim();
  return useQuery({
    queryKey: liturgiKeys.cantusSearch(trimmed),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/cantus/search", {
          body: {
            q: trimmed || undefined,
            filter: { active: true },
            limit: 20,
          },
        }),
      ),
  });
}
