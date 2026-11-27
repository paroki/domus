import { ChevronDown, Church, LogOut, Shield, User } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router";

import {
  useLogout,
  useMe,
  useSessionParoki,
  useSwitchParoki,
} from "~/shared/api/session";
import { ThemeToggle } from "./ThemeToggle";

function getInitials(name?: string): string {
  if (!name) return "";
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "";
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return `${parts[0][0]}${parts[parts.length - 1][0]}`.toUpperCase();
}

export function Header() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: me } = useMe();
  const logout = useLogout();
  const switchParoki = useSwitchParoki();
  const isSuperAdmin = Boolean(me?.roles?.includes("user:super-admin"));
  const parokiQuery = useSessionParoki(
    {
      filter: isSuperAdmin
        ? { active: true }
        : { active: true, memberOnly: true },
    },
    Boolean(me),
  );
  const parokiList =
    parokiQuery.data?.pages.flatMap((page) => page.data ?? []) ?? [];
  const activeParoki = parokiList.find((p) => p.id === me?.activeParokiId);
  const [parokiOpen, setParokiOpen] = useState(false);
  const parokiRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open && !parokiOpen) return;
    function onPointerDown(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
      if (parokiRef.current && !parokiRef.current.contains(e.target as Node)) {
        setParokiOpen(false);
      }
    }
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {
        setOpen(false);
        setParokiOpen(false);
      }
    }
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open, parokiOpen]);

  async function handleLogout() {
    setOpen(false);
    await logout.mutateAsync();
    navigate("/login", { replace: true });
  }

  const initials = getInitials(me?.name);

  return (
    <header className="flex min-h-15 flex-wrap items-center gap-x-4 gap-y-2 border-b border-line bg-surface px-4 py-2 pt-[env(safe-area-inset-top)] sm:px-6">
      <Link
        to="/"
        aria-label={t("header.homeLabel")}
        className="flex items-center gap-2.5 text-lg font-bold text-fg no-underline"
      >
        <img src="/logo.svg" alt="" aria-hidden className="size-8 rounded-lg" />
        Domus
      </Link>

      <div className="ml-auto flex items-center gap-1">
        {me && (
          <div ref={parokiRef} className="relative mr-1">
            <button
              type="button"
              aria-label={t("header.parishSwitcher")}
              aria-expanded={parokiOpen}
              aria-haspopup="menu"
              onClick={() => setParokiOpen((v) => !v)}
              className="inline-flex min-h-11 max-w-44 items-center gap-1.5 rounded-full border border-line bg-canvas px-3 text-xs font-semibold text-fg hover:bg-fg/8"
            >
              <Church aria-hidden className="size-3.5 shrink-0 text-muted" />
              <span className="truncate">
                {activeParoki?.nama ?? t("header.selectParish")}
              </span>
              <ChevronDown
                aria-hidden
                className="size-3.5 shrink-0 text-muted"
              />
            </button>
            {parokiOpen && (
              <div
                role="menu"
                aria-label={t("header.parishSwitcher")}
                className="absolute right-0 z-50 mt-2 max-h-64 w-60 overflow-y-auto rounded-xl border border-line bg-surface p-1.5 shadow-lg"
              >
                {parokiList.length === 0 ? (
                  <p className="px-3 py-2 text-xs text-muted">
                    {t("ref.emptyList")}
                  </p>
                ) : (
                  parokiList.map((p) => (
                    <button
                      key={p.id}
                      type="button"
                      role="menuitem"
                      disabled={switchParoki.isPending}
                      onClick={async () => {
                        if (!p.id) return;
                        setParokiOpen(false);
                        await switchParoki.mutateAsync({
                          activeParokiId: p.id,
                        });
                      }}
                      className={`flex min-h-11 w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm hover:bg-fg/8 ${
                        p.id === me.activeParokiId
                          ? "font-semibold text-primary"
                          : "text-fg"
                      }`}
                    >
                      <span className="truncate">{p.nama}</span>
                      <span className="ml-2 text-xs text-muted">{p.kode}</span>
                    </button>
                  ))
                )}
              </div>
            )}
          </div>
        )}
        <ThemeToggle />
        <div ref={menuRef} className="relative ml-1">
          <button
            type="button"
            aria-label={t("header.account")}
            aria-expanded={open}
            aria-haspopup="menu"
            onClick={() => setOpen((v) => !v)}
            className="flex size-11 items-center justify-center overflow-hidden rounded-full bg-primary text-sm font-bold text-on-primary transition-opacity hover:opacity-90"
          >
            {me?.avatarUrl ? (
              <img
                src={me.avatarUrl}
                alt={me.name}
                className="size-full object-cover"
              />
            ) : initials ? (
              <span>{initials}</span>
            ) : (
              <User aria-hidden className="size-5" />
            )}
          </button>

          {open && (
            <div
              role="menu"
              aria-label={t("header.account")}
              className="absolute right-0 z-50 mt-2 w-60 rounded-xl border border-line bg-surface p-1.5 shadow-lg"
            >
              {me && (
                <div className="border-b border-line px-3 py-2">
                  <p className="truncate text-sm font-semibold text-fg">
                    {me.name}
                  </p>
                  <p className="truncate text-xs text-muted">{me.email}</p>
                </div>
              )}
              <Link
                to="/me"
                role="menuitem"
                onClick={() => setOpen(false)}
                className="mt-1 flex min-h-11 items-center gap-2.5 rounded-lg px-3 py-2 text-sm font-medium text-fg no-underline hover:bg-fg/8"
              >
                <Shield aria-hidden className="size-4 text-muted" />
                {t("header.profileSecurity")}
              </Link>
              <button
                type="button"
                role="menuitem"
                disabled={logout.isPending}
                onClick={handleLogout}
                className="flex min-h-11 w-full items-center gap-2.5 rounded-lg px-3 py-2 text-left text-sm font-medium text-danger hover:bg-danger/10 disabled:opacity-50"
              >
                <LogOut aria-hidden className="size-4" />
                {t("header.logout")}
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
