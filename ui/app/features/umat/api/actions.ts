import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap, unwrapVoid } from "~/shared/api/client";
import { umatKeys } from "./keys";
import type {
  CreateKeluargaRequest,
  CreateUmatReferensiRequest,
  CreateUmatRequest,
  UpdateKeluargaRequest,
  UpdateUmatReferensiRequest,
  UpdateUmatRequest,
  UpdateUmatSakramenRequest,
  UpsertUmatSakramenRequest,
} from "./types";

export function useCreateKeluarga(parokiId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateKeluargaRequest) =>
      unwrap(api.POST("/umat/keluarga", { body })),
    onSuccess: (created) => {
      qc.invalidateQueries({ queryKey: umatKeys.keluargaLists() });
      if (created?.id) {
        qc.setQueryData(umatKeys.keluargaDetail(parokiId, created.id), created);
      }
    },
  });
}

export function useUpdateKeluarga(parokiId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: { id: string; body: UpdateKeluargaRequest }) =>
      unwrap(
        api.PATCH("/umat/keluarga/{id}", {
          params: { path: { id: vars.id } },
          body: vars.body,
        }),
      ),
    onSuccess: (updated, vars) => {
      qc.invalidateQueries({ queryKey: umatKeys.keluargaLists() });
      qc.setQueryData(umatKeys.keluargaDetail(parokiId, vars.id), updated);
    },
  });
}

export function useDeleteKeluarga(parokiId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/umat/keluarga/{id}", {
          params: { path: { id } },
        }),
      ),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: umatKeys.keluargaLists() });
      qc.removeQueries({ queryKey: umatKeys.keluargaDetail(parokiId, id) });
    },
  });
}

export function useCreateUmat(parokiId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateUmatRequest) =>
      unwrap(api.POST("/umat", { body })),
    onSuccess: (created) => {
      qc.invalidateQueries({ queryKey: umatKeys.umatLists() });
      qc.invalidateQueries({ queryKey: umatKeys.keluargaDetails() });
      if (created?.id) {
        qc.setQueryData(umatKeys.umatDetail(parokiId, created.id), created);
      }
    },
  });
}

export function useUpdateUmat(parokiId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: { id: string; body: UpdateUmatRequest }) =>
      unwrap(
        api.PATCH("/umat/{id}", {
          params: { path: { id: vars.id } },
          body: vars.body,
        }),
      ),
    onSuccess: (updated, vars) => {
      qc.invalidateQueries({ queryKey: umatKeys.umatLists() });
      qc.invalidateQueries({ queryKey: umatKeys.keluargaDetails() });
      qc.setQueryData(umatKeys.umatDetail(parokiId, vars.id), updated);
    },
  });
}

export function useDeleteUmat(parokiId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/umat/{id}", {
          params: { path: { id } },
        }),
      ),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: umatKeys.umatLists() });
      qc.invalidateQueries({ queryKey: umatKeys.keluargaDetails() });
      qc.removeQueries({ queryKey: umatKeys.umatDetail(parokiId, id) });
    },
  });
}

export function useUpsertSakramen(parokiId: string, umatId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpsertUmatSakramenRequest) =>
      unwrap(
        api.POST("/umat/{id}/sakramen", {
          params: { path: { id: umatId } },
          body,
        }),
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: umatKeys.umatDetail(parokiId, umatId) });
    },
  });
}

export function useUpdateSakramen(parokiId: string, umatId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: {
      sakramenId: string;
      body: UpdateUmatSakramenRequest;
    }) =>
      unwrap(
        api.PATCH("/umat/{id}/sakramen/{sakramenId}", {
          params: { path: { id: umatId, sakramenId: vars.sakramenId } },
          body: vars.body,
        }),
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: umatKeys.umatDetail(parokiId, umatId) });
    },
  });
}

export function useDeleteSakramen(parokiId: string, umatId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (sakramenId: string) =>
      unwrapVoid(
        api.DELETE("/umat/{id}/sakramen/{sakramenId}", {
          params: { path: { id: umatId, sakramenId } },
        }),
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: umatKeys.umatDetail(parokiId, umatId) });
    },
  });
}

export function useCreateUmatReferensi() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateUmatReferensiRequest) =>
      unwrap(api.POST("/umat/referensi", { body })),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: umatKeys.referensiLists() });
      qc.invalidateQueries({ queryKey: umatKeys.refs() });
    },
  });
}

export function useUpdateUmatReferensi(parokiId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: { id: string; body: UpdateUmatReferensiRequest }) =>
      unwrap(
        api.PATCH("/umat/referensi/{id}", {
          params: { path: { id: vars.id } },
          body: vars.body,
        }),
      ),
    onSuccess: (updated, vars) => {
      qc.invalidateQueries({ queryKey: umatKeys.referensiLists() });
      qc.invalidateQueries({ queryKey: umatKeys.refs() });
      qc.setQueryData(umatKeys.referensiDetail(parokiId, vars.id), updated);
    },
  });
}

export function useDeleteUmatReferensi(parokiId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/umat/referensi/{id}", {
          params: { path: { id } },
        }),
      ),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: umatKeys.referensiLists() });
      qc.invalidateQueries({ queryKey: umatKeys.refs() });
      qc.removeQueries({ queryKey: umatKeys.referensiDetail(parokiId, id) });
    },
  });
}
