import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap, unwrapVoid } from "~/shared/api/client";
import { umatKeys } from "./keys";
import type { CreateKeluargaRequest, UpdateKeluargaRequest } from "./types";

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
