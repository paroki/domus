import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type SettingResponse = Schemas["setting.SettingResponse"];
export type UpdateSettingsRequest = Schemas["setting.UpdateSettingsRequest"];
export type ProviderModelItem = Schemas["domain.ProviderModelDTO"] & {
  id?: string;
  name?: string;
  capabilities?: string[];
};
export type SyncProviderModelsResponse =
  Schemas["domain.SyncProviderModelsResponse"];
