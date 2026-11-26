import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type Keuskupan = Schemas["domain.KeuskupanResponse"];
export type SearchKeuskupanRequest = Schemas["domain.SearchKeuskupanRequest"];
export type CreateKeuskupanRequest = Schemas["domain.CreateKeuskupanRequest"];
export type UpdateKeuskupanRequest = Schemas["domain.UpdateKeuskupanRequest"];

export type Kevikepan = Schemas["domain.KevikepanResponse"];
export type SearchKevikepanRequest = Schemas["domain.SearchKevikepanRequest"];
export type CreateKevikepanRequest = Schemas["domain.CreateKevikepanRequest"];
export type UpdateKevikepanRequest = Schemas["domain.UpdateKevikepanRequest"];

export type Kota = Schemas["domain.KotaResponse"];
export type SearchKotaRequest = Schemas["domain.SearchKotaRequest"];
export type CreateKotaRequest = Schemas["domain.CreateKotaRequest"];
export type UpdateKotaRequest = Schemas["domain.UpdateKotaRequest"];

export type Kecamatan = Schemas["domain.KecamatanResponse"];
export type SearchKecamatanRequest = Schemas["domain.SearchKecamatanRequest"];
export type CreateKecamatanRequest = Schemas["domain.CreateKecamatanRequest"];
export type UpdateKecamatanRequest = Schemas["domain.UpdateKecamatanRequest"];

export type Kelurahan = Schemas["domain.KelurahanResponse"];
export type SearchKelurahanRequest = Schemas["domain.SearchKelurahanRequest"];
export type CreateKelurahanRequest = Schemas["domain.CreateKelurahanRequest"];
export type UpdateKelurahanRequest = Schemas["domain.UpdateKelurahanRequest"];

export type Paroki = Schemas["domain.ParokiResponse"];
export type SearchParokiRequest = Schemas["domain.SearchParokiRequest"];
export type CreateParokiRequest = Schemas["domain.CreateParokiRequest"];
export type UpdateParokiRequest = Schemas["domain.UpdateParokiRequest"];
export type UpdateParokiProfileRequest =
  Schemas["domain.UpdateParokiProfileRequest"];

export type ParokiMember = Schemas["domain.ParokiMemberResponse"];
export type SearchParokiMembersRequest =
  Schemas["domain.SearchParokiMembersRequest"];
export type AddParokiMemberRequest = Schemas["domain.AddParokiMemberRequest"];
export type UpdateParokiMemberRequest =
  Schemas["domain.UpdateParokiMemberRequest"];

export type Wilayah = Schemas["domain.WilayahResponse"];
export type SearchWilayahRequest = Schemas["domain.SearchWilayahRequest"];
export type CreateWilayahRequest = Schemas["domain.CreateWilayahRequest"];
export type UpdateWilayahRequest = Schemas["domain.UpdateWilayahRequest"];

export type Lingkungan = Schemas["domain.LingkunganResponse"];
export type SearchLingkunganRequest = Schemas["domain.SearchLingkunganRequest"];
export type CreateLingkunganRequest = Schemas["domain.CreateLingkunganRequest"];
export type UpdateLingkunganRequest = Schemas["domain.UpdateLingkunganRequest"];

export type Lookup = Schemas["domain.LookupResponse"];
export type SearchLookupsRequest = Schemas["domain.SearchLookupsRequest"];
export type CreateLookupRequest = Schemas["domain.CreateLookupRequest"];
export type UpdateLookupRequest = Schemas["domain.UpdateLookupRequest"];
