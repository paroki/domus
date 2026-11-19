import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("users", "./features/user/pages/UserPage.tsx"),
] satisfies RouteConfig;
