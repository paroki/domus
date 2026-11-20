import type {
  SearchOrganizationsRequest,
  SearchOrgInvitesRequest,
  SearchOrgMembersRequest,
} from "./types";

export const orgKeys = {
  all: ["org"] as const,
  orgs: (params?: Omit<SearchOrganizationsRequest, "cursor">) =>
    [...orgKeys.all, "list", params ?? {}] as const,
  org: (orgId: string) => [...orgKeys.all, "detail", orgId] as const,
  members: (orgId: string, params?: Omit<SearchOrgMembersRequest, "cursor">) =>
    [...orgKeys.all, "members", orgId, params ?? {}] as const,
  invites: (orgId: string, params?: Omit<SearchOrgInvitesRequest, "cursor">) =>
    [...orgKeys.all, "invites", orgId, params ?? {}] as const,
  invitePreview: (code: string) =>
    [...orgKeys.all, "invite-preview", code] as const,
  zones: () => [...orgKeys.all, "ref-zones"] as const,
  wards: () => [...orgKeys.all, "ref-wards"] as const,
  positionLookups: () => [...orgKeys.all, "ref-position-lookups"] as const,
  periods: () => [...orgKeys.all, "ref-periods"] as const,
  authProviders: () => ["auth", "providers"] as const,
};
