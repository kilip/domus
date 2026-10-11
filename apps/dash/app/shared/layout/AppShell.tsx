import { Layout } from "antd";
import { Outlet } from "react-router";
import { FinanceProvider } from "~/features/Finance/FinanceContext";
import { NavBar } from "./NavBar";

/** Shell aplikasi: navbar di atas, konten (halaman modul) di bawahnya. */
export default function AppShell() {
  return (
    <FinanceProvider>
      <Layout className="min-h-screen">
        <NavBar />
        <Layout.Content className="flex flex-col">
          <Outlet />
        </Layout.Content>
      </Layout>
    </FinanceProvider>
  );
}
