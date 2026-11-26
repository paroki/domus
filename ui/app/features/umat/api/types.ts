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
export type UmatSakramen = components["schemas"]["domain.UmatSakramenResponse"];
export type SearchUmatRequest =
  components["schemas"]["domain.SearchUmatRequest"];
export type CreateUmatRequest =
  components["schemas"]["domain.CreateUmatRequest"];
export type UpdateUmatRequest =
  components["schemas"]["domain.UpdateUmatRequest"];
export type UpsertUmatSakramenRequest =
  components["schemas"]["domain.UpsertUmatSakramenRequest"];
export type UpdateUmatSakramenRequest =
  components["schemas"]["domain.UpdateUmatSakramenRequest"];

export type Keuskupan = components["schemas"]["domain.KeuskupanResponse"];
export type Paroki = components["schemas"]["domain.ParokiResponse"];
export type ParokiMember = components["schemas"]["domain.ParokiMemberResponse"];
export type Wilayah = components["schemas"]["domain.WilayahResponse"];
export type Lingkungan = components["schemas"]["domain.LingkunganResponse"];
export type Kota = components["schemas"]["domain.KotaResponse"];
export type Kecamatan = components["schemas"]["domain.KecamatanResponse"];
export type Kelurahan = components["schemas"]["domain.KelurahanResponse"];
export type Lookup = components["schemas"]["domain.LookupResponse"];
