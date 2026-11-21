import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, setActiveParish, unwrap } from "~/shared/api/client";
import { useLogout, useSwitchParish } from "~/shared/api/session";
import { USER_ERRORS } from "./errors";
import { userKeys } from "./keys";
import type {
  AdminCreateUserRequest,
  AdminUpdateUserRequest,
  LinkAccountOTPRequest,
  RequestOTPRequest,
  UpdateMeRequest,
  VerifyLinkAccountRequest,
  VerifyOTPRequest,
} from "./types";

export { useLogout, useSwitchParish };

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
      setActiveParish(data.activeParishId ?? null);
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
    mutationFn: async (id: string) => {
      await unwrap(
        api.DELETE("/me/sessions/{id}", { params: { path: { id } } }),
        USER_ERRORS,
      );
    },
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.sessions() }),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]),
  });
}

export function useUnlinkAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await unwrap(
        api.DELETE("/me/accounts/{id}", { params: { path: { id } } }),
        USER_ERRORS,
      );
    },
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
    mutationFn: async (id: string) => {
      await unwrap(
        api.DELETE("/users/{id}", { params: { path: { id } } }),
        USER_ERRORS,
      );
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: userKeys.list() }),
  });
}
