import type { SettingResponse } from "./api/types";

export const ROLE_SUPER_ADMIN = "user:super-admin";

export function isSuperAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes(ROLE_SUPER_ADMIN));
}

export interface SettingGroup {
  name: string;
  items: SettingResponse[];
  configured: boolean;
}

export function groupSettings(items: SettingResponse[]): SettingGroup[] {
  const groups = new Map<string, SettingResponse[]>();
  for (const item of items) {
    const name = item.group || "Umum";
    const list = groups.get(name);
    if (list) {
      list.push(item);
    } else {
      groups.set(name, [item]);
    }
  }
  return Array.from(groups.entries()).map(([name, groupItems]) => ({
    name,
    items: groupItems,
    configured: groupItems.some((item) => Boolean(item.configured)),
  }));
}

export function buildClearValues(
  items: SettingResponse[],
): Record<string, string> {
  const values: Record<string, string> = {};
  for (const item of items) {
    if (item.key) {
      values[item.key] = "";
    }
  }
  return values;
}
