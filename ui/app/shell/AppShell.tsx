import { Outlet, useLocation } from "react-router";

import { Header } from "./Header";

export default function AppShell() {
  const { pathname } = useLocation();
  const isAgentWorkspace =
    pathname === "/agent" || pathname.startsWith("/agent/");

  return (
    <div
      className={
        isAgentWorkspace
          ? "flex h-dvh flex-col overflow-hidden pb-[env(safe-area-inset-bottom)]"
          : "flex min-h-screen flex-col pb-[env(safe-area-inset-bottom)]"
      }
    >
      <Header />
      <main className="flex min-h-0 flex-1 flex-col">
        <Outlet />
      </main>
    </div>
  );
}
