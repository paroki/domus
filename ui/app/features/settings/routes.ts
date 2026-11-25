import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("settings", "./features/settings/pages/SettingsPage.tsx"),
  route("settings/system", "./features/settings/pages/SystemHubPage.tsx"),
  route(
    "settings/system/telegram",
    "./features/settings/pages/TelegramPage.tsx",
  ),
  route(
    "settings/system/whatsapp",
    "./features/settings/pages/WhatsAppPage.tsx",
  ),
  route(
    "settings/system/ai-provider",
    "./features/settings/pages/AIProviderPage.tsx",
  ),
  route(
    "settings/system/ai-agent",
    "./features/settings/pages/AIAgentPage.tsx",
  ),
] satisfies RouteConfig;
