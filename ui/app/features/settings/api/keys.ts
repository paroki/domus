export const settingsKeys = {
  all: ["settings"] as const,
  list: () => [...settingsKeys.all, "list"] as const,
  providers: () => [...settingsKeys.all, "providers"] as const,
};
