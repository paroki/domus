import { Church, LogOut, Search, Shield, User } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useSearchParams } from "react-router";

import { useLogout } from "~/features/user/api/actions";
import { useMe } from "~/features/user/api/queries";
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
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const { data: me } = useMe();
  const logout = useLogout();
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function onPointerDown(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  async function handleLogout() {
    setOpen(false);
    await logout.mutateAsync();
    navigate("/login", { replace: true });
  }

  const initials = getInitials(me?.name);

  return (
    <header className="flex min-h-15 flex-wrap items-center gap-x-4 gap-y-2 border-b border-line bg-surface px-4 py-2 sm:px-6">
      <Link
        to="/"
        aria-label={t("header.homeLabel")}
        className="flex items-center gap-2.5 text-lg font-bold text-fg no-underline"
      >
        <span className="flex size-8 items-center justify-center rounded-lg bg-primary text-on-primary">
          <Church aria-hidden className="size-5" />
        </span>
        Domus
      </Link>

      {/* Plain GET form: submits to the launcher's ?q= filter, no client state. */}
      <form
        action="/"
        role="search"
        className="order-last flex basis-full justify-center sm:order-none sm:flex-1 sm:basis-0"
      >
        <div className="relative w-full max-w-md">
          <Search
            aria-hidden
            className="pointer-events-none absolute top-1/2 left-4 size-4 -translate-y-1/2 text-muted"
          />
          <input
            type="search"
            name="q"
            defaultValue={params.get("q") ?? ""}
            aria-label={t("header.searchLabel")}
            placeholder={t("header.searchPlaceholder")}
            className="h-10 w-full rounded-full border border-line bg-canvas pr-4 pl-10 text-sm font-medium text-fg placeholder:text-muted"
          />
        </div>
      </form>

      <div className="ml-auto flex items-center gap-1 sm:ml-0">
        <ThemeToggle />
        <div ref={menuRef} className="relative ml-1">
          <button
            type="button"
            aria-label={t("header.account")}
            aria-expanded={open}
            aria-haspopup="menu"
            onClick={() => setOpen((v) => !v)}
            className="flex size-10 items-center justify-center overflow-hidden rounded-full bg-primary text-sm font-bold text-on-primary transition-opacity hover:opacity-90"
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
                className="mt-1 flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm font-medium text-fg no-underline hover:bg-fg/8"
              >
                <Shield aria-hidden className="size-4 text-muted" />
                {t("header.profileSecurity")}
              </Link>
              <button
                type="button"
                role="menuitem"
                disabled={logout.isPending}
                onClick={handleLogout}
                className="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-left text-sm font-medium text-danger hover:bg-danger/10 disabled:opacity-50"
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
