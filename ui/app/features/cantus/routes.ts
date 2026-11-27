import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("cantus", "./features/cantus/pages/CantusDirectoryPage.tsx"),
  route("cantus/new", "./features/cantus/pages/CantusPlaceholderPage.tsx", {
    id: "cantus-new",
  }),
  route(
    "cantus/:cantusId",
    "./features/cantus/pages/CantusPlaceholderPage.tsx",
    { id: "cantus-detail" },
  ),
] satisfies RouteConfig;
