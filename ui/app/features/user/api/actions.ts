import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, setActiveParoki, unwrap, unwrapVoid } from "~/shared/api/client";
import { useLogout, useSwitchParoki } from "~/shared/api/session";
import { USER_ERRORS } from "./errors";
import { userKeys } from "./keys";
import type {
  AddOrgMemberRequest,
  AddParokiMemberRequest,
  AdminCreateUserRequest,
  AdminUpdateUserRequest,
  CreateUserTokenRequest,
  LinkAccountOTPRequest,
  RequestOTPRequest,
  UpdateMeRequest,
  UpdateOrgMemberRequest,
  UpdateParokiMemberRequest,
  VerifyLinkAccountRequest,
  VerifyOTPRequest,
} from "./types";

export { useLogout, useSwitchParoki };

export function useRequestOTP() {
  return useMutation({
    mutationFn: (body: RequestOTPRequest) =>
      unwrap(api.POST("/auth/otp", { body }), USER_ERRORS),
  });
}

export function useVerifyOTP() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: VerifyOTPRequest) =>
      unwrap(api.POST("/auth/otp/verify", { body }), USER_ERRORS),
    onSuccess: async (data) => {
      setActiveParoki(data.activeParokiId ?? null);
      await qc.invalidateQueries({ queryKey: userKeys.me() });
    },
  });
}

export function useUpdateMe() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateMeRequest) =>
      unwrap(api.PATCH("/me", { body }), USER_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: userKeys.me() }),
  });
}

export function useRevokeSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/me/sessions/{id}", { params: { path: { id } } }),
        USER_ERRORS,
      ),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.sessions() }),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useCreateToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateUserTokenRequest) =>
      unwrap(api.POST("/me/tokens", { body }), USER_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: userKeys.tokens() }),
  });
}

export function useRevokeToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/me/tokens/{id}", { params: { path: { id } } }),
        USER_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: userKeys.tokens() }),
  });
}

export function useUnlinkAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/me/accounts/{id}", { params: { path: { id } } }),
        USER_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: userKeys.accounts() }),
  });
}

export function useRequestLinkAccountOTP() {
  return useMutation({
    mutationFn: (body: LinkAccountOTPRequest) =>
      unwrap(api.POST("/me/accounts/otp", { body }), USER_ERRORS),
  });
}

export function useVerifyLinkAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: VerifyLinkAccountRequest) =>
      unwrap(api.POST("/me/accounts/verify", { body }), USER_ERRORS),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.accounts() }),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useCreateTelegramLinkToken() {
  return useMutation({
    mutationFn: () =>
      unwrap(api.POST("/me/accounts/telegram/token"), USER_ERRORS),
  });
}

export function useAdminCreateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: AdminCreateUserRequest) =>
      unwrap(api.POST("/users", { body }), USER_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: userKeys.list() }),
  });
}

export function useAdminUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: AdminUpdateUserRequest }) =>
      unwrap(
        api.PATCH("/users/{id}", { params: { path: { id } }, body }),
        USER_ERRORS,
      ),
    onSuccess: (_data, { id }) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.list() }),
        qc.invalidateQueries({ queryKey: userKeys.detail(id) }),
      ]),
  });
}

export function useAdminDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/users/{id}", { params: { path: { id } } }),
        USER_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: userKeys.list() }),
  });
}

export function useAddParokiMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      parokiId,
      body,
    }: {
      parokiId: string;
      body: AddParokiMemberRequest;
    }) =>
      unwrap(
        api.POST("/ref/paroki/{id}/members", {
          params: { path: { id: parokiId } },
          body,
        }),
        USER_ERRORS,
      ),
    onSuccess: (_data, { parokiId }) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.parokiMembers(parokiId) }),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useUpdateParokiMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      parokiId,
      userId,
      body,
    }: {
      parokiId: string;
      userId: string;
      body: UpdateParokiMemberRequest;
    }) =>
      unwrap(
        api.PATCH("/ref/paroki/{id}/members/{userId}", {
          params: { path: { id: parokiId, userId } },
          body,
        }),
        USER_ERRORS,
      ),
    onSuccess: (_data, { parokiId }) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.parokiMembers(parokiId) }),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useRemoveParokiMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ parokiId, userId }: { parokiId: string; userId: string }) =>
      unwrapVoid(
        api.DELETE("/ref/paroki/{id}/members/{userId}", {
          params: { path: { id: parokiId, userId } },
        }),
        USER_ERRORS,
      ),
    onSuccess: (_data, { parokiId }) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.parokiMembers(parokiId) }),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useDirectoryAddOrgMember(orgId: string, parishId?: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: AddOrgMemberRequest) =>
      unwrap(
        api.POST("/orgs/{id}/members", {
          params: { path: { id: orgId } },
          body,
        }),
        USER_ERRORS,
      ),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.orgMembers(orgId) }),
        ...(parishId
          ? [qc.invalidateQueries({ queryKey: userKeys.orgs(parishId) })]
          : []),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useDirectoryUpdateOrgMember(orgId: string, parishId?: string) {
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
        USER_ERRORS,
      ),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.orgMembers(orgId) }),
        ...(parishId
          ? [qc.invalidateQueries({ queryKey: userKeys.orgs(parishId) })]
          : []),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useDirectoryRemoveOrgMember(orgId: string, parishId?: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) =>
      unwrapVoid(
        api.DELETE("/orgs/{id}/members/{userId}", {
          params: { path: { id: orgId, userId } },
        }),
        USER_ERRORS,
      ),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.orgMembers(orgId) }),
        ...(parishId
          ? [qc.invalidateQueries({ queryKey: userKeys.orgs(parishId) })]
          : []),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useAddUserOrgMember(parishId?: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      orgId,
      body,
    }: {
      orgId: string;
      body: AddOrgMemberRequest;
    }) =>
      unwrap(
        api.POST("/orgs/{id}/members", {
          params: { path: { id: orgId } },
          body,
        }),
        USER_ERRORS,
      ),
    onSuccess: (_data, { orgId }) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.orgMembers(orgId) }),
        ...(parishId
          ? [qc.invalidateQueries({ queryKey: userKeys.orgs(parishId) })]
          : []),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useUpdateUserOrgMember(parishId?: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      orgId,
      userId,
      body,
    }: {
      orgId: string;
      userId: string;
      body: UpdateOrgMemberRequest;
    }) =>
      unwrap(
        api.PATCH("/orgs/{id}/members/{userId}", {
          params: { path: { id: orgId, userId } },
          body,
        }),
        USER_ERRORS,
      ),
    onSuccess: (_data, { orgId }) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.orgMembers(orgId) }),
        ...(parishId
          ? [qc.invalidateQueries({ queryKey: userKeys.orgs(parishId) })]
          : []),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useRemoveUserOrgMember(parishId?: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, userId }: { orgId: string; userId: string }) =>
      unwrapVoid(
        api.DELETE("/orgs/{id}/members/{userId}", {
          params: { path: { id: orgId, userId } },
        }),
        USER_ERRORS,
      ),
    onSuccess: (_data, { orgId }) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.orgMembers(orgId) }),
        ...(parishId
          ? [qc.invalidateQueries({ queryKey: userKeys.orgs(parishId) })]
          : []),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}
