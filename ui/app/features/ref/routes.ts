import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("settings", "./features/ref/pages/SettingsPage.tsx"),
  route("settings/profile", "./features/ref/pages/ParishProfilePage.tsx"),
  route("settings/parishes", "./features/ref/pages/ParishesPage.tsx"),
] satisfies RouteConfig;
