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
  SearchDiocesesRequest,
  SearchLookupsRequest,
  SearchParishesRequest,
  SearchParishMembersRequest,
  SearchPeriodsRequest,
  SearchWardsRequest,
  SearchZonesRequest,
} from "./types";

export function useCurrentIdentity() {
  return useMe();
}

export function useDiocesesSearch(
  filter: Omit<SearchDiocesesRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.dioceses(filter),
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

export function useParishesSearch(
  filter: Omit<SearchParishesRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.parishes(filter),
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

export function useCurrentParish(enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.currentParish(),
      enabled,
      queryFn: () => unwrap(api.GET("/ref/paroki/current"), REF_ERRORS),
      retry: false,
    }),
  );
}

export function useParishMembersSearch(
  parishId: string,
  filter: Omit<SearchParishMembersRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.parishMembers(parishId, filter),
      initialPageParam: undefined as string | undefined,
      enabled: enabled && Boolean(parishId),
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/paroki/{id}/members/search", {
            params: { path: { id: parishId } },
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useZonesSearch(
  filter: Omit<SearchZonesRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.zones(filter),
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

export function useWardsSearch(
  filter: Omit<SearchWardsRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.wards(filter),
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
