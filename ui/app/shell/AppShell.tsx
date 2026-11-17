import { Outlet } from "react-router";

import { Header } from "./Header";

export default function AppShell() {
  return (
    <div className="flex min-h-screen flex-col">
      <Header />
      <main className="flex flex-1 flex-col">
        <Outlet />
      </main>
    </div>
  );
}
