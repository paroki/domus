import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("ref", "./features/ref/pages/SettingsPage.tsx"),
  route("ref/profile", "./features/ref/pages/ParishProfilePage.tsx"),
  route("ref/parishes", "./features/ref/pages/ParishesPage.tsx"),
  route("ref/territory", "./features/ref/pages/TerritoryPage.tsx"),
  route("ref/lookups", "./features/ref/pages/LookupsPage.tsx"),
  route("ref/periods", "./features/ref/pages/PeriodsPage.tsx"),
] satisfies RouteConfig;
