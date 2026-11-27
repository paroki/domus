import { ChevronRight, Plus, Search, ShieldAlert, Users } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useSearchParams } from "react-router";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  InductiveEmptyState,
  InductiveFilterBar,
  InductivePageHeader,
} from "~/shared/ui/iui";
import {
  useAuthRoles,
  useDirectoryManagedOrgUsers,
  useDirectoryParokiMembers,
  useMe,
  useUsersSearch,
} from "../api/queries";
import { UserAvatar, useDebouncedValue } from "../components/DirectoryShared";
import {
  canManageUsers,
  getAvailableUserScopes,
  isCantusAdmin,
  isSuperAdmin,
  providerLabel,
  USER_ROUTES,
  type UserDirectoryScope,
} from "../model";
import type { Route } from "./+types/UserPage";

export type DirectoryFeedback = {
  tone: "success" | "error";
  message: string;
};

export type UserDirectoryPreview = {
  id: string;
  name: string;
  email: string;
  phone?: string;
};

type UnifiedUserRow = UserDirectoryPreview & {
  status?: string;
  highlight?: boolean;
  isSuperAdmin?: boolean;
  isCantusAdmin?: boolean;
  parishRole?: string;
  umatAdmin?: boolean;
  liturgiAdmin?: boolean;
  onboarded?: boolean;
  metaLine: string;
};

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.user.name")} · Domus` }];
}

