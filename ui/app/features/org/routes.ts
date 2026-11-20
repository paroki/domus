import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("org", "./features/org/pages/OrgsPage.tsx"),
  route("org/:orgId", "./features/org/pages/OrgDetailPage.tsx"),
] satisfies RouteConfig;
