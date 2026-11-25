import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useQuery,
} from "@tanstack/react-query";
import { api, unwrap, unwrapMeta } from "~/shared/api/client";
import { useMe } from "~/shared/api/session";
import { REF_ERRORS } from "./errors";
import { refKeys } from "./keys";
import type {
  SearchKeuskupanRequest,
  SearchLingkunganRequest,
  SearchLookupsRequest,
  SearchParokiMembersRequest,
  SearchParokiRequest,
  SearchPeriodsRequest,
  SearchWilayahRequest,
} from "./types";

export function useCurrentIdentity() {
  return useMe();
}

export function useKeuskupanSearch(
  filter: Omit<SearchKeuskupanRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.keuskupan(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/keuskupan/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useParokiSearch(
  filter: Omit<SearchParokiRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.paroki(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/paroki/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useCurrentParoki(enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.currentParoki(),
      enabled,
      queryFn: () => unwrap(api.GET("/ref/paroki/current"), REF_ERRORS),
      retry: false,
    }),
  );
}

export function useParokiMembersSearch(
  parokiId: string,
  filter: Omit<SearchParokiMembersRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.parokiMembers(parokiId, filter),
      initialPageParam: undefined as string | undefined,
      enabled: enabled && Boolean(parokiId),
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/paroki/{id}/members/search", {
            params: { path: { id: parokiId } },
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useWilayahSearch(
  filter: Omit<SearchWilayahRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.wilayah(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/wilayah/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useLingkunganSearch(
  filter: Omit<SearchLingkunganRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.lingkungan(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/lingkungan/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useLookupsSearch(
  filter: Omit<SearchLookupsRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.lookups(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/lookups/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function usePeriodsSearch(
  filter: Omit<SearchPeriodsRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.periods(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/periods/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}
