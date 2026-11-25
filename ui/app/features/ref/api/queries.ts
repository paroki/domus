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
  SearchKecamatanRequest,
  SearchKelurahanRequest,
  SearchKeuskupanRequest,
  SearchKevikepanRequest,
  SearchKotaRequest,
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

export function useKeuskupanDetail(id: string, enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.keuskupanDetail(id),
      enabled: enabled && Boolean(id),
      queryFn: () =>
        unwrap(
          api.GET("/ref/keuskupan/{id}", { params: { path: { id } } }),
          REF_ERRORS,
        ),
    }),
  );
}

export function useParokiDetail(id: string, enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.parokiDetail(id),
      enabled: enabled && Boolean(id),
      queryFn: () =>
        unwrap(
          api.GET("/ref/paroki/{id}", { params: { path: { id } } }),
          REF_ERRORS,
        ),
    }),
  );
}

export function useWilayahDetail(id: string, enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.wilayahDetail(id),
      enabled: enabled && Boolean(id),
      queryFn: () =>
        unwrap(
          api.GET("/ref/wilayah/{id}", { params: { path: { id } } }),
          REF_ERRORS,
        ),
    }),
  );
}

export function useLingkunganDetail(id: string, enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.lingkunganDetail(id),
      enabled: enabled && Boolean(id),
      queryFn: () =>
        unwrap(
          api.GET("/ref/lingkungan/{id}", { params: { path: { id } } }),
          REF_ERRORS,
        ),
    }),
  );
}

export function useLookupDetail(id: string, enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.lookupDetail(id),
      enabled: enabled && Boolean(id),
      queryFn: () =>
        unwrap(
          api.GET("/ref/lookups/{id}", { params: { path: { id } } }),
          REF_ERRORS,
        ),
    }),
  );
}

export function useKevikepanSearch(
  filter: Omit<SearchKevikepanRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.kevikepan(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/kevikepan/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useKevikepanDetail(id: string, enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.kevikepanDetail(id),
      enabled: enabled && Boolean(id),
      queryFn: () =>
        unwrap(
          api.GET("/ref/kevikepan/{id}", { params: { path: { id } } }),
          REF_ERRORS,
        ),
    }),
  );
}

export function useKotaSearch(
  filter: Omit<SearchKotaRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.kota(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/kota/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useKotaDetail(id: string, enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.kotaDetail(id),
      enabled: enabled && Boolean(id),
      queryFn: () =>
        unwrap(
          api.GET("/ref/kota/{id}", { params: { path: { id } } }),
          REF_ERRORS,
        ),
    }),
  );
}

export function useKecamatanSearch(
  filter: Omit<SearchKecamatanRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.kecamatan(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/kecamatan/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useKecamatanDetail(id: string, enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.kecamatanDetail(id),
      enabled: enabled && Boolean(id),
      queryFn: () =>
        unwrap(
          api.GET("/ref/kecamatan/{id}", { params: { path: { id } } }),
          REF_ERRORS,
        ),
    }),
  );
}

export function useKelurahanSearch(
  filter: Omit<SearchKelurahanRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: refKeys.kelurahan(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/kelurahan/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
          REF_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useKelurahanDetail(id: string, enabled = true) {
  return useQuery(
    queryOptions({
      queryKey: refKeys.kelurahanDetail(id),
      enabled: enabled && Boolean(id),
      queryFn: () =>
        unwrap(
          api.GET("/ref/kelurahan/{id}", { params: { path: { id } } }),
          REF_ERRORS,
        ),
    }),
  );
}
