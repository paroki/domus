import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type LiturgiMinistry = Schemas["domain.LiturgiMinistryResponse"];
export type UpdateLiturgiMinistryRequest =
  Schemas["domain.UpdateLiturgiMinistryRequest"];

export type LiturgiSchedule = Schemas["domain.LiturgiScheduleResponse"];
export type LiturgiPetugas = Schemas["domain.LiturgiPetugasResponse"];
export type LiturgiLagu = Schemas["domain.LiturgiLaguResponse"];
export type LiturgiRSVP = Schemas["domain.LiturgiRSVPResponse"];
export type LiturgiRSVPSummary = Schemas["domain.LiturgiRSVPSummaryResponse"];

export type SearchLiturgiSchedulesRequest =
  Schemas["domain.SearchLiturgiSchedulesRequest"];
export type LiturgiScheduleSearchFilter =
  Schemas["domain.LiturgiScheduleSearchFilter"];
export type CreateLiturgiScheduleRequest =
  Schemas["domain.CreateLiturgiScheduleRequest"];
export type UpdateLiturgiScheduleRequest =
  Schemas["domain.UpdateLiturgiScheduleRequest"];

export type CreateLiturgiPetugasRequest =
  Schemas["domain.CreateLiturgiPetugasRequest"];
export type UpdateLiturgiPetugasRequest =
  Schemas["domain.UpdateLiturgiPetugasRequest"];
export type RespondLiturgiPetugasRequest =
  Schemas["domain.RespondLiturgiPetugasRequest"];
export type ApproveLiturgiSubstituteRequest =
  Schemas["domain.ApproveLiturgiSubstituteRequest"];

export type CreateLiturgiLaguRequest =
  Schemas["domain.CreateLiturgiLaguRequest"];
export type UpdateLiturgiLaguRequest =
  Schemas["domain.UpdateLiturgiLaguRequest"];

export type UpsertEventRSVPRequest = Schemas["domain.UpsertEventRSVPRequest"];

export type Paroki = Schemas["domain.ParokiResponse"];
export type ParokiMember = Schemas["domain.ParokiMemberResponse"];
export type Organization = Schemas["domain.OrganizationResponse"];
export type Lookup = Schemas["domain.LookupResponse"];
