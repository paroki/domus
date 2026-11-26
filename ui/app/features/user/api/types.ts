import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type RoleDescriptor = Schemas["authz.RoleDescriptor"];
export type User = Schemas["domain.UserResponse"];
export type AuthProviders = Schemas["domain.AuthProvidersResponse"];
export type Paroki = Schemas["domain.ParokiResponse"];
export type ParokiMember = Schemas["domain.ParokiMemberResponse"];
export type AddParokiMemberRequest = Schemas["domain.AddParokiMemberRequest"];
export type UpdateParokiMemberRequest =
  Schemas["domain.UpdateParokiMemberRequest"];
export type SearchParokiMembersRequest =
  Schemas["domain.SearchParokiMembersRequest"];
export type Organization = Schemas["domain.OrganizationResponse"];
export type OrgMember = Schemas["domain.OrgMemberResponse"];
export type AddOrgMemberRequest = Schemas["domain.AddOrgMemberRequest"];
export type UpdateOrgMemberRequest = Schemas["domain.UpdateOrgMemberRequest"];
export type SearchOrgMembersRequest = Schemas["domain.SearchOrgMembersRequest"];
export type Lookup = Schemas["domain.LookupResponse"];
export type AuthSession = Schemas["domain.AuthSessionResponse"];
export type RequestOTPRequest = Schemas["domain.RequestOTPRequest"];
export type RequestOTPResponse = Schemas["domain.RequestOTPResponse"];
export type VerifyOTPRequest = Schemas["domain.VerifyOTPRequest"];
export type UpdateMeRequest = Schemas["domain.UpdateMeRequest"];
export type SwitchSessionParokiRequest =
  Schemas["domain.SwitchSessionParokiRequest"];
export type UserSession = Schemas["domain.UserSessionResponse"];
export type SearchSessionsRequest =
  Schemas["github_com_paroki_domus_internal_user_domain.SearchSessionsRequest"];
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
export type UserToken = Schemas["domain.UserTokenResponse"];
export type CreateUserTokenRequest = Schemas["domain.CreateUserTokenRequest"];
export type CreateUserTokenResponse = Schemas["domain.CreateUserTokenResponse"];
export type SearchTokensRequest = Schemas["domain.SearchTokensRequest"];
