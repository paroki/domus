import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type Diocese = Schemas["domain.KeuskupanResponse"];
export type SearchDiocesesRequest = Schemas["domain.SearchKeuskupanRequest"];
export type CreateDioceseRequest = Schemas["domain.CreateKeuskupanRequest"];
export type UpdateDioceseRequest = Schemas["domain.UpdateKeuskupanRequest"];

export type Parish = Schemas["domain.ParokiResponse"];
export type SearchParishesRequest = Schemas["domain.SearchParokiRequest"];
export type CreateParishRequest = Schemas["domain.CreateParokiRequest"];
export type UpdateParishRequest = Schemas["domain.UpdateParokiRequest"];
export type UpdateParishProfileRequest =
  Schemas["domain.UpdateParokiProfileRequest"];

export type ParishMember = Schemas["domain.ParokiMemberResponse"];
export type SearchParishMembersRequest =
  Schemas["domain.SearchParokiMembersRequest"];
export type AddParishMemberRequest = Schemas["domain.AddParokiMemberRequest"];
export type UpdateParishMemberRequest =
  Schemas["domain.UpdateParokiMemberRequest"];

export type Zone = Schemas["domain.WilayahResponse"];
export type SearchZonesRequest = Schemas["domain.SearchWilayahRequest"];
export type CreateZoneRequest = Schemas["domain.CreateWilayahRequest"];
export type UpdateZoneRequest = Schemas["domain.UpdateWilayahRequest"];

export type Ward = Schemas["domain.LingkunganResponse"];
export type SearchWardsRequest = Schemas["domain.SearchLingkunganRequest"];
export type CreateWardRequest = Schemas["domain.CreateLingkunganRequest"];
export type UpdateWardRequest = Schemas["domain.UpdateLingkunganRequest"];

export type Lookup = Schemas["domain.LookupResponse"];
export type SearchLookupsRequest = Schemas["domain.SearchLookupsRequest"];
export type CreateLookupRequest = Schemas["domain.CreateLookupRequest"];
export type UpdateLookupRequest = Schemas["domain.UpdateLookupRequest"];

export type Period = Schemas["domain.PeriodResponse"];
export type SearchPeriodsRequest = Schemas["domain.SearchPeriodsRequest"];
export type CreatePeriodRequest = Schemas["domain.CreatePeriodRequest"];
export type UpdatePeriodRequest = Schemas["domain.UpdatePeriodRequest"];
