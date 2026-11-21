import type {
  SearchAccountsRequest,
  SearchOrgMembersRequest,
  SearchParishMembersRequest,
  SearchSessionsRequest,
  SearchUsersRequest,
} from "./types";

export const userKeys = {
  all: ["user"] as const,
  providers: () => [...userKeys.all, "providers"] as const,
  roles: () => [...userKeys.all, "roles"] as const,
  me: () => [...userKeys.all, "me"] as const,
  sessions: (params?: SearchSessionsRequest) =>
    [...userKeys.all, "sessions", params ?? {}] as const,
  accounts: (params?: SearchAccountsRequest) =>
    [...userKeys.all, "accounts", params ?? {}] as const,
  list: (params?: Omit<SearchUsersRequest, "cursor">) =>
    [...userKeys.all, "list", params ?? {}] as const,
  detail: (id: string) => [...userKeys.all, "detail", id] as const,
  parishes: (q?: string) => [...userKeys.all, "parishes", q ?? ""] as const,
  parishMembers: (
    parishId: string,
    params?: Omit<SearchParishMembersRequest, "cursor">,
  ) => [...userKeys.all, "parishMembers", parishId, params ?? {}] as const,
  orgs: (parishId: string) => [...userKeys.all, "orgs", parishId] as const,
  orgMembers: (
    orgId: string,
    params?: Omit<SearchOrgMembersRequest, "cursor">,
  ) => [...userKeys.all, "orgMembers", orgId, params ?? {}] as const,
  lookups: (category: string) =>
    [...userKeys.all, "lookups", category] as const,
  periods: (parishId: string) =>
    [...userKeys.all, "periods", parishId] as const,
};
