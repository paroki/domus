import { queryOptions, useQuery } from "@tanstack/react-query";
import { api, unwrap } from "~/shared/api/client";
import { SETTING_ERRORS } from "./errors";
import { settingsKeys } from "./keys";

export function settingsQueryOptions() {
  return queryOptions({
    queryKey: settingsKeys.list(),
    queryFn: () => unwrap(api.GET("/settings"), SETTING_ERRORS),
  });
}

export function useSettings(enabled = true) {
  return useQuery({
    ...settingsQueryOptions(),
    enabled,
  });
}

export function agentProvidersQueryOptions() {
  return queryOptions({
    queryKey: settingsKeys.providers(),
    queryFn: () => unwrap(api.GET("/agent/providers"), SETTING_ERRORS),
  });
}

export function useAgentProviders(enabled = true) {
  return useQuery({
    ...agentProvidersQueryOptions(),
    enabled,
  });
}
