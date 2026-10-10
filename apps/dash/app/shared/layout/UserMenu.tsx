import {
  LoadingOutlined,
  LoginOutlined,
  LogoutOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { getSession, signOut } from "@domus/better-auth/client";
import { Avatar, Button, Dropdown, Skeleton, Tag } from "antd";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useLocation, useNavigate } from "react-router";

interface SessionUser {
  name: string;
  email: string;
  image?: string | null;
  role?: string | null;
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

/** Inisial dari nama (maks. 2 huruf); jatuh ke email kalau nama kosong. */
function getInitials(user: SessionUser): string {
  const source = user.name.trim() || user.email;
  const parts = source.split(/\s+/).filter(Boolean);
  const letters =
    parts.length > 1
      ? `${parts[0]?.[0] ?? ""}${parts[parts.length - 1]?.[0] ?? ""}`
      : source.slice(0, 2);
  return letters.toUpperCase();
}

function UserAvatar({ user, size }: { user: SessionUser; size: number }) {
  const [broken, setBroken] = useState(false);

  // Reset kalau URL foto berganti (mis. setelah login ulang).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset hanya saat URL berubah
  useEffect(() => setBroken(false), [user.image]);

  const photo =
    user.image && !broken ? (
      // Foto Google sering ditolak (403/429) kalau browser mengirim referrer.
      <img
        src={user.image}
        alt={user.name}
        referrerPolicy="no-referrer"
        onError={() => setBroken(true)}
      />
    ) : undefined;

  return (
    <Avatar
      size={size}
      src={photo}
      style={{
        background: "var(--domus-tint)",
        color: "var(--domus-tint-text)",
        fontWeight: 500,
        flexShrink: 0,
      }}
    >
      {getInitials(user) || <UserOutlined />}
    </Avatar>
  );
}

/** Menu akun di navbar: avatar, info akun, dan keluar. */
export function UserMenu() {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const user = useSessionUser();
  const [open, setOpen] = useState(false);
  const [signingOut, setSigningOut] = useState(false);

  if (user === undefined) {
    return (
      <Skeleton.Avatar
        active
        size={26}
        aria-label={t("common.loadingAccount")}
      />
    );
  }

  if (user === null) {
    const target = `${location.pathname}${location.search}`;
    const to =
      target === "/"
        ? "/login"
        : `/login?redirect=${encodeURIComponent(target)}`;
    return (
      <Button type="text" icon={<LoginOutlined />} onClick={() => navigate(to)}>
        {t("common.login")}
      </Button>
    );
  }

  async function handleSignOut() {
    if (signingOut) return;
    setSigningOut(true);
    try {
      await signOut();
    } finally {
      setSigningOut(false);
      setOpen(false);
      navigate("/login");
    }
  }

  const roleLabel = user.role
    ? i18n.exists(`roles.${user.role}`)
      ? t(`roles.${user.role}` as "roles.admin")
      : user.role
    : null;

  return (
    <Dropdown
      trigger={["click"]}
      placement="bottomRight"
      open={open}
      // Biarkan menu terbuka selama proses keluar supaya status loading terlihat.
      onOpenChange={(next) => {
        if (!signingOut) setOpen(next);
      }}
      menu={{
        style: { minWidth: 240 },
        items: [
          {
            key: "who",
            disabled: true,
            style: { cursor: "default", opacity: 1 },
            label: (
              <span className="flex items-center gap-3 py-1">
                <UserAvatar user={user} size={36} />
                <span className="flex min-w-0 flex-col leading-tight">
                  <strong
                    className="truncate"
                    style={{ color: "var(--domus-text)" }}
                  >
                    {user.name || user.email}
                  </strong>
                  <span
                    className="truncate"
                    style={{
                      fontSize: 13,
                      color: "var(--domus-text-secondary)",
                    }}
                  >
                    {user.email}
                  </span>
                  {roleLabel && (
                    <Tag
                      className="mt-1 w-fit"
                      variant="filled"
                      style={{
                        margin: "4px 0 0",
                        background: "var(--domus-tint)",
                        color: "var(--domus-tint-text)",
                      }}
                    >
                      {roleLabel}
                    </Tag>
                  )}
                </span>
              </span>
            ),
          },
          { type: "divider" },
          {
            key: "logout",
            danger: true,
            disabled: signingOut,
            icon: signingOut ? <LoadingOutlined /> : <LogoutOutlined />,
            label: signingOut ? t("common.loggingOut") : t("common.logout"),
          },
        ],
        onClick: ({ key }) => {
          if (key === "logout") void handleSignOut();
        },
      }}
    >
      <Button
        type="text"
        aria-label={t("userMenu.accountMenu", {
          name: user.name || user.email,
        })}
        aria-haspopup="menu"
        aria-expanded={open}
        className="!px-1"
      >
        <UserAvatar user={user} size={26} />
      </Button>
    </Dropdown>
  );
}
