import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("me", "./features/user/pages/MePage.tsx"),
  route("me/verify", "./features/user/pages/VerifyPage.tsx"),
  route("users", "./features/user/pages/UserPage.tsx"),
  route("users/:id", "./features/user/pages/UserDetailPage.tsx"),
] satisfies RouteConfig;
