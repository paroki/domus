import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useQueries,
  useQuery,
} from "@tanstack/react-query";
import { api, unwrap, unwrapMeta } from "~/shared/api/client";
import { meQueryOptions, useMe } from "~/shared/api/session";
import { type DirectoryOrgUser, deduplicateOrgMembers } from "../model";
import { USER_ERRORS } from "./errors";
import { userKeys } from "./keys";
import type {
  SearchOrgMembersRequest,
  SearchParokiMembersRequest,
  SearchUsersRequest,
} from "./types";

export { meQueryOptions, useMe };

export function useAuthRoles() {
  return useQuery({
    queryKey: userKeys.roles(),
    queryFn: () => unwrap(api.GET("/auth/roles"), USER_ERRORS),
    staleTime: 5 * 60_000,
  });
}

export function authProvidersQueryOptions() {
  return queryOptions({
    queryKey: userKeys.providers(),
    queryFn: async () => {
      const res = await unwrap(api.GET("/auth/providers"), USER_ERRORS);
      return {
        providers: res.providers ?? [],
        otpChannels: res.otpChannels ?? ["email"],
      };
    },
    staleTime: 5 * 60_000,
  });
}

export function useAuthProviders() {
  return useQuery(authProvidersQueryOptions());
}

export function useVapidPublicKey() {
  return useQuery({
    queryKey: userKeys.vapidPublicKey(),
    queryFn: () => unwrap(api.GET("/auth/push/vapid-public-key"), USER_ERRORS),
    staleTime: 10 * 60_000,
  });
}

export function useMyUmat(enabled = true) {
  return useQuery({
    queryKey: userKeys.myUmat(),
    enabled,
    queryFn: () => unwrap(api.GET("/me/umat"), USER_ERRORS),
  });
}

export function useVerifyWilayahOptions(parokiId: string, enabled = true) {
  return useQuery({
    queryKey: userKeys.wilayahOptions(parokiId),
    enabled: Boolean(parokiId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/wilayah/search", {
          body: { filter: { active: true }, limit: 100 },
        }),
        USER_ERRORS,
      ),
    staleTime: 5 * 60_000,
  });
}

export function useVerifyLingkunganOptions(
  parokiId: string,
  wilayahId?: string,
  enabled = true,
) {
  return useQuery({
    queryKey: userKeys.lingkunganOptions(parokiId, wilayahId),
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
        USER_ERRORS,
      ),
    staleTime: 5 * 60_000,
  });
}

export function useMySessions(limit = 20) {
  return useQuery({
    queryKey: userKeys.sessions({ limit }),
    queryFn: () =>
      unwrapMeta(
        api.POST("/me/sessions/search", { body: { limit } }),
        USER_ERRORS,
      ),
  });
}

export function useMyTokens(limit = 20) {
  return useQuery({
    queryKey: userKeys.tokens({ limit }),
    queryFn: () =>
      unwrapMeta(
        api.POST("/me/tokens/search", { body: { limit } }),
        USER_ERRORS,
      ),
  });
}

export function useMyAccounts(limit = 20) {
  return useQuery({
    queryKey: userKeys.accounts({ limit }),
    queryFn: () =>
      unwrapMeta(
        api.POST("/me/accounts/search", { body: { limit } }),
        USER_ERRORS,
      ),
  });
}

export function useUsersSearch(
  filter: Omit<SearchUsersRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: userKeys.list(filter),
      enabled,
      initialPageParam: undefined as string | undefined,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/users/search", {
            body: {
              ...filter,
              cursor: pageParam,
              limit: filter.limit ?? 20,
            },
          }),
          USER_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useUserDetail(id: string | undefined, enabled = true) {
  return useQuery({
    queryKey: userKeys.detail(id ?? ""),
    enabled: Boolean(id) && enabled,
    queryFn: () =>
      unwrap(
        api.GET("/users/{id}", { params: { path: { id: id ?? "" } } }),
        USER_ERRORS,
      ),
  });
}

export function useDirectoryParoki(enabled = true) {
  return useQuery({
    queryKey: userKeys.paroki(),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/paroki/search", { body: { limit: 100 } }),
        USER_ERRORS,
      ),
  });
}

export function useUserParokiMemberships(
  userId: string | undefined,
  email: string | undefined,
  enabled = true,
) {
  const parokiQuery = useDirectoryParoki(enabled);
  const parokiList = parokiQuery.data ?? [];

  const memberQueries = useQueries({
    queries: parokiList.map((paroki) => ({
      queryKey: userKeys.parokiMembers(paroki.id ?? "", {
        q: email,
        limit: 20,
      }),
      enabled: Boolean(enabled && userId && email && paroki.id),
      queryFn: () =>
        unwrap(
          api.POST("/ref/paroki/{id}/members/search", {
            params: { path: { id: paroki.id ?? "" } },
            body: { q: email, limit: 20 },
          }),
          USER_ERRORS,
        ),
    })),
  });

  const memberships = parokiList.flatMap((paroki, idx) => {
    const match = memberQueries[idx]?.data?.find((m) => m.userId === userId);
    return match ? [{ paroki, member: match }] : [];
  });

  return {
    paroki: parokiList,
    memberships,
    isLoading: parokiQuery.isLoading || memberQueries.some((q) => q.isLoading),
  };
}

