import type {
  SearchAccountsRequest,
  SearchOrgMembersRequest,
  SearchParokiMembersRequest,
  SearchSessionsRequest,
  SearchTokensRequest,
  SearchUsersRequest,
} from "./types";

export const userKeys = {
  all: ["user"] as const,
  providers: () => [...userKeys.all, "providers"] as const,
  roles: () => [...userKeys.all, "roles"] as const,
  me: () => [...userKeys.all, "me"] as const,
  sessions: (params?: SearchSessionsRequest) =>
    [...userKeys.all, "sessions", params ?? {}] as const,
  tokens: (params?: SearchTokensRequest) =>
    [...userKeys.all, "tokens", params ?? {}] as const,
  accounts: (params?: SearchAccountsRequest) =>
    [...userKeys.all, "accounts", params ?? {}] as const,
  list: (params?: Omit<SearchUsersRequest, "cursor">) =>
    [...userKeys.all, "list", params ?? {}] as const,
  detail: (id: string) => [...userKeys.all, "detail", id] as const,
  paroki: (q?: string) => [...userKeys.all, "paroki", q ?? ""] as const,
  parokiMembers: (
    parokiId: string,
    params?: Omit<SearchParokiMembersRequest, "cursor">,
  ) => [...userKeys.all, "parokiMembers", parokiId, params ?? {}] as const,
  orgs: (parokiId: string) => [...userKeys.all, "orgs", parokiId] as const,
  orgMembers: (
    orgId: string,
    params?: Omit<SearchOrgMembersRequest, "cursor">,
  ) => [...userKeys.all, "orgMembers", orgId, params ?? {}] as const,
  lookups: (category: string) =>
    [...userKeys.all, "lookups", category] as const,
};
