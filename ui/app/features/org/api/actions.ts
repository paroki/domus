import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, setActiveParish, unwrap, unwrapVoid } from "~/shared/api/client";
import { sessionKeys } from "~/shared/api/session";
import { orgKeys } from "./keys";
import type {
  AddOrgMemberRequest,
  CreateOrganizationRequest,
  CreateOrgInviteRequest,
  RequestOTPRequest,
  UpdateOrganizationRequest,
  UpdateOrgMemberRequest,
  VerifyOTPRequest,
} from "./types";

export function useCreateOrg() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateOrganizationRequest) =>
      unwrap(api.POST("/orgs", { body })),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useUpdateOrg(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateOrganizationRequest) =>
      unwrap(
        api.PATCH("/orgs/{id}", { params: { path: { id: orgId } }, body }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useDeleteOrg() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (orgId: string) =>
      unwrapVoid(api.DELETE("/orgs/{id}", { params: { path: { id: orgId } } })),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useAddOrgMember(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: AddOrgMemberRequest) =>
      unwrap(
        api.POST("/orgs/{id}/members", {
          params: { path: { id: orgId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useQuickSelfAssignOrgAdmin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, userId }: { orgId: string; userId: string }) =>
      unwrap(
        api.POST("/orgs/{id}/members", {
          params: { path: { id: orgId } },
          body: { userId, role: "admin" },
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useUpdateOrgMember(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      userId,
      body,
    }: {
      userId: string;
      body: UpdateOrgMemberRequest;
    }) =>
      unwrap(
        api.PATCH("/orgs/{id}/members/{userId}", {
          params: { path: { id: orgId, userId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useRemoveOrgMember(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) =>
      unwrapVoid(
        api.DELETE("/orgs/{id}/members/{userId}", {
          params: { path: { id: orgId, userId } },
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useCreateOrgInvite(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateOrgInviteRequest) =>
      unwrap(
        api.POST("/orgs/{id}/invites", {
          params: { path: { id: orgId } },
          body,
        }),
      ),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: [...orgKeys.all, "invites", orgId] }),
  });
}

export function useRevokeOrgInvite(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (inviteId: string) =>
      unwrapVoid(
        api.DELETE("/orgs/{id}/invites/{inviteId}", {
          params: { path: { id: orgId, inviteId } },
        }),
      ),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: [...orgKeys.all, "invites", orgId] }),
  });
}

export function useRedeemOrgInvite(code: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      unwrap(
        api.POST("/orgs/invites/{code}/redeem", {
          params: { path: { code } },
        }),
      ),
    onSuccess: async (res) => {
      setActiveParish(res.parishId ?? null);
      await Promise.all([
        qc.invalidateQueries({ queryKey: sessionKeys.me }),
        qc.invalidateQueries({ queryKey: sessionKeys.parishes() }),
        qc.invalidateQueries({ queryKey: orgKeys.all }),
      ]);
    },
  });
}

export function useOrgRequestOTP() {
  return useMutation({
    mutationFn: (body: RequestOTPRequest) =>
      unwrap(api.POST("/auth/otp", { body })),
  });
}

export function useOrgVerifyOTP() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: VerifyOTPRequest) =>
      unwrap(api.POST("/auth/otp/verify", { body })),
    onSuccess: async (res) => {
      const parishId = res.activeParishId ?? null;
      if (parishId) {
        setActiveParish(parishId);
      }
      await Promise.all([
        qc.invalidateQueries({ queryKey: sessionKeys.me }),
        qc.invalidateQueries({ queryKey: sessionKeys.parishes() }),
        qc.invalidateQueries({ queryKey: orgKeys.all }),
      ]);
    },
  });
}
