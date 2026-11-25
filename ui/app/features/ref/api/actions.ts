import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap, unwrapVoid } from "~/shared/api/client";
import { useSwitchParish } from "~/shared/api/session";
import { REF_ERRORS } from "./errors";
import { refKeys } from "./keys";
import type {
  AddParishMemberRequest,
  CreateDioceseRequest,
  CreateLookupRequest,
  CreateParishRequest,
  CreatePeriodRequest,
  CreateWardRequest,
  CreateZoneRequest,
  UpdateDioceseRequest,
  UpdateLookupRequest,
  UpdateParishMemberRequest,
  UpdateParishProfileRequest,
  UpdateParishRequest,
  UpdatePeriodRequest,
  UpdateWardRequest,
  UpdateZoneRequest,
} from "./types";

export function useActivateSessionParish() {
  const switchParish = useSwitchParish();
  return {
    ...switchParish,
    mutate: (parishId: string) =>
      switchParish.mutate({ activeParokiId: parishId }),
    mutateAsync: (parishId: string) =>
      switchParish.mutateAsync({ activeParokiId: parishId }),
  };
}

export function useCreateDiocese() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateDioceseRequest) =>
      unwrap(api.POST("/ref/keuskupan", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateDiocese() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateDioceseRequest }) =>
      unwrap(
        api.PATCH("/ref/keuskupan/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteDiocese() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/ref/keuskupan/{id}", { params: { path: { id } } }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useCreateParish() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateParishRequest) =>
      unwrap(api.POST("/ref/paroki", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateParish() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateParishRequest }) =>
      unwrap(
        api.PATCH("/ref/paroki/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateCurrentParishProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateParishProfileRequest) =>
      unwrap(api.PATCH("/ref/paroki/current", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteParish() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/ref/paroki/{id}", { params: { path: { id } } }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useAddParishMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      parishId,
      body,
    }: {
      parishId: string;
      body: AddParishMemberRequest;
    }) =>
      unwrap(
        api.POST("/ref/paroki/{id}/members", {
          params: { path: { id: parishId } },
          body,
        }),
        REF_ERRORS,
      ),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: refKeys.all }),
        qc.invalidateQueries({ queryKey: ["user", "me"] }),
      ]),
  });
}

export function useUpdateParishMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      parishId,
      userId,
      body,
    }: {
      parishId: string;
      userId: string;
      body: UpdateParishMemberRequest;
    }) =>
      unwrap(
        api.PATCH("/ref/paroki/{id}/members/{userId}", {
          params: { path: { id: parishId, userId } },
          body,
        }),
        REF_ERRORS,
      ),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: refKeys.all }),
        qc.invalidateQueries({ queryKey: ["user", "me"] }),
      ]),
  });
}

export function useRemoveParishMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ parishId, userId }: { parishId: string; userId: string }) =>
      unwrapVoid(
        api.DELETE("/ref/paroki/{id}/members/{userId}", {
          params: { path: { id: parishId, userId } },
        }),
        REF_ERRORS,
      ),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: refKeys.all }),
        qc.invalidateQueries({ queryKey: ["user", "me"] }),
      ]),
  });
}

export function useCreateZone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateZoneRequest) =>
      unwrap(api.POST("/ref/wilayah", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateZone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateZoneRequest }) =>
      unwrap(
        api.PATCH("/ref/wilayah/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteZone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/ref/wilayah/{id}", { params: { path: { id } } }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useCreateWard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateWardRequest) =>
      unwrap(api.POST("/ref/lingkungan", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateWard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateWardRequest }) =>
      unwrap(
        api.PATCH("/ref/lingkungan/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteWard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/ref/lingkungan/{id}", { params: { path: { id } } }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useCreateLookup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateLookupRequest) =>
      unwrap(api.POST("/ref/lookups", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateLookup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateLookupRequest }) =>
      unwrap(
        api.PATCH("/ref/lookups/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteLookup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/ref/lookups/{id}", { params: { path: { id } } }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useCreatePeriod() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreatePeriodRequest) =>
      unwrap(api.POST("/ref/periods", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdatePeriod() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdatePeriodRequest }) =>
      unwrap(
        api.PATCH("/ref/periods/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeletePeriod() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/ref/periods/{id}", { params: { path: { id } } }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}
