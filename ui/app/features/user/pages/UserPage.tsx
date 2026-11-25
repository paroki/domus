import { ChevronRight, Search, ShieldAlert } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
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
  isOrgAdmin,
  isParokiAdmin,
  isSuperAdmin,
  providerLabel,
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
  parishRole?: string;
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

  const [q, setQ] = useState("");
  const debouncedQ = useDebouncedValue(q.trim(), 250);
  const [statusFilter, setStatusFilter] = useState<
    "all" | "active" | "disabled"
  >("all");
  const [parokiRoleFilter, setParokiRoleFilter] = useState<
    "all" | "admin" | "member"
  >("all");

  const allowed = canManageUsers(me);
  const superAdminCaller = isSuperAdmin(me);
  const parokiAdminCaller = !superAdminCaller && isParokiAdmin(me);
  const orgAdminCaller =
    !superAdminCaller && !parokiAdminCaller && isOrgAdmin(me);
  const activeParokiId = me?.activeParokiId ?? "";

  const usersQuery = useUsersSearch(
    {
      q: debouncedQ || undefined,
      filter: statusFilter === "all" ? undefined : { status: statusFilter },
    },
    Boolean(me && superAdminCaller),
  );

  const parokiMembersQuery = useDirectoryParokiMembers(
    activeParokiId,
    {
      q: debouncedQ || undefined,
      filter:
        parokiRoleFilter === "all" ? undefined : { role: parokiRoleFilter },
    },
    Boolean(me && parokiAdminCaller && activeParokiId),
  );

  const orgUsersQuery = useDirectoryManagedOrgUsers(
    activeParokiId,
    debouncedQ,
    Boolean(me && orgAdminCaller && activeParokiId),
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

  const needsActiveParoki = Boolean(me) && !superAdminCaller && !activeParokiId;

  const rows: UnifiedUserRow[] = (() => {
    if (superAdminCaller) {
      const users = usersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
      return users
        .filter((u): u is typeof u & { id: string } => Boolean(u.id))
        .map((u) => {
          const superAdmin = isSuperAdmin(u);
          const providers = u.linkedProviders?.map(providerLabel) ?? [];
          return {
            id: u.id,
            name: u.name ?? "",
            email: u.email ?? "",
            phone: u.phone,
            status: u.status,
            highlight: superAdmin,
            isSuperAdmin: superAdmin,
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

    if (parokiAdminCaller) {
      const members =
        parokiMembersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
      return members
        .filter((m): m is typeof m & { userId: string } => Boolean(m.userId))
        .map((m) => ({
          id: m.userId,
          name: m.name ?? "",
          email: m.email ?? "",
          phone: m.phone,
          highlight: m.role === "admin",
          parishRole: m.role,
          metaLine: [m.email, m.phone].filter(Boolean).join(" · "),
        }));
    }

    if (orgAdminCaller) {
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
    }

    return [];
  })();

  const isLoading =
    !me ||
    (superAdminCaller
      ? usersQuery.isLoading
      : parokiAdminCaller
        ? Boolean(activeParokiId) && parokiMembersQuery.isLoading
        : Boolean(activeParokiId) && orgUsersQuery.isLoading);

  const hasNextPage = superAdminCaller
    ? usersQuery.hasNextPage
    : parokiAdminCaller
      ? parokiMembersQuery.hasNextPage
      : false;
  const isFetchingNextPage = superAdminCaller
    ? usersQuery.isFetchingNextPage
    : parokiAdminCaller
      ? parokiMembersQuery.isFetchingNextPage
      : false;
  const fetchNextPage = superAdminCaller
    ? () => usersQuery.fetchNextPage()
    : () => parokiMembersQuery.fetchNextPage();

  const statusOptions: Array<{
    value: "all" | "active" | "disabled";
    label: string;
  }> = [
    { value: "all", label: t("users.statusAll") },
    { value: "active", label: t("users.statusActive") },
    { value: "disabled", label: t("users.statusDisabled") },
  ];

  const parishRoleOptions: Array<{
    value: "all" | "admin" | "member";
    label: string;
  }> = [
    { value: "all", label: t("users.roleAll") },
    { value: "admin", label: t("users.roleParishAdmin") },
    { value: "member", label: t("users.roleMember") },
  ];

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-4 py-8 sm:px-6">
      <div>
        <h1 className="text-2xl font-bold text-fg">{t("users.title")}</h1>
        <p className="text-sm text-muted">{t("users.subtitle")}</p>
      </div>

      {needsActiveParoki ? (
        <Card className="p-6 text-sm text-muted">
          {t("users.noActiveParish")}
        </Card>
      ) : (
        <>
          <div className="flex flex-wrap items-center justify-between gap-3">
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
                className="h-11 w-full rounded-lg border border-line bg-surface pr-3.5 pl-10 text-sm text-fg placeholder:text-muted"
              />
            </div>

            {superAdminCaller && (
              <div className="flex flex-wrap items-center gap-1.5">
                {statusOptions.map((opt) => (
                  <button
                    key={opt.value}
                    type="button"
                    onClick={() => setStatusFilter(opt.value)}
                    className={`rounded-full px-3.5 py-1.5 text-xs font-semibold transition-colors ${
                      statusFilter === opt.value
                        ? "bg-primary text-on-primary"
                        : "border border-line bg-surface text-muted hover:text-fg"
                    }`}
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
            )}

            {parokiAdminCaller && (
              <div className="flex flex-wrap items-center gap-1.5">
                {parishRoleOptions.map((opt) => (
                  <button
                    key={opt.value}
                    type="button"
                    onClick={() => setParokiRoleFilter(opt.value)}
                    className={`rounded-full px-3.5 py-1.5 text-xs font-semibold transition-colors ${
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
          </div>

          <Card className="divide-y divide-line">
            {isLoading ? (
              <p className="p-6 text-sm text-muted">{t("auth.loading")}</p>
            ) : rows.length === 0 ? (
              <p className="p-6 text-sm text-muted">{t("users.empty")}</p>
            ) : (
              rows.map((u) => {
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
                    to={`/users/${u.id}`}
                    state={{ userPreview: preview }}
                    className="flex flex-wrap items-center justify-between gap-4 p-4 transition-colors hover:bg-surface-2/50 sm:px-6"
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

                    <span
                      className={`${buttonClass("outline")} h-9 px-3 text-xs`}
                    >
                      {t("users.manageUser")}
                      <ChevronRight aria-hidden className="size-3.5" />
                    </span>
                  </Link>
                );
              })
            )}
          </Card>

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
