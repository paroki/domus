import type { SearchCantusRequest } from "./types";

export const cantusKeys = {
  all: ["cantus"] as const,
  lists: () => [...cantusKeys.all, "list"] as const,
  list: (body: SearchCantusRequest) => [...cantusKeys.lists(), body] as const,
  details: () => [...cantusKeys.all, "detail"] as const,
  detail: (id: string) => [...cantusKeys.details(), id] as const,
};
