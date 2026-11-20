import type {
  SearchDiocesesRequest,
  SearchLookupsRequest,
  SearchParishesRequest,
  SearchParishMembersRequest,
  SearchPeriodsRequest,
  SearchWardsRequest,
  SearchZonesRequest,
} from "./types";

export const refKeys = {
  all: ["ref"] as const,
  dioceses: (params?: Omit<SearchDiocesesRequest, "cursor">) =>
    [...refKeys.all, "dioceses", params ?? {}] as const,
  parishes: (params?: Omit<SearchParishesRequest, "cursor">) =>
    [...refKeys.all, "parishes", params ?? {}] as const,
  currentParish: () => [...refKeys.all, "current-parish"] as const,
  parishMembers: (
    parishId: string,
    params?: Omit<SearchParishMembersRequest, "cursor">,
  ) => [...refKeys.all, "parish-members", parishId, params ?? {}] as const,
  zones: (params?: Omit<SearchZonesRequest, "cursor">) =>
    [...refKeys.all, "zones", params ?? {}] as const,
  wards: (params?: Omit<SearchWardsRequest, "cursor">) =>
    [...refKeys.all, "wards", params ?? {}] as const,
  lookups: (params?: Omit<SearchLookupsRequest, "cursor">) =>
    [...refKeys.all, "lookups", params ?? {}] as const,
  periods: (params?: Omit<SearchPeriodsRequest, "cursor">) =>
    [...refKeys.all, "periods", params ?? {}] as const,
};
