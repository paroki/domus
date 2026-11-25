import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap, unwrapVoid } from "~/shared/api/client";
import { useSwitchParoki } from "~/shared/api/session";
import { REF_ERRORS } from "./errors";
import { refKeys } from "./keys";
import type {
  AddParokiMemberRequest,
  CreateKecamatanRequest,
  CreateKelurahanRequest,
  CreateKeuskupanRequest,
  CreateKevikepanRequest,
  CreateKotaRequest,
  CreateLingkunganRequest,
  CreateLookupRequest,
  CreateParokiRequest,
  CreatePeriodRequest,
  CreateWilayahRequest,
  UpdateKecamatanRequest,
  UpdateKelurahanRequest,
  UpdateKeuskupanRequest,
  UpdateKevikepanRequest,
  UpdateKotaRequest,
  UpdateLingkunganRequest,
  UpdateLookupRequest,
  UpdateParokiMemberRequest,
  UpdateParokiProfileRequest,
  UpdateParokiRequest,
  UpdatePeriodRequest,
  UpdateWilayahRequest,
} from "./types";

export function useActivateSessionParoki() {
  const switchParoki = useSwitchParoki();
  return {
    ...switchParoki,
    mutate: (parokiId: string) =>
      switchParoki.mutate({ activeParokiId: parokiId }),
    mutateAsync: (parokiId: string) =>
      switchParoki.mutateAsync({ activeParokiId: parokiId }),
  };
}

export function useCreateKeuskupan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateKeuskupanRequest) =>
      unwrap(api.POST("/ref/keuskupan", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateKeuskupan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateKeuskupanRequest }) =>
      unwrap(
        api.PATCH("/ref/keuskupan/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteKeuskupan() {
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

export function useCreateParoki() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateParokiRequest) =>
      unwrap(api.POST("/ref/paroki", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateParoki() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateParokiRequest }) =>
      unwrap(
        api.PATCH("/ref/paroki/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateCurrentParokiProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateParokiProfileRequest) =>
      unwrap(api.PATCH("/ref/paroki/current", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteParoki() {
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
        REF_ERRORS,
      ),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: refKeys.all }),
        qc.invalidateQueries({ queryKey: ["user", "me"] }),
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
        REF_ERRORS,
      ),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: refKeys.all }),
        qc.invalidateQueries({ queryKey: ["user", "me"] }),
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
        REF_ERRORS,
      ),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: refKeys.all }),
        qc.invalidateQueries({ queryKey: ["user", "me"] }),
      ]),
  });
}

export function useCreateWilayah() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateWilayahRequest) =>
      unwrap(api.POST("/ref/wilayah", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateWilayah() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateWilayahRequest }) =>
      unwrap(
        api.PATCH("/ref/wilayah/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteWilayah() {
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

export function useCreateLingkungan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateLingkunganRequest) =>
      unwrap(api.POST("/ref/lingkungan", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateLingkungan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateLingkunganRequest }) =>
      unwrap(
        api.PATCH("/ref/lingkungan/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteLingkungan() {
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

export function useCreateKevikepan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateKevikepanRequest) =>
      unwrap(api.POST("/ref/kevikepan", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateKevikepan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateKevikepanRequest }) =>
      unwrap(
        api.PATCH("/ref/kevikepan/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteKevikepan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/ref/kevikepan/{id}", { params: { path: { id } } }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useCreateKota() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateKotaRequest) =>
      unwrap(api.POST("/ref/kota", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateKota() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateKotaRequest }) =>
      unwrap(
        api.PATCH("/ref/kota/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteKota() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/ref/kota/{id}", { params: { path: { id } } }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useCreateKecamatan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateKecamatanRequest) =>
      unwrap(api.POST("/ref/kecamatan", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateKecamatan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateKecamatanRequest }) =>
      unwrap(
        api.PATCH("/ref/kecamatan/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteKecamatan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/ref/kecamatan/{id}", { params: { path: { id } } }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useCreateKelurahan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateKelurahanRequest) =>
      unwrap(api.POST("/ref/kelurahan", { body }), REF_ERRORS),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useUpdateKelurahan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateKelurahanRequest }) =>
      unwrap(
        api.PATCH("/ref/kelurahan/{id}", { params: { path: { id } }, body }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}

export function useDeleteKelurahan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/ref/kelurahan/{id}", { params: { path: { id } } }),
        REF_ERRORS,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: refKeys.all }),
  });
}
