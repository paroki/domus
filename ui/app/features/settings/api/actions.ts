import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap } from "~/shared/api/client";
import { SETTING_ERRORS } from "./errors";
import { settingsKeys } from "./keys";
import type { UpdateSettingsRequest } from "./types";

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

export function useSyncProviderModels() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (provider: string) =>
      unwrap(
        api.POST("/agent/providers/{provider}/sync-models", {
          params: { path: { provider } },
        }),
        SETTING_ERRORS,
      ),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: settingsKeys.all });
    },
  });
}
