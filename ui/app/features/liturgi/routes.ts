import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("liturgi", "./features/liturgi/pages/LiturgiSchedulesPage.tsx"),
  route(
    "liturgi/new",
    "./features/liturgi/pages/LiturgiScheduleCreatePage.tsx",
    { id: "liturgi-new" },
  ),
  route(
    "liturgi/ministries",
    "./features/liturgi/pages/LiturgiMinistriesPage.tsx",
    { id: "liturgi-ministries" },
  ),
] satisfies RouteConfig;
