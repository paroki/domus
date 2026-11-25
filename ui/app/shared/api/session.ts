import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  api,
  setActiveParish,
  unwrap,
  unwrapMeta,
  unwrapVoid,
} from "~/shared/api/client";
import type { components } from "~/shared/api/schema";

type SearchParokiRequest = components["schemas"]["domain.SearchParokiRequest"];
type SwitchSessionParokiRequest =
  components["schemas"]["domain.SwitchSessionParokiRequest"];

export const sessionKeys = {
  me: ["user", "me"] as const,
  sessions: ["user", "sessions"] as const,
  ref: ["ref"] as const,
  parishes: (params: Omit<SearchParokiRequest, "cursor"> = {}) =>
    ["ref", "paroki", params] as const,
};

export function meQueryOptions() {
  return queryOptions({
    queryKey: sessionKeys.me,
    queryFn: () => unwrap(api.GET("/me")),
    retry: false,
  });
}

export function useMe() {
  return useQuery(meQueryOptions());
}

export function useSwitchParish() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: SwitchSessionParokiRequest) => {
      await unwrapVoid(api.PATCH("/me/session", { body }));
      return body.activeParokiId;
    },
    onSuccess: async (parokiId) => {
      setActiveParish(parokiId);
      await Promise.all([
        qc.invalidateQueries({ queryKey: sessionKeys.me }),
        qc.invalidateQueries({ queryKey: sessionKeys.sessions }),
        qc.invalidateQueries({ queryKey: sessionKeys.ref }),
      ]);
    },
  });
}

export function useLogout() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      await api.POST("/auth/logout");
    },
    onSettled: () => {
      setActiveParish(null);
      qc.clear();
    },
  });
}

export function useSessionParishes(
  filter: Omit<SearchParokiRequest, "cursor"> = {},
  enabled = true,
) {
  return useInfiniteQuery(
    infiniteQueryOptions({
      queryKey: sessionKeys.parishes(filter),
      initialPageParam: undefined as string | undefined,
      enabled,
      queryFn: ({ pageParam }) =>
        unwrapMeta(
          api.POST("/ref/paroki/search", {
            body: { ...filter, cursor: pageParam, limit: filter.limit ?? 50 },
          }),
        ),
      getNextPageParam: (lastPage) =>
        lastPage.meta?.hasMore ? lastPage.meta.nextCursor : undefined,
    }),
  );
}
