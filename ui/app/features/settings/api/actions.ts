import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap, unwrapVoid } from "~/shared/api/client";
import { SETTING_ERRORS } from "./errors";
import { settingsKeys } from "./keys";
import type {
  CreateAgentProviderRequest,
  UpdateAgentProviderRequest,
  UpdateSettingsRequest,
} from "./types";

export function useUpdateSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateSettingsRequest) =>
      unwrap(api.PATCH("/settings", { body }), SETTING_ERRORS),
    onSuccess: (data) => {
      queryClient.setQueryData(settingsKeys.list(), data);
      void queryClient.invalidateQueries({ queryKey: settingsKeys.all });
    },
  });
}

export function useCreateAgentProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateAgentProviderRequest) =>
      unwrap(api.POST("/agent/providers", { body }), SETTING_ERRORS),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: settingsKeys.providers(),
      });
    },
  });
}

export function useUpdateAgentProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      providerId,
      body,
    }: {
      providerId: string;
      body: UpdateAgentProviderRequest;
    }) =>
      unwrap(
        api.PATCH("/agent/providers/{providerId}", {
          params: { path: { providerId } },
          body,
        }),
        SETTING_ERRORS,
      ),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: settingsKeys.providers(),
      });
    },
  });
}

export function useDeleteAgentProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (providerId: string) =>
      unwrapVoid(
        api.DELETE("/agent/providers/{providerId}", {
          params: { path: { providerId } },
        }),
        SETTING_ERRORS,
      ),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: settingsKeys.providers(),
      });
    },
  });
}

export function useSyncProviderModels() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (providerId: string) =>
      unwrap(
        api.POST("/agent/providers/{providerId}/sync-models", {
          params: { path: { providerId } },
        }),
        SETTING_ERRORS,
      ),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: settingsKeys.providers(),
      });
    },
  });
}
