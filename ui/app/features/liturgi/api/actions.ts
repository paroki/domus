import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap, unwrapVoid } from "~/shared/api/client";
import { liturgiKeys } from "./keys";
import type {
  ApproveLiturgiSubstituteRequest,
  CreateLiturgiLaguRequest,
  CreateLiturgiPetugasRequest,
  CreateLiturgiScheduleRequest,
  RespondLiturgiPetugasRequest,
  UpdateLiturgiLaguRequest,
  UpdateLiturgiMinistryRequest,
  UpdateLiturgiPetugasRequest,
  UpdateLiturgiScheduleRequest,
  UpsertEventRSVPRequest,
} from "./types";

export function useUpdateLiturgiMinistry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      roleSlug,
      body,
    }: {
      roleSlug: string;
      body: UpdateLiturgiMinistryRequest;
    }) =>
      unwrap(
        api.PATCH("/liturgi/ministries/{roleSlug}", {
          params: { path: { roleSlug } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useCreateLiturgiSchedule() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateLiturgiScheduleRequest) =>
      unwrap(api.POST("/liturgi/schedules", { body })),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useUpdateLiturgiSchedule(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateLiturgiScheduleRequest) =>
      unwrap(
        api.PATCH("/liturgi/schedules/{eventId}", {
          params: { path: { eventId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useDeleteLiturgiSchedule() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (eventId: string) =>
      unwrapVoid(
        api.DELETE("/liturgi/schedules/{eventId}", {
          params: { path: { eventId } },
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useCreateLiturgiPetugas(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateLiturgiPetugasRequest) =>
      unwrap(
        api.POST("/liturgi/schedules/{eventId}/petugas", {
          params: { path: { eventId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useUpdateLiturgiPetugas(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      petugasId,
      body,
    }: {
      petugasId: string;
      body: UpdateLiturgiPetugasRequest;
    }) =>
      unwrap(
        api.PATCH("/liturgi/schedules/{eventId}/petugas/{petugasId}", {
          params: { path: { eventId, petugasId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useDeleteLiturgiPetugas(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (petugasId: string) =>
      unwrapVoid(
        api.DELETE("/liturgi/schedules/{eventId}/petugas/{petugasId}", {
          params: { path: { eventId, petugasId } },
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useRespondLiturgiPetugas(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      petugasId,
      body,
    }: {
      petugasId: string;
      body: RespondLiturgiPetugasRequest;
    }) =>
      unwrap(
        api.PATCH("/liturgi/schedules/{eventId}/petugas/{petugasId}/respond", {
          params: { path: { eventId, petugasId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useApproveLiturgiSubstitute(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      petugasId,
      body,
    }: {
      petugasId: string;
      body: ApproveLiturgiSubstituteRequest;
    }) =>
      unwrap(
        api.POST(
          "/liturgi/schedules/{eventId}/petugas/{petugasId}/approve-substitute",
          {
            params: { path: { eventId, petugasId } },
            body,
          },
        ),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useCreateLiturgiLagu(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateLiturgiLaguRequest) =>
      unwrap(
        api.POST("/liturgi/schedules/{eventId}/lagu", {
          params: { path: { eventId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useUpdateLiturgiLagu(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      laguId,
      body,
    }: {
      laguId: string;
      body: UpdateLiturgiLaguRequest;
    }) =>
      unwrap(
        api.PATCH("/liturgi/schedules/{eventId}/lagu/{laguId}", {
          params: { path: { eventId, laguId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useDeleteLiturgiLagu(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (laguId: string) =>
      unwrapVoid(
        api.DELETE("/liturgi/schedules/{eventId}/lagu/{laguId}", {
          params: { path: { eventId, laguId } },
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useUpsertLiturgiRSVP(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpsertEventRSVPRequest) =>
      unwrap(
        api.POST("/events/{id}/rsvp", {
          params: { path: { id: eventId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}

export function useDeleteLiturgiRSVP(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      unwrapVoid(
        api.DELETE("/events/{id}/rsvp", {
          params: { path: { id: eventId } },
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: liturgiKeys.all }),
  });
}
