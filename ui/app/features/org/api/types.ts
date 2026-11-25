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
export type PeriodResponse = components["schemas"]["domain.PeriodResponse"];

export type RequestOTPRequest =
  components["schemas"]["domain.RequestOTPRequest"];
export type VerifyOTPRequest = components["schemas"]["domain.VerifyOTPRequest"];
export type AuthSessionResponse =
  components["schemas"]["domain.AuthSessionResponse"];
