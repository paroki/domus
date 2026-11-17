import { Church, Search, User } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link, useSearchParams } from "react-router";

import { ThemeToggle } from "./ThemeToggle";

export function Header() {
  const { t } = useTranslation();
  const [params] = useSearchParams();

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
      <search className="order-last flex basis-full justify-center sm:order-none sm:flex-1 sm:basis-0">
        <form action="/" className="flex w-full justify-center">
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
      </search>

      <div className="ml-auto flex items-center gap-1 sm:ml-0">
        <ThemeToggle />
        <span
          role="img"
          aria-label={t("header.account")}
          className="ml-1 flex size-10 items-center justify-center rounded-full bg-primary text-on-primary"
        >
          <User aria-hidden className="size-5" />
        </span>
      </div>
    </header>
  );
}
