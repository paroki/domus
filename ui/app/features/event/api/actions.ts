import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap, unwrapVoid } from "~/shared/api/client";
import { eventKeys } from "./keys";
import type {
  CreateEventPetugasRequest,
  CreateEventRequest,
  UpdateEventPetugasRequest,
  UpdateEventRequest,
  UpsertEventRSVPRequest,
} from "./types";

export function useCreateEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateEventRequest) =>
      unwrap(api.POST("/events", { body })),
    onSuccess: () => qc.invalidateQueries({ queryKey: eventKeys.all }),
  });
}

export function useUpdateEvent(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateEventRequest) =>
      unwrap(
        api.PATCH("/events/{id}", { params: { path: { id: eventId } }, body }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: eventKeys.all }),
  });
}

export function useDeleteEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (eventId: string) =>
      unwrapVoid(
        api.DELETE("/events/{id}", { params: { path: { id: eventId } } }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: eventKeys.all }),
  });
}

export function useUpsertEventRSVP() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      eventId,
      body,
    }: {
      eventId: string;
      body: UpsertEventRSVPRequest;
    }) =>
      unwrap(
        api.POST("/events/{id}/rsvp", {
          params: { path: { id: eventId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: eventKeys.all }),
  });
}

export function useDeleteEventRSVP(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      unwrapVoid(
        api.DELETE("/events/{id}/rsvp", { params: { path: { id: eventId } } }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: eventKeys.all }),
  });
}

export function useCreateEventPetugas(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateEventPetugasRequest) =>
      unwrap(
        api.POST("/events/{id}/petugas", {
          params: { path: { id: eventId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: eventKeys.all }),
  });
}

export function useUpdateEventPetugas(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      petugasId,
      body,
    }: {
      petugasId: string;
      body: UpdateEventPetugasRequest;
    }) =>
      unwrap(
        api.PATCH("/events/{id}/petugas/{petugasId}", {
          params: { path: { id: eventId, petugasId } },
          body,
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: eventKeys.all }),
  });
}

export function useDeleteEventPetugas(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (petugasId: string) =>
      unwrapVoid(
        api.DELETE("/events/{id}/petugas/{petugasId}", {
          params: { path: { id: eventId, petugasId } },
        }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: eventKeys.all }),
  });
}
