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
