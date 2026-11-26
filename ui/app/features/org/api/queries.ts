import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useQuery,
} from "@tanstack/react-query";
import { api, unwrap, unwrapMeta } from "~/shared/api/client";
import { orgKeys } from "./keys";
import type {
  LookupUmatRequest,
  SearchOrganizationsRequest,
  SearchOrgInvitesRequest,
  SearchOrgJabatanRequest,
  SearchOrgMembersRequest,
  SearchOrgPeriodeRequest,
} from "./types";

export function orgsInfiniteOptions(
  params: Omit<SearchOrganizationsRequest, "cursor"> = {},
  enabled = true,
) {
  return infiniteQueryOptions({
    queryKey: orgKeys.orgs(params),
    initialPageParam: undefined as string | undefined,
    enabled,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/orgs/search", {
          body: { ...params, cursor: pageParam, limit: params.limit ?? 20 },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useOrgs(
  params: Omit<SearchOrganizationsRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(orgsInfiniteOptions(params, enabled));
}

export function orgDetailOptions(orgId: string, enabled = true) {
  return queryOptions({
    queryKey: orgKeys.org(orgId),
    enabled: Boolean(orgId) && enabled,
    queryFn: () =>
      unwrap(api.GET("/orgs/{id}", { params: { path: { id: orgId } } })),
  });
}

export function useOrg(orgId: string, enabled = true) {
  return useQuery(orgDetailOptions(orgId, enabled));
}

export function orgMembersInfiniteOptions(
  orgId: string,
  params: Omit<SearchOrgMembersRequest, "cursor"> = {},
  enabled = true,
) {
  return infiniteQueryOptions({
    queryKey: orgKeys.members(orgId, params),
    initialPageParam: undefined as string | undefined,
    enabled: Boolean(orgId) && enabled,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/orgs/{id}/members/search", {
          params: { path: { id: orgId } },
          body: { ...params, cursor: pageParam, limit: params.limit ?? 30 },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useOrgMembers(
  orgId: string,
  params: Omit<SearchOrgMembersRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(orgMembersInfiniteOptions(orgId, params, enabled));
}

export function orgInvitesInfiniteOptions(
  orgId: string,
  params: Omit<SearchOrgInvitesRequest, "cursor"> = {},
  enabled = true,
) {
  return infiniteQueryOptions({
    queryKey: orgKeys.invites(orgId, params),
    initialPageParam: undefined as string | undefined,
    enabled: Boolean(orgId) && enabled,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/orgs/{id}/invites/search", {
          params: { path: { id: orgId } },
          body: { ...params, cursor: pageParam, limit: params.limit ?? 20 },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useOrgInvites(
  orgId: string,
  params: Omit<SearchOrgInvitesRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(orgInvitesInfiniteOptions(orgId, params, enabled));
}

export function useOrgInvitePreview(code: string) {
  return useQuery({
    queryKey: orgKeys.invitePreview(code),
    enabled: Boolean(code),
    retry: false,
    queryFn: () =>
      unwrap(api.GET("/orgs/invites/{code}", { params: { path: { code } } })),
  });
}

export function useOrgWilayah(enabled = true) {
  return useQuery({
    queryKey: orgKeys.wilayah(),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/wilayah/search", {
          body: { filter: { active: true }, limit: 100 },
        }),
      ),
  });
}

export function useOrgLingkungan(enabled = true) {
  return useQuery({
    queryKey: orgKeys.lingkungan(),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/lingkungan/search", {
          body: { filter: { active: true }, limit: 100 },
        }),
      ),
  });
}

export function useOrgPositionLookups(enabled = true) {
  return useQuery({
    queryKey: orgKeys.positionLookups(),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/lookups/search", {
          body: {
            filter: { category: "org_position", active: true },
            limit: 100,
          },
        }),
      ),
  });
}

export function useOrgPeriodeList(
  orgId: string,
  params: Omit<SearchOrgPeriodeRequest, "cursor"> = {},
  enabled = true,
) {
  return useQuery({
    queryKey: orgKeys.periodeList(orgId, params),
    enabled: Boolean(orgId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/orgs/{id}/periode/search", {
          params: { path: { id: orgId } },
          body: { ...params, limit: params.limit ?? 100 },
        }),
      ),
  });
}

export function useOrgPeriodeDetail(
  orgId: string,
  periodeId: string,
  enabled = true,
) {
  return useQuery({
    queryKey: orgKeys.periodeDetail(orgId, periodeId),
    enabled: Boolean(orgId) && Boolean(periodeId) && enabled,
    queryFn: () =>
      unwrap(
        api.GET("/orgs/{id}/periode/{periodeId}", {
          params: { path: { id: orgId, periodeId } },
        }),
      ),
  });
}

export function useOrgJabatanList(
  orgId: string,
  params: Omit<SearchOrgJabatanRequest, "cursor"> = {},
  enabled = true,
) {
  return useQuery({
    queryKey: orgKeys.jabatanList(orgId, params),
    enabled: Boolean(orgId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/orgs/{id}/jabatan/search", {
          params: { path: { id: orgId } },
          body: { ...params, limit: params.limit ?? 100 },
        }),
      ),
  });
}

export function useOrgJabatanDetail(
  orgId: string,
  jabatanId: string,
  enabled = true,
) {
  return useQuery({
    queryKey: orgKeys.jabatanDetail(orgId, jabatanId),
    enabled: Boolean(orgId) && Boolean(jabatanId) && enabled,
    queryFn: () =>
      unwrap(
        api.GET("/orgs/{id}/jabatan/{jabatanId}", {
          params: { path: { id: orgId, jabatanId } },
        }),
      ),
  });
}

export function useOrgUmatLookup(
  params: LookupUmatRequest = {},
  enabled = true,
) {
  return useQuery({
    queryKey: orgKeys.umatLookup(params),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/umat/lookup", {
          body: { ...params, limit: params.limit ?? 50 },
        }),
      ),
  });
}

export function useOrgAuthProviders() {
  return useQuery({
    queryKey: orgKeys.authProviders(),
    queryFn: async () => {
      const res = await unwrap(api.GET("/auth/providers"));
      return res.providers ?? [];
    },
  });
}
