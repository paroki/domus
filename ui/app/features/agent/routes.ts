import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("agent", "./features/agent/pages/AgentPage.tsx"),
  route(
    "agent/sessions/:sessionId",
    "./features/agent/pages/AgentSessionPage.tsx",
  ),
] satisfies RouteConfig;
