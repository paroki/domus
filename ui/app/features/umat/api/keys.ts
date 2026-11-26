import type { SearchKeluargaRequest, SearchUmatRequest } from "./types";

export const umatKeys = {
  all: ["umat"] as const,
  keluargaLists: () => [...umatKeys.all, "keluarga", "list"] as const,
  keluargaList: (parokiId: string, body: SearchKeluargaRequest) =>
    [...umatKeys.keluargaLists(), parokiId, body] as const,
  keluargaDetails: () => [...umatKeys.all, "keluarga", "detail"] as const,
  keluargaDetail: (parokiId: string, id: string) =>
    [...umatKeys.keluargaDetails(), parokiId, id] as const,
  umatLists: () => [...umatKeys.all, "list"] as const,
  umatList: (parokiId: string, body: SearchUmatRequest) =>
    [...umatKeys.umatLists(), parokiId, body] as const,
  umatDetails: () => [...umatKeys.all, "detail"] as const,
  umatDetail: (parokiId: string, id: string) =>
    [...umatKeys.umatDetails(), parokiId, id] as const,
  refs: () => [...umatKeys.all, "refs"] as const,
  keuskupan: () => [...umatKeys.refs(), "keuskupan"] as const,
  paroki: (keuskupanId?: string) =>
    [...umatKeys.refs(), "paroki", keuskupanId ?? "all"] as const,
  parokiMembers: (parokiId: string) =>
    [...umatKeys.refs(), "parokiMembers", parokiId] as const,
  wilayah: (parokiId: string) =>
    [...umatKeys.refs(), "wilayah", parokiId] as const,
  lingkungan: (parokiId: string, wilayahId?: string) =>
    [...umatKeys.refs(), "lingkungan", parokiId, wilayahId ?? "all"] as const,
  kota: () => [...umatKeys.refs(), "kota"] as const,
  kecamatan: (kotaId?: string) =>
    [...umatKeys.refs(), "kecamatan", kotaId ?? "none"] as const,
  kelurahan: (kecamatanId?: string) =>
    [...umatKeys.refs(), "kelurahan", kecamatanId ?? "none"] as const,
  lookups: (parokiId: string, category: string) =>
    [...umatKeys.refs(), "lookups", parokiId, category] as const,
};
