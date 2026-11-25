import type {
  SearchKeuskupanRequest,
  SearchLingkunganRequest,
  SearchLookupsRequest,
  SearchParokiMembersRequest,
  SearchParokiRequest,
  SearchPeriodsRequest,
  SearchWilayahRequest,
} from "./types";

export const refKeys = {
  all: ["ref"] as const,
  keuskupan: (params?: Omit<SearchKeuskupanRequest, "cursor">) =>
    [...refKeys.all, "keuskupan", params ?? {}] as const,
  paroki: (params?: Omit<SearchParokiRequest, "cursor">) =>
    [...refKeys.all, "paroki", params ?? {}] as const,
  currentParoki: () => [...refKeys.all, "current-paroki"] as const,
  parokiMembers: (
    parokiId: string,
    params?: Omit<SearchParokiMembersRequest, "cursor">,
  ) => [...refKeys.all, "paroki-members", parokiId, params ?? {}] as const,
  wilayah: (params?: Omit<SearchWilayahRequest, "cursor">) =>
    [...refKeys.all, "wilayah", params ?? {}] as const,
  lingkungan: (params?: Omit<SearchLingkunganRequest, "cursor">) =>
    [...refKeys.all, "lingkungan", params ?? {}] as const,
  lookups: (params?: Omit<SearchLookupsRequest, "cursor">) =>
    [...refKeys.all, "lookups", params ?? {}] as const,
  periods: (params?: Omit<SearchPeriodsRequest, "cursor">) =>
    [...refKeys.all, "periods", params ?? {}] as const,
};
