import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap, unwrapVoid } from "~/shared/api/client";
import { cantusKeys } from "./keys";
import type { CreateCantusRequest, UpdateCantusRequest } from "./types";

export function useCreateCantus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateCantusRequest) =>
      unwrap(api.POST("/cantus", { body })),
    onSuccess: (created) => {
      qc.invalidateQueries({ queryKey: cantusKeys.lists() });
      if (created?.id) {
        qc.setQueryData(cantusKeys.detail(created.id), created);
      }
    },
  });
}

export function useUpdateCantus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: { id: string; body: UpdateCantusRequest }) =>
      unwrap(
        api.PATCH("/cantus/{id}", {
          params: { path: { id: vars.id } },
          body: vars.body,
        }),
      ),
    onSuccess: (updated, vars) => {
      qc.invalidateQueries({ queryKey: cantusKeys.lists() });
      qc.setQueryData(cantusKeys.detail(vars.id), updated);
    },
  });
}

export function useDeleteCantus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      unwrapVoid(
        api.DELETE("/cantus/{id}", {
          params: { path: { id } },
        }),
      ),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: cantusKeys.lists() });
      qc.removeQueries({ queryKey: cantusKeys.detail(id) });
    },
  });
}
