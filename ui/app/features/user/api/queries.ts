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
  SearchParishMembersRequest,
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

export function useDirectoryParishes(enabled = true) {
  return useQuery({
    queryKey: userKeys.parishes(),
    enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/parishes/search", { body: { limit: 100 } }),
        USER_ERRORS,
      ),
  });
}

export function useUserParishMemberships(
  userId: string | undefined,
  email: string | undefined,
  enabled = true,
) {
  const parishesQuery = useDirectoryParishes(enabled);
  const parishes = parishesQuery.data ?? [];

  const memberQueries = useQueries({
    queries: parishes.map((parish) => ({
      queryKey: userKeys.parishMembers(parish.id ?? "", {
        q: email,
        limit: 20,
      }),
      enabled: Boolean(enabled && userId && email && parish.id),
      queryFn: () =>
        unwrap(
          api.POST("/ref/parishes/{id}/members/search", {
            params: { path: { id: parish.id ?? "" } },
            body: { q: email, limit: 20 },
          }),
          USER_ERRORS,
        ),
    })),
  });

  const memberships = parishes.flatMap((parish, idx) => {
    const match = memberQueries[idx]?.data?.find((m) => m.userId === userId);
    return match ? [{ parish, member: match }] : [];
  });

  return {
    parishes,
    memberships,
    isLoading:
      parishesQuery.isLoading || memberQueries.some((q) => q.isLoading),
  };
}

export function useDirectoryParishMembers(
  parishId: string | undefined,
  filter: Omit<SearchParishMembersRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: userKeys.parishMembers(parishId ?? "", filter),
      enabled: Boolean(parishId) && enabled,
      initialPageParam: undefined as string | undefined,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/parishes/{id}/members/search", {
            params: { path: { id: parishId ?? "" } },
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

export function useDirectoryOrgs(parishId: string | undefined, enabled = true) {
  return useQuery({
    queryKey: userKeys.orgs(parishId ?? ""),
    enabled: Boolean(parishId) && enabled,
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
  parishId: string | undefined,
  enabled = true,
) {
  const orgsQuery = useDirectoryOrgs(parishId, enabled);
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

export function useDirectoryPeriods(
  parishId: string | undefined,
  enabled = true,
) {
  return useQuery({
    queryKey: userKeys.periods(parishId ?? ""),
    enabled: Boolean(parishId) && enabled,
    queryFn: () =>
      unwrap(
        api.POST("/ref/periods/search", { body: { limit: 100 } }),
        USER_ERRORS,
      ),
  });
}

export function useDirectoryManagedOrgUsers(
  parishId: string | undefined,
  q: string,
  enabled = true,
): {
  users: DirectoryOrgUser[];
  isLoading: boolean;
} {
  const orgsQuery = useDirectoryOrgs(parishId, enabled);
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
      enabled: Boolean(enabled && parishId && org.id),
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
