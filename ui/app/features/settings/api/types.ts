import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type SettingResponse = Schemas["setting.SettingResponse"] & {
  multiline?: boolean;
  overridden?: boolean;
};
export type UpdateSettingsRequest = Schemas["setting.UpdateSettingsRequest"];
export type ProviderModelItem = Schemas["domain.ProviderModelDTO"] & {
  id?: string;
  name?: string;
  capabilities?: string[];
};
export type SyncProviderModelsResponse =
  Schemas["domain.SyncProviderModelsResponse"];
export type AgentProviderResponse = Schemas["domain.AgentProviderResponse"];
export type CreateAgentProviderRequest =
  Schemas["domain.CreateAgentProviderRequest"];
export type UpdateAgentProviderRequest =
  Schemas["domain.UpdateAgentProviderRequest"];
