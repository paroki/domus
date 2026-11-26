import type {
  LookupUmatRequest,
  SearchOrganizationsRequest,
  SearchOrgInvitesRequest,
  SearchOrgJabatanRequest,
  SearchOrgMembersRequest,
  SearchOrgPeriodeRequest,
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
  periodeList: (
    orgId: string,
    params?: Omit<SearchOrgPeriodeRequest, "cursor">,
  ) => [...orgKeys.all, "periode", orgId, params ?? {}] as const,
  periodeDetail: (orgId: string, periodeId: string) =>
    [...orgKeys.all, "periode-detail", orgId, periodeId] as const,
  jabatanList: (
    orgId: string,
    params?: Omit<SearchOrgJabatanRequest, "cursor">,
  ) => [...orgKeys.all, "jabatan", orgId, params ?? {}] as const,
  jabatanDetail: (orgId: string, jabatanId: string) =>
    [...orgKeys.all, "jabatan-detail", orgId, jabatanId] as const,
  umatLookup: (params?: LookupUmatRequest) =>
    [...orgKeys.all, "umat-lookup", params ?? {}] as const,
  wilayah: () => [...orgKeys.all, "ref-wilayah"] as const,
  lingkungan: () => [...orgKeys.all, "ref-lingkungan"] as const,
  positionLookups: () => [...orgKeys.all, "ref-position-lookups"] as const,
  authProviders: () => ["auth", "providers"] as const,
};
