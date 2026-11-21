import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useQuery,
} from "@tanstack/react-query";
import { api, unwrap, unwrapMeta } from "~/shared/api/client";
import { meQueryOptions, useMe } from "~/shared/api/session";
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
