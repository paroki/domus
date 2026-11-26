import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { api, unwrap, unwrapMeta } from "~/shared/api/client";
import { umatKeys } from "./keys";
import type { SearchKeluargaRequest, SearchUmatRequest } from "./types";

export function useUmatSearch(
  parokiId: string,
  body: Omit<SearchUmatRequest, "cursor">,
  enabled = true,
) {
  return useInfiniteQuery({
    queryKey: umatKeys.umatList(parokiId, body),
    enabled: Boolean(parokiId) && enabled,
    initialPageParam: undefined as string | undefined,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/umat/search", {
          body: { ...body, limit: body.limit ?? 25, cursor: pageParam },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useKeluargaSearch(
  parokiId: string,
  body: Omit<SearchKeluargaRequest, "cursor">,
  enabled = true,
) {
  return useInfiniteQuery({
    queryKey: umatKeys.keluargaList(parokiId, body),
    enabled: Boolean(parokiId) && enabled,
    initialPageParam: undefined as string | undefined,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/umat/keluarga/search", {
          body: { ...body, limit: body.limit ?? 25, cursor: pageParam },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useKeluargaDetail(
  parokiId: string,
  id?: string,
  enabled = true,
) {
  return useQuery({
    queryKey: umatKeys.keluargaDetail(parokiId, id ?? ""),
    enabled: Boolean(parokiId) && Boolean(id) && enabled,
    queryFn: () =>
      unwrap(
        api.GET("/umat/keluarga/{id}", {
          params: { path: { id: id ?? "" } },
        }),
      ),
  });
}

export function useUmatWilayahList(parokiId: string, enabled = true) {
  return useQuery({
    queryKey: umatKeys.wilayah(parokiId),
    enabled: Boolean(parokiId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/wilayah/search", {
          body: { filter: { active: true }, limit: 100 },
        }),
      ),
  });
}

export function useUmatLingkunganList(
  parokiId: string,
  wilayahId?: string,
  enabled = true,
) {
  return useQuery({
    queryKey: umatKeys.lingkungan(parokiId, wilayahId),
    enabled: Boolean(parokiId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/lingkungan/search", {
          body: {
            filter: {
              active: true,
              ...(wilayahId ? { wilayahId } : {}),
            },
            limit: 100,
          },
        }),
      ),
  });
}

export function useUmatKotaList(enabled = true) {
  return useQuery({
    queryKey: umatKeys.kota(),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/kota/search", {
          body: { filter: { active: true }, limit: 100 },
        }),
      ),
  });
}

export function useUmatKecamatanList(kotaId?: string, enabled = true) {
  return useQuery({
    queryKey: umatKeys.kecamatan(kotaId),
    enabled: Boolean(kotaId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/kecamatan/search", {
          body: { filter: { active: true, kotaId }, limit: 100 },
        }),
      ),
  });
}

export function useUmatKelurahanList(kecamatanId?: string, enabled = true) {
  return useQuery({
    queryKey: umatKeys.kelurahan(kecamatanId),
    enabled: Boolean(kecamatanId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/kelurahan/search", {
          body: { filter: { active: true, kecamatanId }, limit: 100 },
        }),
      ),
  });
}

export function useUmatDetail(parokiId: string, id?: string, enabled = true) {
  return useQuery({
    queryKey: umatKeys.umatDetail(parokiId, id ?? ""),
    enabled: Boolean(parokiId) && Boolean(id) && enabled,
    queryFn: () =>
      unwrap(
        api.GET("/umat/{id}", {
          params: { path: { id: id ?? "" } },
        }),
      ),
  });
}

export function useUmatKeuskupanList(enabled = true) {
  return useQuery({
    queryKey: umatKeys.keuskupan(),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/keuskupan/search", {
          body: { filter: { active: true }, limit: 100 },
        }),
      ),
  });
}

export function useUmatParokiList(keuskupanId?: string, enabled = true) {
  return useQuery({
    queryKey: umatKeys.paroki(keuskupanId),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/paroki/search", {
          body: {
            filter: {
              active: true,
              ...(keuskupanId ? { keuskupanId } : {}),
            },
            limit: 100,
          },
        }),
      ),
  });
}

export function useUmatParokiMembersList(parokiId: string, enabled = true) {
  return useQuery({
    queryKey: umatKeys.parokiMembers(parokiId),
    enabled: Boolean(parokiId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/paroki/{id}/members/search", {
          params: { path: { id: parokiId } },
          body: { limit: 100 },
        }),
      ),
  });
}

export function useUmatLookups(
  parokiId: string,
  category: string,
  enabled = true,
) {
  return useQuery({
    queryKey: umatKeys.lookups(parokiId, category),
    enabled: Boolean(parokiId) && Boolean(category) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/lookups/search", {
          body: { filter: { active: true, category }, limit: 100 },
        }),
      ),
  });
}
