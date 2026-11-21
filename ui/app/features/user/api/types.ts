import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type RoleDescriptor = Schemas["authz.RoleDescriptor"];
export type User = Schemas["domain.UserResponse"];
export type AuthProviders = Schemas["domain.AuthProvidersResponse"];
export type Parish = Schemas["domain.ParishResponse"];
export type ParishMember = Schemas["domain.ParishMemberResponse"];
export type AddParishMemberRequest = Schemas["domain.AddParishMemberRequest"];
export type UpdateParishMemberRequest =
  Schemas["domain.UpdateParishMemberRequest"];
export type SearchParishMembersRequest =
  Schemas["domain.SearchParishMembersRequest"];
export type Organization = Schemas["domain.OrganizationResponse"];
export type OrgMember = Schemas["domain.OrgMemberResponse"];
export type AddOrgMemberRequest = Schemas["domain.AddOrgMemberRequest"];
export type UpdateOrgMemberRequest = Schemas["domain.UpdateOrgMemberRequest"];
export type SearchOrgMembersRequest = Schemas["domain.SearchOrgMembersRequest"];
export type Lookup = Schemas["domain.LookupResponse"];
export type Period = Schemas["domain.PeriodResponse"];
export type AuthSession = Schemas["domain.AuthSessionResponse"];
export type RequestOTPRequest = Schemas["domain.RequestOTPRequest"];
export type RequestOTPResponse = Schemas["domain.RequestOTPResponse"];
export type VerifyOTPRequest = Schemas["domain.VerifyOTPRequest"];
export type UpdateMeRequest = Schemas["domain.UpdateMeRequest"];
export type SwitchSessionParishRequest =
  Schemas["domain.SwitchSessionParishRequest"];
export type UserSession = Schemas["domain.UserSessionResponse"];
export type SearchSessionsRequest = Schemas["domain.SearchSessionsRequest"];
export type UserAccount = Schemas["domain.UserAccountResponse"];
export type SearchAccountsRequest = Schemas["domain.SearchAccountsRequest"];
export type SearchUsersRequest = Schemas["domain.SearchUsersRequest"];
export type AdminCreateUserRequest = Schemas["domain.AdminCreateUserRequest"];
export type AdminUpdateUserRequest = Schemas["domain.AdminUpdateUserRequest"];
export type LinkAccountOTPRequest = Schemas["domain.LinkAccountOTPRequest"];
export type VerifyLinkAccountRequest =
  Schemas["domain.VerifyLinkAccountRequest"];
export type TelegramLinkTokenResponse =
  Schemas["domain.TelegramLinkTokenResponse"];
