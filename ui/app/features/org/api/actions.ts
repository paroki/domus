import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, setActiveParoki, unwrap, unwrapVoid } from "~/shared/api/client";
import { sessionKeys } from "~/shared/api/session";
import { orgKeys } from "./keys";
import type {
  AddOrgMemberRequest,
  CreateOrganizationRequest,
  CreateOrgInviteRequest,
  CreateOrgJabatanRequest,
  CreateOrgPeriodeRequest,
  RequestOTPRequest,
  UpdateOrganizationRequest,
  UpdateOrgJabatanRequest,
  UpdateOrgMemberRequest,
  UpdateOrgPeriodeRequest,
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
      setActiveParoki(res.parokiId ?? null);
      await Promise.all([
        qc.invalidateQueries({ queryKey: sessionKeys.me }),
        qc.invalidateQueries({ queryKey: sessionKeys.paroki() }),
        qc.invalidateQueries({ queryKey: orgKeys.all }),
      ]);
    },
  });
}

export function useCreateOrgPeriode(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateOrgPeriodeRequest) =>
      unwrap(
        api.POST("/orgs/{id}/periode", {
          params: { path: { id: orgId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useUpdateOrgPeriode(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      periodeId,
      body,
    }: {
      periodeId: string;
      body: UpdateOrgPeriodeRequest;
    }) =>
      unwrap(
        api.PATCH("/orgs/{id}/periode/{periodeId}", {
          params: { path: { id: orgId, periodeId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useDeleteOrgPeriode(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (periodeId: string) =>
      unwrapVoid(
        api.DELETE("/orgs/{id}/periode/{periodeId}", {
          params: { path: { id: orgId, periodeId } },
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useCreateOrgJabatan(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateOrgJabatanRequest) =>
      unwrap(
        api.POST("/orgs/{id}/jabatan", {
          params: { path: { id: orgId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useUpdateOrgJabatan(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      jabatanId,
      body,
    }: {
      jabatanId: string;
      body: UpdateOrgJabatanRequest;
    }) =>
      unwrap(
        api.PATCH("/orgs/{id}/jabatan/{jabatanId}", {
          params: { path: { id: orgId, jabatanId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
  });
}

export function useDeleteOrgJabatan(orgId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (jabatanId: string) =>
      unwrapVoid(
        api.DELETE("/orgs/{id}/jabatan/{jabatanId}", {
          params: { path: { id: orgId, jabatanId } },
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: orgKeys.all }),
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
      const parokiId = res.activeParokiId ?? null;
      if (parokiId) {
        setActiveParoki(parokiId);
      }
      await Promise.all([
        qc.invalidateQueries({ queryKey: sessionKeys.me }),
        qc.invalidateQueries({ queryKey: sessionKeys.paroki() }),
        qc.invalidateQueries({ queryKey: orgKeys.all }),
      ]);
    },
  });
}
