import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap } from "~/shared/api/client";
import { eventKeys } from "./keys";
import type {
  CreateEventRequest,
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
      unwrap(api.DELETE("/events/{id}", { params: { path: { id: eventId } } })),
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
      unwrap(
        api.DELETE("/events/{id}/rsvp", { params: { path: { id: eventId } } }),
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: eventKeys.all }),
  });
}
