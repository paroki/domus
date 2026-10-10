import {
  AppstoreOutlined,
  LogoutOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { getSession, signOut } from "@domus/better-auth/client";
import { Avatar, Button, Dropdown, Layout, Menu, type MenuProps } from "antd";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router";
import { modulePath } from "~/shared/modules/registry";
import type { ModuleMenuItem } from "~/shared/modules/types";
import { useActiveModule } from "~/shared/modules/useActiveModule";

type AntMenuItems = NonNullable<MenuProps["items"]>;

function toAntMenuItems(
  moduleId: string,
  items: ModuleMenuItem[],
): AntMenuItems {
  return items.map((item) =>
    item.children
      ? {
          key: `group:${item.label}`,
          label: item.label,
          children: toAntMenuItems(moduleId, item.children),
        }
      : {
          key: item.slug ?? item.label,
          label: <Link to={modulePath(moduleId, item.slug)}>{item.label}</Link>,
        },
  );
}

interface SessionUser {
  name: string;
  email: string;
}

/** `undefined` = masih memuat, `null` = belum login. */
function useSessionUser(): SessionUser | null | undefined {
  const [user, setUser] = useState<SessionUser | null | undefined>(undefined);

  useEffect(() => {
    let active = true;
    getSession()
      .then(({ data }) => {
        if (active) setUser(data?.user ?? null);
      })
      .catch(() => {
        if (active) setUser(null);
      });
    return () => {
      active = false;
    };
  }, []);

  return user;
}

function UserMenu() {
  const navigate = useNavigate();
  const user = useSessionUser();

  if (user === undefined) return null;

  if (user === null) {
    return (
      <Button type="text" onClick={() => navigate("/login")}>
        Masuk
      </Button>
    );
  }

  async function handleSignOut() {
    try {
      await signOut();
    } finally {
      navigate("/login");
    }
  }

  return (
    <Dropdown
      trigger={["click"]}
      placement="bottomRight"
      menu={{
        items: [
          {
            key: "who",
            disabled: true,
            label: (
              <span className="flex flex-col leading-tight">
                <strong>{user.name}</strong>
                <span style={{ fontSize: 12 }}>{user.email}</span>
              </span>
            ),
          },
          { type: "divider" },
          {
            key: "logout",
            danger: true,
            icon: <LogoutOutlined />,
            label: "Keluar",
          },
        ],
        onClick: ({ key }) => {
          if (key === "logout") void handleSignOut();
        },
      }}
    >
      <Button type="text" aria-label="Menu akun" className="!px-1">
        <Avatar
          size={26}
          icon={<UserOutlined />}
          style={{
            background: "var(--domus-tint)",
            color: "var(--domus-tint-text)",
          }}
        />
      </Button>
    </Dropdown>
  );
}

/** Navbar atas: launcher, nama modul aktif, menu modul, dan menu akun. */
export function NavBar() {
  const navigate = useNavigate();
  const { activeModule, activeItem } = useActiveModule();

  return (
    <Layout.Header className="sticky top-0 z-20 flex items-center gap-2">
      <Button
        type="text"
        icon={<AppstoreOutlined />}
        aria-label="Semua modul"
        onClick={() => navigate("/")}
      />
      <Link
        to={activeModule ? modulePath(activeModule.id) : "/"}
        className="shrink-0 px-1"
        style={{ fontWeight: 500, color: "var(--domus-text)" }}
      >
        {activeModule ? activeModule.name : "Domus"}
      </Link>
      {activeModule ? (
        <Menu
          mode="horizontal"
          className="min-w-0 flex-1"
          style={{ borderBottom: "none" }}
          selectedKeys={activeItem ? [activeItem.slug] : []}
          items={toAntMenuItems(activeModule.id, activeModule.menu)}
        />
      ) : (
        <div className="flex-1" />
      )}
      <UserMenu />
    </Layout.Header>
  );
}
