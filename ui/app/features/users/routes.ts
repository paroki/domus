import { route, type RouteConfig } from "@react-router/dev/routes";

export default [
  route("users", "./features/users/pages/UsersPage.tsx"),
] satisfies RouteConfig;
