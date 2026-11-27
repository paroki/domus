import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { api, unwrap, unwrapMeta } from "~/shared/api/client";
import { cantusKeys } from "./keys";
import type { SearchCantusRequest } from "./types";

export function useCantusSearch(
  body: Omit<SearchCantusRequest, "cursor">,
  enabled = true,
) {
  return useInfiniteQuery({
    queryKey: cantusKeys.list(body),
    enabled,
    initialPageParam: undefined as string | undefined,
    queryFn: ({ pageParam }) =>
      unwrapMeta(
        api.POST("/cantus/search", {
          body: { ...body, limit: body.limit ?? 20, cursor: pageParam },
        }),
      ),
    getNextPageParam: (last) =>
      last.meta?.hasMore ? last.meta.nextCursor : undefined,
  });
}

export function useCantusDetail(id?: string, enabled = true) {
  return useQuery({
    queryKey: cantusKeys.detail(id ?? ""),
    enabled: Boolean(id) && enabled,
    queryFn: () =>
      unwrap(
        api.GET("/cantus/{id}", {
          params: { path: { id: id ?? "" } },
        }),
      ),
  });
}
