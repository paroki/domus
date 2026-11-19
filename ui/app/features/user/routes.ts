import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("me", "./features/user/pages/MePage.tsx"),
  route("users", "./features/user/pages/UserPage.tsx"),
] satisfies RouteConfig;
