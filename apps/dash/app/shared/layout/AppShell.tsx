import { Layout } from "antd";
import { Outlet } from "react-router";
import { NavBar } from "./NavBar";

/** Shell aplikasi: navbar di atas, konten (halaman modul) di bawahnya. */
export default function AppShell() {
  return (
    <Layout className="min-h-screen">
      <NavBar />
      <Layout.Content className="flex flex-col">
        <Outlet />
      </Layout.Content>
    </Layout>
  );
}