export default function UserPage() {
  const { t } = useTranslation();
  const { data: me } = useMe();
  useAuthRoles();
  const [searchParams, setSearchParams] = useSearchParams();

  const [q, setQ] = useState("");
  const debouncedQ = useDebouncedValue(q.trim(), 250);
  const [statusFilter, setStatusFilter] = useState<
    "all" | "active" | "disabled"
  >("all");
  const [cantusAdminFilter, setCantusAdminFilter] = useState<
    "all" | "cantusAdmin"
  >("all");
  const [parokiRoleFilter, setParokiRoleFilter] = useState<
    "all" | "admin" | "umatAdmin" | "liturgiAdmin" | "member"
  >("all");

  const availableScopes = getAvailableUserScopes(me);
  const allowed = canManageUsers(me);
  const requestedScope = searchParams.get("scope") as UserDirectoryScope | null;
  const activeScope: UserDirectoryScope =
    requestedScope && availableScopes.includes(requestedScope)
      ? requestedScope
      : (availableScopes[0] ?? "global");

  const activeParokiId = me?.activeParokiId ?? "";

  function selectScope(scope: UserDirectoryScope) {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (scope === availableScopes[0]) {
          next.delete("scope");
        } else {
          next.set("scope", scope);
        }
        return next;
      },
      { replace: true },
    );
  }

  const globalFilter =
    statusFilter === "all" && cantusAdminFilter === "all"
      ? undefined
      : {
          ...(statusFilter !== "all" ? { status: statusFilter } : {}),
          ...(cantusAdminFilter === "cantusAdmin" ? { cantusAdmin: true } : {}),
        };

  const usersQuery = useUsersSearch(
    {
      q: debouncedQ || undefined,
      filter: globalFilter,
    },
    Boolean(me && allowed && activeScope === "global"),
  );

  const parokiMembersQuery = useDirectoryParokiMembers(
    activeParokiId,
    {
      q: debouncedQ || undefined,
      filter:
        parokiRoleFilter === "all"
          ? undefined
          : parokiRoleFilter === "umatAdmin"
            ? { umatAdmin: true }
            : parokiRoleFilter === "liturgiAdmin"
              ? { liturgiAdmin: true }
              : { role: parokiRoleFilter },
    },
    Boolean(me && allowed && activeScope === "paroki" && activeParokiId),
  );

  const orgUsersQuery = useDirectoryManagedOrgUsers(
    activeParokiId,
    debouncedQ,
    Boolean(me && allowed && activeScope === "org" && activeParokiId),
  );

  if (me && !allowed) {
    return (
      <div className="mx-auto flex w-full max-w-2xl flex-col items-center gap-3 px-4 py-16 text-center">
        <span className="flex size-12 items-center justify-center rounded-xl bg-danger-bg text-danger-fg">
          <ShieldAlert aria-hidden className="size-6" />
        </span>
        <h1 className="text-xl font-bold text-fg">
          {t("users.forbiddenTitle")}
        </h1>
        <p className="text-sm text-muted">{t("users.forbiddenDescription")}</p>
      </div>
    );
  }

  const needsActiveParoki =
    Boolean(me) && activeScope !== "global" && !activeParokiId;

  const rows: UnifiedUserRow[] = (() => {
    if (activeScope === "global") {
      const users = usersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
      return users
        .filter((u): u is typeof u & { id: string } => Boolean(u.id))
        .map((u) => {
          const superAdmin = isSuperAdmin(u);
          const cantusAdmin = isCantusAdmin(u);
          const providers = u.linkedProviders?.map(providerLabel) ?? [];
          return {
            id: u.id,
            name: u.name ?? "",
            email: u.email ?? "",
            phone: u.phone,
            status: u.status,
            highlight: superAdmin || cantusAdmin,
            isSuperAdmin: superAdmin,
            isCantusAdmin: cantusAdmin,
            onboarded: u.onboarded,
            metaLine: [
              u.email,
              u.phone,
              providers.length > 0 ? providers.join(", ") : undefined,
            ]
              .filter(Boolean)
              .join(" · "),
          };
        });
    }

    if (activeScope === "paroki") {
      const members =
        parokiMembersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
      return members
        .filter((m): m is typeof m & { userId: string } => Boolean(m.userId))
        .map((m) => ({
          id: m.userId,
          name: m.name ?? "",
          email: m.email ?? "",
          phone: m.phone,
          highlight:
            m.role === "admin" ||
            Boolean(m.umatAdmin) ||
            Boolean(m.liturgiAdmin),
          parishRole: m.role,
          umatAdmin: Boolean(m.umatAdmin),
          liturgiAdmin: Boolean(m.liturgiAdmin),
          metaLine: [m.email, m.phone].filter(Boolean).join(" · "),
        }));
    }

    return orgUsersQuery.users.map((u) => ({
      id: u.id,
      name: u.name,
      email: u.email,
      phone: u.phone,
      highlight: u.orgs.some((o) => o.role === "admin"),
      metaLine: [
        u.email,
        u.phone,
        u.orgs
          .map((o) =>
            o.role === "admin"
              ? `${o.orgName} (${t("users.roleOrgAdmin")})`
              : o.orgName,
          )
          .join(", "),
      ]
        .filter(Boolean)
        .join(" · "),
    }));
  })();

  const isLoading =
    !me ||
    (activeScope === "global"
      ? usersQuery.isLoading
      : activeScope === "paroki"
        ? Boolean(activeParokiId) && parokiMembersQuery.isLoading
        : Boolean(activeParokiId) && orgUsersQuery.isLoading);

  const hasNextPage =
    activeScope === "global"
      ? usersQuery.hasNextPage
      : activeScope === "paroki"
        ? parokiMembersQuery.hasNextPage
        : false;
  const isFetchingNextPage =
    activeScope === "global"
      ? usersQuery.isFetchingNextPage
      : activeScope === "paroki"
        ? parokiMembersQuery.isFetchingNextPage
        : false;
  const fetchNextPage =
    activeScope === "global"
      ? () => usersQuery.fetchNextPage()
      : () => parokiMembersQuery.fetchNextPage();

  const scopeLabels: Record<UserDirectoryScope, string> = {
    global: t("users.scopeGlobal"),
    paroki: t("users.scopeParish"),
    org: t("users.scopeOrg"),
  };

  const statusOptions: Array<{
    value: "all" | "active" | "disabled";
    label: string;
  }> = [
    { value: "all", label: t("users.statusAll") },
    { value: "active", label: t("users.statusActive") },
    { value: "disabled", label: t("users.statusDisabled") },
  ];

  const parishRoleOptions: Array<{
    value: "all" | "admin" | "umatAdmin" | "liturgiAdmin" | "member";
    label: string;
  }> = [
    { value: "all", label: t("users.roleAll") },
    { value: "admin", label: t("users.roleParishAdmin") },
    { value: "umatAdmin", label: t("users.roleUmatAdmin") },
    { value: "liturgiAdmin", label: t("users.roleLiturgiAdmin") },
    { value: "member", label: t("users.roleMember") },
  ];

  const createHref = `${USER_ROUTES.new}?scope=${activeScope}`;

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("me.homeBreadcrumb"), to: "/" },
          { label: t("users.title") },
        ]}
        title={t("users.title")}
        subtitle={t("users.subtitle")}
        action={
          allowed && !needsActiveParoki ? (
            <Link to={createHref} className={buttonClass("primary")}>
              <Plus aria-hidden className="size-4" />
              {t("users.addUser")}
            </Link>
          ) : undefined
        }
      />

      <InductiveFilterBar className="items-center justify-between">
        <div className="relative min-w-60 flex-1">
          <Search
            aria-hidden
            className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-muted"
          />
          <input
            type="search"
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder={t("users.searchPlaceholder")}
            aria-label={t("users.searchPlaceholder")}
            className="h-11 w-full rounded-lg border border-line bg-canvas pr-3.5 pl-10 text-sm text-fg placeholder:text-muted"
          />
        </div>

        {availableScopes.length > 1 && (
          <div className="flex flex-wrap items-center gap-1.5">
            {availableScopes.map((scope) => (
              <button
                key={scope}
                type="button"
                onClick={() => selectScope(scope)}
                className={`min-h-10 rounded-full px-3.5 py-2 text-xs font-semibold transition-colors ${
                  activeScope === scope
                    ? "bg-primary text-on-primary"
                    : "border border-line bg-surface text-muted hover:text-fg"
                }`}
              >
                {scopeLabels[scope]}
              </button>
            ))}
          </div>
        )}

        {activeScope === "global" && (
          <div className="flex flex-wrap items-center gap-1.5">
            {statusOptions.map((opt) => (
              <button
                key={opt.value}
                type="button"
                onClick={() => setStatusFilter(opt.value)}
                className={`min-h-10 rounded-full px-3.5 py-2 text-xs font-semibold transition-colors ${
                  statusFilter === opt.value
                    ? "bg-primary text-on-primary"
                    : "border border-line bg-surface text-muted hover:text-fg"
                }`}
              >
                {opt.label}
              </button>
            ))}
            <button
              type="button"
              onClick={() =>
                setCantusAdminFilter((prev) =>
                  prev === "cantusAdmin" ? "all" : "cantusAdmin",
                )
              }
              className={`min-h-10 rounded-full px-3.5 py-2 text-xs font-semibold transition-colors ${
                cantusAdminFilter === "cantusAdmin"
                  ? "bg-primary text-on-primary"
                  : "border border-line bg-surface text-muted hover:text-fg"
              }`}
            >
              {t("users.filterCantusAdminOnly")}
            </button>
          </div>
        )}

        {activeScope === "paroki" && !needsActiveParoki && (
          <div className="flex flex-wrap items-center gap-1.5">
            {parishRoleOptions.map((opt) => (
              <button
                key={opt.value}
                type="button"
                onClick={() => setParokiRoleFilter(opt.value)}
                className={`min-h-10 rounded-full px-3.5 py-2 text-xs font-semibold transition-colors ${
                  parokiRoleFilter === opt.value
                    ? "bg-primary text-on-primary"
                    : "border border-line bg-surface text-muted hover:text-fg"
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
        )}
      </InductiveFilterBar>

      {needsActiveParoki ? (
        <Card className="p-6 text-sm text-muted">
          {t("users.noActiveParish")}
        </Card>
      ) : (
        <>
          {isLoading ? (
            <Card className="p-6">
              <p className="text-sm text-muted">{t("auth.loading")}</p>
            </Card>
          ) : rows.length === 0 ? (
            <InductiveEmptyState
              icon={Users}
              title={t("users.empty")}
              description={t("users.emptyDescription")}
              action={
                <Link to={createHref} className={buttonClass("primary")}>
                  <Plus aria-hidden className="size-4" />
                  {t("users.addUser")}
                </Link>
              }
            />
          ) : (
            <Card className="divide-y divide-line">
              {rows.map((u) => {
                const isSelf = u.id === me?.id;
                const preview: UserDirectoryPreview = {
                  id: u.id,
                  name: u.name,
                  email: u.email,
                  phone: u.phone,
                };
                return (
                  <Link
                    key={u.id}
                    to={USER_ROUTES.detail(u.id)}
                    state={{ userPreview: preview }}
                    className="flex items-center justify-between gap-4 p-4 transition-colors hover:bg-surface-2/50 sm:px-6"
                  >
                    <div className="flex min-w-0 items-center gap-3.5">
                      <UserAvatar
                        name={u.name}
                        email={u.email}
                        highlight={u.highlight}
                      />
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-1.5">
                          <span className="truncate text-sm font-bold text-fg">
                            {u.name || u.email}
                          </span>
                          {isSelf && (
                            <Badge tone="neutral">{t("users.youBadge")}</Badge>
                          )}
                          {u.status && (
                            <Badge
                              tone={
                                u.status === "active" ? "success" : "danger"
                              }
                            >
                              {u.status === "active"
                                ? t("users.statusActive")
                                : t("users.statusDisabled")}
                            </Badge>
                          )}
                          {u.isSuperAdmin && (
                            <Badge tone="warning">
                              {t("users.roleSuperAdmin")}
                            </Badge>
                          )}
                          {u.isCantusAdmin && (
                            <Badge tone="success">
                              {t("users.roleCantusAdmin")}
                            </Badge>
                          )}
                          {u.parishRole && (
                            <Badge
                              tone={
                                u.parishRole === "admin" ? "warning" : "neutral"
                              }
                            >
                              {u.parishRole === "admin"
                                ? t("users.roleParishAdmin")
                                : t("users.roleMember")}
                            </Badge>
                          )}
                          {u.umatAdmin && (
                            <Badge tone="success">
                              {t("users.roleUmatAdmin")}
                            </Badge>
                          )}
                          {u.liturgiAdmin && (
                            <Badge tone="success">
                              {t("users.roleLiturgiAdmin")}
                            </Badge>
                          )}
                          {u.onboarded === false && (
                            <Badge tone="neutral">
                              {t("users.notOnboarded")}
                            </Badge>
                          )}
                        </div>
                        {u.metaLine && (
                          <p className="mt-0.5 truncate text-xs text-muted">
                            {u.metaLine}
                          </p>
                        )}
                      </div>
                    </div>

                    <span className="inline-flex shrink-0 items-center gap-1.5">
                      <span
                        className={`${buttonClass("outline")} hidden h-9 px-3 text-xs sm:inline-flex`}
                      >
                        {t("users.manageUser")}
                        <ChevronRight aria-hidden className="size-3.5" />
                      </span>
                      <ChevronRight
                        aria-hidden
                        className="size-5 text-muted sm:hidden"
                      />
                    </span>
                  </Link>
                );
              })}
            </Card>
          )}

          {hasNextPage && (
            <div className="flex justify-center">
              <Button
                variant="outline"
                disabled={isFetchingNextPage}
                onClick={() => fetchNextPage()}
              >
                {t("users.loadMore")}
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
