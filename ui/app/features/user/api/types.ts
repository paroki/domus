import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type User = Schemas["domain.UserResponse"];
export type AuthProviders = Schemas["domain.AuthProvidersResponse"];
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
