import type { components } from "~/shared/api/schema";

export type OrganizationResponse =
  components["schemas"]["domain.OrganizationResponse"];
export type SearchOrganizationsRequest =
  components["schemas"]["domain.SearchOrganizationsRequest"];
export type OrganizationSearchFilter =
  components["schemas"]["domain.OrganizationSearchFilter"];
export type CreateOrganizationRequest =
  components["schemas"]["domain.CreateOrganizationRequest"];
export type UpdateOrganizationRequest =
  components["schemas"]["domain.UpdateOrganizationRequest"];

export type OrgMemberResponse =
  components["schemas"]["domain.OrgMemberResponse"];
export type SearchOrgMembersRequest =
  components["schemas"]["domain.SearchOrgMembersRequest"];
export type OrgMemberSearchFilter =
  components["schemas"]["domain.OrgMemberSearchFilter"];
export type AddOrgMemberRequest =
  components["schemas"]["domain.AddOrgMemberRequest"];
export type UpdateOrgMemberRequest =
  components["schemas"]["domain.UpdateOrgMemberRequest"];

export type OrgInviteResponse =
  components["schemas"]["domain.OrgInviteResponse"];
export type SearchOrgInvitesRequest =
  components["schemas"]["domain.SearchOrgInvitesRequest"];
export type OrgInviteSearchFilter =
  components["schemas"]["domain.OrgInviteSearchFilter"];
export type CreateOrgInviteRequest =
  components["schemas"]["domain.CreateOrgInviteRequest"];
export type OrgInvitePreviewResponse =
  components["schemas"]["domain.OrgInvitePreviewResponse"];
export type RedeemOrgInviteResponse =
  components["schemas"]["domain.RedeemOrgInviteResponse"];

export type WilayahResponse = components["schemas"]["domain.WilayahResponse"];
export type LingkunganResponse =
  components["schemas"]["domain.LingkunganResponse"];
export type LookupResponse = components["schemas"]["domain.LookupResponse"];

export type OrgPeriodeResponse =
  components["schemas"]["domain.OrgPeriodeResponse"];
export type OrgPeriodeSearchFilter =
  components["schemas"]["domain.OrgPeriodeSearchFilter"];
export type SearchOrgPeriodeRequest =
  components["schemas"]["domain.SearchOrgPeriodeRequest"];
export type CreateOrgPeriodeRequest =
  components["schemas"]["domain.CreateOrgPeriodeRequest"];
export type UpdateOrgPeriodeRequest =
  components["schemas"]["domain.UpdateOrgPeriodeRequest"];

export type OrgJabatanResponse =
  components["schemas"]["domain.OrgJabatanResponse"];
export type OrgJabatanSearchFilter =
  components["schemas"]["domain.OrgJabatanSearchFilter"];
export type SearchOrgJabatanRequest =
  components["schemas"]["domain.SearchOrgJabatanRequest"];
export type CreateOrgJabatanRequest =
  components["schemas"]["domain.CreateOrgJabatanRequest"];
export type UpdateOrgJabatanRequest =
  components["schemas"]["domain.UpdateOrgJabatanRequest"];

export type UmatLookupItemResponse =
  components["schemas"]["domain.UmatLookupItemResponse"];
export type LookupUmatRequest =
  components["schemas"]["domain.LookupUmatRequest"];

export type RequestOTPRequest =
  components["schemas"]["domain.RequestOTPRequest"];
export type VerifyOTPRequest = components["schemas"]["domain.VerifyOTPRequest"];
export type AuthSessionResponse =
  components["schemas"]["domain.AuthSessionResponse"];
