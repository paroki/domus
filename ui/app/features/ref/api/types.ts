import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type Diocese = Schemas["domain.DioceseResponse"];
export type SearchDiocesesRequest = Schemas["domain.SearchDiocesesRequest"];
export type CreateDioceseRequest = Schemas["domain.CreateDioceseRequest"];
export type UpdateDioceseRequest = Schemas["domain.UpdateDioceseRequest"];

export type Parish = Schemas["domain.ParishResponse"];
export type SearchParishesRequest = Schemas["domain.SearchParishesRequest"];
export type CreateParishRequest = Schemas["domain.CreateParishRequest"];
export type UpdateParishRequest = Schemas["domain.UpdateParishRequest"];
export type UpdateParishProfileRequest =
  Schemas["domain.UpdateParishProfileRequest"];

export type ParishMember = Schemas["domain.ParishMemberResponse"];
export type SearchParishMembersRequest =
  Schemas["domain.SearchParishMembersRequest"];
export type AddParishMemberRequest = Schemas["domain.AddParishMemberRequest"];
export type UpdateParishMemberRequest =
  Schemas["domain.UpdateParishMemberRequest"];

export type Zone = Schemas["domain.ZoneResponse"];
export type SearchZonesRequest = Schemas["domain.SearchZonesRequest"];
export type CreateZoneRequest = Schemas["domain.CreateZoneRequest"];
export type UpdateZoneRequest = Schemas["domain.UpdateZoneRequest"];

export type Ward = Schemas["domain.WardResponse"];
export type SearchWardsRequest = Schemas["domain.SearchWardsRequest"];
export type CreateWardRequest = Schemas["domain.CreateWardRequest"];
export type UpdateWardRequest = Schemas["domain.UpdateWardRequest"];

export type Lookup = Schemas["domain.LookupResponse"];
export type SearchLookupsRequest = Schemas["domain.SearchLookupsRequest"];
export type CreateLookupRequest = Schemas["domain.CreateLookupRequest"];
export type UpdateLookupRequest = Schemas["domain.UpdateLookupRequest"];

export type Period = Schemas["domain.PeriodResponse"];
export type SearchPeriodsRequest = Schemas["domain.SearchPeriodsRequest"];
export type CreatePeriodRequest = Schemas["domain.CreatePeriodRequest"];
export type UpdatePeriodRequest = Schemas["domain.UpdatePeriodRequest"];
