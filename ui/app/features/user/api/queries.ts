import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useQuery,
} from "@tanstack/react-query";
import { api, unwrap, unwrapMeta } from "~/shared/api/client";
import { USER_ERRORS } from "./errors";
import { userKeys } from "./keys";
import type { SearchUsersRequest } from "./types";

export function authProvidersQueryOptions() {
  return queryOptions({
    queryKey: userKeys.providers(),
    queryFn: async () => {
      const res = await unwrap(api.GET("/auth/providers"), USER_ERRORS);
      return res.providers ?? [];
    },
    staleTime: 5 * 60_000,
  });
}

export function useAuthProviders() {
  return useQuery(authProvidersQueryOptions());
}

export function meQueryOptions() {
  return queryOptions({
    queryKey: userKeys.me(),
    queryFn: () => unwrap(api.GET("/me"), USER_ERRORS),
    retry: false,
  });
}

export function useMe() {
  return useQuery(meQueryOptions());
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
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: userKeys.list(filter),
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
