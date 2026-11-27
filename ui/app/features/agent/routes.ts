import { type RouteConfig, route } from "@react-router/dev/routes";

export default [
  route("agent", "./features/agent/pages/AgentPage.tsx"),
  route("agent/memories", "./features/agent/pages/AgentMemoriesPage.tsx"),
  route(
    "agent/memories/new",
    "./features/agent/pages/AgentMemoryFormPage.tsx",
    { id: "agent-memory-new" },
  ),
  route(
    "agent/memories/:memoryId/edit",
    "./features/agent/pages/AgentMemoryFormPage.tsx",
    { id: "agent-memory-edit" },
  ),
  route(
    "agent/sessions/:sessionId",
    "./features/agent/pages/AgentSessionPage.tsx",
  ),
] satisfies RouteConfig;
