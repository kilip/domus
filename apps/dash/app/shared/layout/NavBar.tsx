import { AppstoreOutlined } from "@ant-design/icons";
import { Button, Layout, Menu, type MenuProps } from "antd";
import type { TFunction } from "i18next";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router";
import { FinanceUnitSelector } from "~/features/Finance/FinanceUnitSelector";
import { modulePath } from "~/shared/modules/registry";
import type { ModuleMenuItem } from "~/shared/modules/types";
import { useActiveModule } from "~/shared/modules/useActiveModule";
import { LanguageSwitcher } from "./LanguageSwitcher";
import { UserMenu } from "./UserMenu";

type AntMenuItems = NonNullable<MenuProps["items"]>;

function toAntMenuItems(
  moduleId: string,
  items: ModuleMenuItem[],
  t: TFunction,
): AntMenuItems {
  return items.map((item) =>
    item.children
      ? {
          key: `group:${item.label}`,
          label: t(item.label),
          children: toAntMenuItems(moduleId, item.children, t),
        }
      : {
          key: item.slug ?? item.label,
          label: (
            <Link to={modulePath(moduleId, item.slug)}>{t(item.label)}</Link>
          ),
        },
  );
}

/** Navbar atas: launcher, nama modul aktif, menu modul, dan menu akun. */
export function NavBar() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { activeModule, activeItem } = useActiveModule();

  return (
    <Layout.Header className="sticky top-0 z-20 flex items-center gap-2">
      <Button
        type="text"
        icon={<AppstoreOutlined />}
        aria-label={t("common.allModules")}
        onClick={() => navigate("/")}
      />
      <Link
        to={activeModule ? modulePath(activeModule.id) : "/"}
        className="shrink-0 px-1"
        style={{ fontWeight: 500, color: "var(--domus-text)" }}
      >
        {activeModule ? t(activeModule.name) : t("common.appName")}
      </Link>
      {activeModule ? (
        <Menu
          mode="horizontal"
          className="min-w-0 flex-1"
          style={{ borderBottom: "none" }}
          selectedKeys={activeItem ? [activeItem.slug] : []}
          items={toAntMenuItems(activeModule.id, activeModule.menu, t)}
        />
      ) : (
        <div className="flex-1" />
      )}
      {activeModule?.id === "fin" && <FinanceUnitSelector />}
      <LanguageSwitcher />
      <UserMenu />
    </Layout.Header>
  );
}
