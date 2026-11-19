import type {
  SearchAccountsRequest,
  SearchSessionsRequest,
  SearchUsersRequest,
} from "./types";

export const userKeys = {
  all: ["user"] as const,
  providers: () => [...userKeys.all, "providers"] as const,
  me: () => [...userKeys.all, "me"] as const,
  sessions: (params?: SearchSessionsRequest) =>
    [...userKeys.all, "sessions", params ?? {}] as const,
  accounts: (params?: SearchAccountsRequest) =>
    [...userKeys.all, "accounts", params ?? {}] as const,
  list: (params?: Omit<SearchUsersRequest, "cursor">) =>
    [...userKeys.all, "list", params ?? {}] as const,
  detail: (id: string) => [...userKeys.all, "detail", id] as const,
};
