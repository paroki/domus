import type {
  SearchKecamatanRequest,
  SearchKelurahanRequest,
  SearchKeuskupanRequest,
  SearchKevikepanRequest,
  SearchKotaRequest,
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
  keuskupanDetail: (id: string) =>
    [...refKeys.all, "keuskupan-detail", id] as const,
  kevikepan: (params?: Omit<SearchKevikepanRequest, "cursor">) =>
    [...refKeys.all, "kevikepan", params ?? {}] as const,
  kevikepanDetail: (id: string) =>
    [...refKeys.all, "kevikepan-detail", id] as const,
  kota: (params?: Omit<SearchKotaRequest, "cursor">) =>
    [...refKeys.all, "kota", params ?? {}] as const,
  kotaDetail: (id: string) => [...refKeys.all, "kota-detail", id] as const,
  kecamatan: (params?: Omit<SearchKecamatanRequest, "cursor">) =>
    [...refKeys.all, "kecamatan", params ?? {}] as const,
  kecamatanDetail: (id: string) =>
    [...refKeys.all, "kecamatan-detail", id] as const,
  kelurahan: (params?: Omit<SearchKelurahanRequest, "cursor">) =>
    [...refKeys.all, "kelurahan", params ?? {}] as const,
  kelurahanDetail: (id: string) =>
    [...refKeys.all, "kelurahan-detail", id] as const,
  paroki: (params?: Omit<SearchParokiRequest, "cursor">) =>
    [...refKeys.all, "paroki", params ?? {}] as const,
  parokiDetail: (id: string) => [...refKeys.all, "paroki-detail", id] as const,
  currentParoki: () => [...refKeys.all, "current-paroki"] as const,
  parokiMembers: (
    parokiId: string,
    params?: Omit<SearchParokiMembersRequest, "cursor">,
  ) => [...refKeys.all, "paroki-members", parokiId, params ?? {}] as const,
  wilayah: (params?: Omit<SearchWilayahRequest, "cursor">) =>
    [...refKeys.all, "wilayah", params ?? {}] as const,
  wilayahDetail: (id: string) =>
    [...refKeys.all, "wilayah-detail", id] as const,
  lingkungan: (params?: Omit<SearchLingkunganRequest, "cursor">) =>
    [...refKeys.all, "lingkungan", params ?? {}] as const,
  lingkunganDetail: (id: string) =>
    [...refKeys.all, "lingkungan-detail", id] as const,
  lookups: (params?: Omit<SearchLookupsRequest, "cursor">) =>
    [...refKeys.all, "lookups", params ?? {}] as const,
  lookupDetail: (id: string) => [...refKeys.all, "lookup-detail", id] as const,
  periods: (params?: Omit<SearchPeriodsRequest, "cursor">) =>
    [...refKeys.all, "periods", params ?? {}] as const,
};
