import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("ref", "./features/ref/pages/SettingsPage.tsx"),
  route("ref/profile", "./features/ref/pages/ParokiProfilePage.tsx"),
  route("ref/paroki", "./features/ref/pages/ParokiPage.tsx"),
  route("ref/wilayah", "./features/ref/pages/WilayahPage.tsx"),
  route("ref/lookups", "./features/ref/pages/LookupsPage.tsx"),
  route("ref/periods", "./features/ref/pages/PeriodsPage.tsx"),
] satisfies RouteConfig;
