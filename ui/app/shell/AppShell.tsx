import { Outlet, useLocation } from "react-router";
import { AgentBottomDock } from "~/features/agent/components/AgentBottomDock";
import { Header } from "./Header";

export default function AppShell() {
  const { pathname } = useLocation();
  const isAgentRoute = pathname === "/agent" || pathname.startsWith("/agent/");
  const isAgentWorkspace =
    pathname === "/agent" || pathname.startsWith("/agent/sessions/");

  return (
    <div
      className={
        isAgentWorkspace
          ? "flex h-dvh flex-col overflow-hidden pb-[env(safe-area-inset-bottom)]"
          : isAgentRoute
            ? "flex min-h-screen flex-col pb-[env(safe-area-inset-bottom)]"
            : "flex min-h-screen flex-col pb-[calc(5.5rem+env(safe-area-inset-bottom))]"
      }
    >
      <Header />
      <main className="flex min-h-0 flex-1 flex-col">
        <Outlet />
      </main>
      {!isAgentRoute ? <AgentBottomDock /> : null}
    </div>
  );
}
