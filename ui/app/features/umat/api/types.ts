import type { components } from "~/shared/api/schema";

export type User = components["schemas"]["domain.UserResponse"];

export type KeluargaListItem =
  components["schemas"]["domain.KeluargaListItemResponse"];
export type KeluargaDetail =
  components["schemas"]["domain.KeluargaDetailResponse"];
export type KeluargaAnggotaSummary =
  components["schemas"]["domain.KeluargaAnggotaSummary"];
export type SearchKeluargaRequest =
  components["schemas"]["domain.SearchKeluargaRequest"];
export type CreateKeluargaRequest =
  components["schemas"]["domain.CreateKeluargaRequest"];
export type UpdateKeluargaRequest =
  components["schemas"]["domain.UpdateKeluargaRequest"];

export type UmatListItem = components["schemas"]["domain.UmatListItemResponse"];
export type UmatDetail = components["schemas"]["domain.UmatDetailResponse"];
export type SearchUmatRequest =
  components["schemas"]["domain.SearchUmatRequest"];

export type Wilayah = components["schemas"]["domain.WilayahResponse"];
export type Lingkungan = components["schemas"]["domain.LingkunganResponse"];
export type Kota = components["schemas"]["domain.KotaResponse"];
export type Kecamatan = components["schemas"]["domain.KecamatanResponse"];
export type Kelurahan = components["schemas"]["domain.KelurahanResponse"];
export type Lookup = components["schemas"]["domain.LookupResponse"];