export function useDirectoryParokiMembers(
  parokiId: string | undefined,
  filter: Omit<SearchParokiMembersRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: userKeys.parokiMembers(parokiId ?? "", filter),
      enabled: Boolean(parokiId) && enabled,
      initialPageParam: undefined as string | undefined,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/paroki/{id}/members/search", {
            params: { path: { id: parokiId ?? "" } },
            body: {
              ...filter,
              cursor: pageParam,
              limit: filter.limit ?? 20,
            },
          }),
          USER_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useDirectoryOrgs(parokiId: string | undefined, enabled = true) {
  return useQuery({
    queryKey: userKeys.orgs(parokiId ?? ""),
    enabled: Boolean(parokiId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/orgs/search", {
          body: { filter: { active: true }, limit: 100 },
        }),
        USER_ERRORS,
      ),
  });
}

export function useUserOrgMemberships(
  userId: string | undefined,
  email: string | undefined,
  parokiId: string | undefined,
  enabled = true,
) {
  const orgsQuery = useDirectoryOrgs(parokiId, enabled);
  const orgs = orgsQuery.data ?? [];
  const inspectableOrgs = orgs.filter(
    (o) => Boolean(o.id) && (o.myRole === "admin" || o.myRole === "member"),
  );

  const memberQueries = useQueries({
    queries: inspectableOrgs.map((org) => ({
      queryKey: userKeys.orgMembers(org.id ?? "", {
        q: email,
        limit: 20,
      }),
      enabled: Boolean(enabled && userId && email && org.id),
      queryFn: () =>
        unwrap(
          api.POST("/orgs/{id}/members/search", {
            params: { path: { id: org.id ?? "" } },
            body: { q: email, limit: 20 },
          }),
          USER_ERRORS,
        ),
    })),
  });

  const memberships = inspectableOrgs.flatMap((org, idx) => {
    const match = memberQueries[idx]?.data?.find((m) => m.userId === userId);
    return match ? [{ org, member: match }] : [];
  });

  return {
    orgs,
    inspectableOrgs,
    memberships,
    isLoading: orgsQuery.isLoading || memberQueries.some((q) => q.isLoading),
  };
}

export function useDirectoryOrgMembers(
  orgId: string | undefined,
  filter: Omit<SearchOrgMembersRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: userKeys.orgMembers(orgId ?? "", filter),
      enabled: Boolean(orgId) && enabled,
      initialPageParam: undefined as string | undefined,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/orgs/{id}/members/search", {
            params: { path: { id: orgId ?? "" } },
            body: {
              ...filter,
              cursor: pageParam,
              limit: filter.limit ?? 20,
            },
          }),
          USER_ERRORS,
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}

export function useDirectoryLookups(category: string, enabled = true) {
  return useQuery({
    queryKey: userKeys.lookups(category),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/lookups/search", {
          body: { filter: { category, active: true }, limit: 100 },
        }),
        USER_ERRORS,
      ),
  });
}

export function useDirectoryManagedOrgUsers(
  parokiId: string | undefined,
  q: string,
  enabled = true,
): {
  users: DirectoryOrgUser[];
  isLoading: boolean;
} {
  const orgsQuery = useDirectoryOrgs(parokiId, enabled);
  const manageableOrgs = (orgsQuery.data ?? []).filter(
    (o) => Boolean(o.id) && o.myRole === "admin",
  );

  // ponytail: fan-out search over manageable orgs and deduplicate by userId
  const memberQueries = useQueries({
    queries: manageableOrgs.map((org) => ({
      queryKey: userKeys.orgMembers(org.id ?? "", {
        q: q || undefined,
        limit: 100,
      }),
      enabled: Boolean(enabled && parokiId && org.id),
      queryFn: () =>
        unwrap(
          api.POST("/orgs/{id}/members/search", {
            params: { path: { id: org.id ?? "" } },
            body: { q: q || undefined, limit: 100 },
          }),
          USER_ERRORS,
        ),
    })),
  });

  const users = deduplicateOrgMembers(
    manageableOrgs.map((org, idx) => ({
      org,
      members: memberQueries[idx]?.data ?? [],
    })),
  );

  return {
    users,
    isLoading:
      orgsQuery.isLoading || memberQueries.some((query) => query.isLoading),
  };
}
