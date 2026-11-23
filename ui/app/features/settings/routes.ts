import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("settings", "./features/settings/pages/SettingsPage.tsx"),
  route("settings/telegram", "./features/settings/pages/TelegramPage.tsx"),
  route("settings/whatsapp", "./features/settings/pages/WhatsAppPage.tsx"),
  route("settings/ai-provider", "./features/settings/pages/AIProviderPage.tsx"),
] satisfies RouteConfig;
