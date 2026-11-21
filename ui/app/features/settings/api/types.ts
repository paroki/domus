import type { components } from "~/shared/api/schema";

type Schemas = components["schemas"];

export type SettingResponse = Schemas["setting.SettingResponse"];
export type UpdateSettingsRequest = Schemas["setting.UpdateSettingsRequest"];
