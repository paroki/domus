import { Outlet, useLocation, useSearchParams } from "react-router";

import { Header } from "./Header";

export default function AppShell() {
  const { pathname } = useLocation();
  const [searchParams] = useSearchParams();
  const isGabrielHome = pathname === "/" && searchParams.get("tab") !== "menu";

  return (
    <div
      className={
        isGabrielHome
          ? "flex h-dvh flex-col overflow-hidden"
          : "flex min-h-screen flex-col"
      }
    >
      <Header />
      <main className="flex min-h-0 flex-1 flex-col">
        <Outlet />
      </main>
    </div>
  );
}
