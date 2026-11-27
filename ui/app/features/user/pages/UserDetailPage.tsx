import {
  ArrowLeft,
  Building2,
  CheckCircle2,
  Church,
  Pencil,
  Plus,
  ShieldAlert,
  ShieldCheck,
  Trash2,
  UserCog,
  X,
} from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Link,
  useLocation,
  useNavigate,
  useParams,
  useSearchParams,
} from "react-router";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import { InductiveDangerZone, InductivePageHeader } from "~/shared/ui/iui";
import {
  useAdminDeleteUser,
  useAdminUpdateUser,
  useRemoveParokiMember,
  useUpdateParokiMember,
} from "../api/actions";
import {
  useDirectoryManagedOrgUsers,
  useDirectoryParokiMembers,
  useMe,
  useUserDetail,
  useUserOrgMemberships,
  useUserParokiMemberships,
} from "../api/queries";
import { UserAvatar } from "../components/DirectoryShared";
import { UserOrgTab } from "../components/UserOrgTab";
import {
  canManageUsers,
  getAvailableUserDetailTabs,
  isCantusAdmin,
  isOrgAdmin,
  isParokiAdmin,
  isSuperAdmin,
  isUmatAdmin,
  parseUserDetailTab,
  providerLabel,
  USER_ROUTES,
  type UserDetailTab,
} from "../model";
import type { Route } from "./+types/UserDetailPage";
import type { DirectoryFeedback, UserDirectoryPreview } from "./UserPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.user.name")} · Domus` }];
}

type PendingConfirm =
  | { kind: "deleteUser" }
  | { kind: "demoteSuperAdmin" }
  | { kind: "revokeCantusAdmin" }
  | { kind: "disableUser" }
  | { kind: "demoteParokiAdmin"; parokiId: string; parokiName: string }
  | { kind: "revokeUmatAdmin"; parokiId: string; parokiName: string }
  | { kind: "revokeLiturgiAdmin"; parokiId: string; parokiName: string }
  | { kind: "removeParokiMember"; parokiId: string; parokiName: string };

function formatDate(iso?: string | null): string | null {
  if (!iso) return null;
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return null;
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(d);
}

export default function UserDetailPage() {
  const { id } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { t } = useTranslation();

  const { data: me } = useMe();
  const allowedCaller = canManageUsers(me);
  const superAdminCaller = isSuperAdmin(me);
  const parokiAdminCaller = isParokiAdmin(me);
  const umatAdminCaller = isUmatAdmin(me);
  const canAccessParokiDirectory =
    superAdminCaller || parokiAdminCaller || umatAdminCaller;
  const canDelegateParokiRole = superAdminCaller || parokiAdminCaller;
  const orgAdminCaller = isOrgAdmin(me);
  const activeParokiId = me?.activeParokiId ?? "";

  const navPreview = (
    location.state as { userPreview?: UserDirectoryPreview } | null
  )?.userPreview;
  const matchedPreview =
    navPreview && navPreview.id === id ? navPreview : undefined;

  const userQuery = useUserDetail(id, superAdminCaller);

  const parokiFallbackQuery = useDirectoryParokiMembers(
    activeParokiId,
    { limit: 100 },
    Boolean(
      !superAdminCaller &&
        !matchedPreview?.email &&
        (parokiAdminCaller || umatAdminCaller) &&
        activeParokiId &&
        id,
    ),
  );
  const orgFallbackQuery = useDirectoryManagedOrgUsers(
    activeParokiId,
    "",
    Boolean(
      !superAdminCaller &&
        !matchedPreview?.email &&
        !parokiAdminCaller &&
        !umatAdminCaller &&
        orgAdminCaller &&
        activeParokiId &&
        id,
    ),
  );

  const resolvedPreview = (() => {
    if (matchedPreview) return matchedPreview;
    if (parokiAdminCaller || umatAdminCaller) {
      const m = parokiFallbackQuery.data?.pages
        .flatMap((p) => p.data ?? [])
        .find((item) => item.userId === id);
      if (m?.userId) {
        return {
          id: m.userId,
          name: m.name ?? "",
          email: m.email ?? "",
          phone: m.phone,
        };
      }
    }
    if (orgAdminCaller) {
      const u = orgFallbackQuery.users.find((item) => item.id === id);
      if (u) {
        return {
          id: u.id,
          name: u.name,
          email: u.email,
          phone: u.phone,
        };
      }
    }
    return undefined;
  })();

  const user = superAdminCaller ? userQuery.data : resolvedPreview;
  const isUserLoading = superAdminCaller
    ? userQuery.isLoading
    : !resolvedPreview &&
      (((parokiAdminCaller || umatAdminCaller) &&
        parokiFallbackQuery.isLoading) ||
        (orgAdminCaller && orgFallbackQuery.isLoading));

  const {
    paroki: allParoki,
    memberships: allParokiMemberships,
    isLoading: parokiLoading,
  } = useUserParokiMemberships(
    user?.id,
    user?.email,
    Boolean(canAccessParokiDirectory && user?.id && user?.email),
  );

  const parokiList = superAdminCaller
    ? allParoki
    : allParoki.filter((p) => p.id === activeParokiId);
  const parokiMemberships = superAdminCaller
    ? allParokiMemberships
    : allParokiMemberships.filter((m) => m.paroki.id === activeParokiId);

  const {
    orgs,
    memberships: orgMemberships,
    isLoading: orgsLoading,
  } = useUserOrgMemberships(
    user?.id,
    user?.email,
    activeParokiId,
    Boolean(allowedCaller && user?.id && user?.email && activeParokiId),
  );

  const updateUser = useAdminUpdateUser();
  const deleteUser = useAdminDeleteUser();
  const updateParokiMember = useUpdateParokiMember();
  const removeParokiMember = useRemoveParokiMember();

  const [editingProfile, setEditingProfile] = useState(false);
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [feedback, setFeedback] = useState<DirectoryFeedback | null>(null);
  const [pendingConfirm, setPendingConfirm] = useState<PendingConfirm | null>(
    null,
  );

  const allowedTabs = getAvailableUserDetailTabs(me?.roles);
  const activeTab = parseUserDetailTab(searchParams.get("tab"), allowedTabs);

  function selectTab(tab: UserDetailTab) {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (tab === allowedTabs[0]) {
          next.delete("tab");
        } else {
          next.set("tab", tab);
        }
        return next;
      },
      { replace: true },
    );
  }

  useEffect(() => {
    if (!user) return;
    setName(user.name ?? "");
    setPhone(user.phone ?? "");
  }, [user]);

  const joinedParokiIds = new Set(parokiMemberships.map((m) => m.paroki.id));
  const availableParoki = parokiList.filter((p) => !joinedParokiIds.has(p.id));

  useEffect(() => {
    if (feedback?.tone !== "success") return;
    const timer = window.setTimeout(() => setFeedback(null), 4000);
    return () => window.clearTimeout(timer);
  }, [feedback]);

  if (me && !allowedCaller) {
    return (
      <div className="mx-auto flex w-full max-w-2xl flex-col items-center gap-3 px-4 py-16 text-center">
        <span className="flex size-12 items-center justify-center rounded-xl bg-danger-bg text-danger-fg">
          <ShieldAlert aria-hidden className="size-6" />
        </span>
        <h1 className="text-xl font-bold text-fg">
          {t("users.forbiddenTitle")}
        </h1>
        <p className="text-sm text-muted">{t("users.forbiddenDescription")}</p>
        <Link to={USER_ROUTES.index} className={buttonClass("outline")}>
          <ArrowLeft aria-hidden className="size-4" />
          {t("users.backToUsers")}
        </Link>
      </div>
    );
  }

  if (isUserLoading) {
    return (
      <div className="mx-auto w-full max-w-4xl px-4 py-8 sm:px-6">
        <p className="text-sm text-muted">{t("auth.loading")}</p>
      </div>
    );
  }

  if (!user?.id) {
    return (
      <div className="mx-auto flex w-full max-w-2xl flex-col items-center gap-4 px-4 py-16 text-center">
        <p className="text-base font-semibold text-fg">
          {t("users.userNotFound")}
        </p>
        <Link to={USER_ROUTES.index} className={buttonClass("outline")}>
          <ArrowLeft aria-hidden className="size-4" />
          {t("users.backToUsers")}
        </Link>
      </div>
    );
  }

  const userId = user.id;
  const userDisplayName = user.name || user.email || "";
  const fullUser = superAdminCaller ? userQuery.data : undefined;
  const targetIsSuperAdmin = isSuperAdmin(fullUser);
  const targetIsCantusAdmin = isCantusAdmin(fullUser);
  const isSelf = userId === me?.id;
  const providers = fullUser?.linkedProviders?.map(providerLabel) ?? [];
  const lastLoginText = formatDate(fullUser?.lastLoginAt);
  const createdAtText = formatDate(fullUser?.createdAt);
  const activeParokiName = parokiList.find(
    (p) => p.id === activeParokiId,
  )?.nama;

  async function handleSaveProfile(e: FormEvent) {
    e.preventDefault();
    setFeedback(null);
    const trimmedName = name.trim();
    try {
      await updateUser.mutateAsync({
        id: userId,
        body: {
          name: trimmedName,
          phone: phone.trim(),
        },
      });
      setEditingProfile(false);
      setFeedback({
        tone: "success",
        message: t("users.userUpdated", { name: trimmedName }),
      });
    } catch (err) {
      if (err instanceof ApiError) {
        setFeedback({
          tone: "error",
          message: err.details[0]?.message ?? err.message,
        });
      }
    }
  }

  async function executeUpdateParokiRole(
    parokiId: string,
    body: {
      role?: "admin" | "member";
      umatAdmin?: boolean;
      liturgiAdmin?: boolean;
    },
  ) {
    setFeedback(null);
    try {
      await updateParokiMember.mutateAsync({
        parokiId,
        userId,
        body,
      });
      setFeedback({
        tone: "success",
        message: t("users.parishMemberUpdated", { name: userDisplayName }),
      });
    } catch (err) {
      if (err instanceof ApiError) {
        setFeedback({ tone: "error", message: err.message });
      }
    }
  }

  async function executeConfirm() {
    if (!pendingConfirm) return;
    const action = pendingConfirm;
    setPendingConfirm(null);
    setFeedback(null);
    try {
      if (action.kind === "deleteUser") {
        await deleteUser.mutateAsync(userId);
        navigate(USER_ROUTES.index, { replace: true });
        return;
      }
      if (action.kind === "demoteSuperAdmin") {
        await updateUser.mutateAsync({
          id: userId,
          body: { isSuperAdmin: false },
        });
        setFeedback({
          tone: "success",
          message: t("users.userUpdated", { name: userDisplayName }),
        });
        return;
      }
      if (action.kind === "revokeCantusAdmin") {
        await updateUser.mutateAsync({
          id: userId,
          body: { isCantusAdmin: false },
        });
        setFeedback({
          tone: "success",
          message: t("users.userUpdated", { name: userDisplayName }),
        });
        return;
      }
      if (action.kind === "disableUser") {
        await updateUser.mutateAsync({
          id: userId,
          body: { status: "disabled" },
        });
        setFeedback({
          tone: "success",
          message: t("users.userUpdated", { name: userDisplayName }),
        });
        return;
      }
      if (action.kind === "demoteParokiAdmin") {
        await updateParokiMember.mutateAsync({
          parokiId: action.parokiId,
          userId,
          body: { role: "member" },
        });
        setFeedback({
          tone: "success",
          message: t("users.parishMemberUpdated", { name: userDisplayName }),
        });
        return;
      }
      if (action.kind === "revokeUmatAdmin") {
        await updateParokiMember.mutateAsync({
          parokiId: action.parokiId,
          userId,
          body: { umatAdmin: false },
        });
        setFeedback({
          tone: "success",
          message: t("users.parishMemberUpdated", { name: userDisplayName }),
        });
        return;
      }
      if (action.kind === "revokeLiturgiAdmin") {
        await updateParokiMember.mutateAsync({
          parokiId: action.parokiId,
          userId,
          body: { liturgiAdmin: false },
        });
        setFeedback({
          tone: "success",
          message: t("users.parishMemberUpdated", { name: userDisplayName }),
        });
        return;
      }
      if (action.kind === "removeParokiMember") {
        await removeParokiMember.mutateAsync({
          parokiId: action.parokiId,
          userId,
        });
        setFeedback({
          tone: "success",
          message: t("users.parishMemberRemoved", { name: userDisplayName }),
        });
      }
    } catch (err) {
      if (err instanceof ApiError) {
        setFeedback({ tone: "error", message: err.message });
      }
    }
  }

  const confirmProps = (() => {
    if (!pendingConfirm) return null;
    switch (pendingConfirm.kind) {
      case "deleteUser":
        return {
          title: t("users.confirmDeleteTitle"),
          description: t("users.confirmDeleteDescription", {
            name: userDisplayName,
          }),
          confirmLabel: t("users.deleteUser"),
        };
      case "demoteSuperAdmin":
        return {
          title: t("users.confirmDemoteSuperAdminTitle"),
          description: t("users.confirmDemoteSuperAdminDescription", {
            name: userDisplayName,
          }),
          confirmLabel: t("users.demoteSuperAdmin"),
        };
      case "revokeCantusAdmin":
        return {
          title: t("users.confirmRevokeCantusAdminTitle"),
          description: t("users.confirmRevokeCantusAdminDescription", {
            name: userDisplayName,
          }),
          confirmLabel: t("users.revokeCantusAdmin"),
        };
      case "disableUser":
        return {
          title: t("users.confirmDisableUserTitle"),
          description: t("users.confirmDisableUserDescription", {
            name: userDisplayName,
          }),
          confirmLabel: t("users.disableUser"),
        };
      case "demoteParokiAdmin":
        return {
          title: t("users.confirmDemoteParishAdminTitle"),
          description: `${t("users.confirmDemoteParishAdminDescription", {
            name: userDisplayName,
          })} (${pendingConfirm.parokiName})`,
          confirmLabel: t("users.demoteParishAdmin"),
        };
      case "revokeUmatAdmin":
        return {
          title: t("users.confirmRevokeUmatAdminTitle"),
          description: `${t("users.confirmRevokeUmatAdminDescription", {
            name: userDisplayName,
          })} (${pendingConfirm.parokiName})`,
          confirmLabel: t("users.revokeUmatAdmin"),
        };
      case "revokeLiturgiAdmin":
        return {
          title: t("users.confirmRevokeLiturgiAdminTitle"),
          description: `${t("users.confirmRevokeLiturgiAdminDescription", {
            name: userDisplayName,
          })} (${pendingConfirm.parokiName})`,
          confirmLabel: t("users.revokeLiturgiAdmin"),
        };
      case "removeParokiMember":
        return {
          title: t("users.confirmDeleteTitle"),
          description: `${t("users.confirmDeleteDescription", {
            name: userDisplayName,
          })} (${pendingConfirm.parokiName})`,
          confirmLabel: t("users.deleteUser"),
        };
    }
  })();

  const allTabs: Array<{
    id: UserDetailTab;
    label: string;
    icon: typeof UserCog;
    count?: number;
  }> = [
    { id: "profile", label: t("users.tabProfile"), icon: UserCog },
    {
      id: "paroki",
      label: t("users.tabParish"),
      icon: Church,
      count: parokiLoading ? undefined : parokiMemberships.length,
    },
    {
      id: "org",
      label: t("users.tabOrg"),
      icon: Building2,
      count: activeParokiId && !orgsLoading ? orgMemberships.length : undefined,
    },
  ];
  const tabs = allTabs.filter((tab) => allowedTabs.includes(tab.id));

  const headerSubtitle = [
    user.email,
    user.phone,
    providers.length > 0 ? providers.join(", ") : undefined,
  ]
    .filter(Boolean)
    .join(" · ");

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("me.homeBreadcrumb"), to: "/" },
          { label: t("users.title"), to: USER_ROUTES.index },
          { label: userDisplayName },
        ]}
        title={userDisplayName}
        subtitle={headerSubtitle}
      />

      {feedback && (
        <div
          role={feedback.tone === "error" ? "alert" : "status"}
          className={`flex items-center justify-between gap-3 rounded-lg px-4 py-3 text-sm font-medium ${
            feedback.tone === "error"
              ? "bg-danger-bg text-danger-fg"
              : "bg-success-bg text-success-fg"
          }`}
        >
          <span className="flex items-center gap-2">
            {feedback.tone === "success" ? (
              <CheckCircle2 aria-hidden className="size-4 shrink-0" />
            ) : (
              <ShieldAlert aria-hidden className="size-4 shrink-0" />
            )}
            {feedback.message}
          </span>
          <button
            type="button"
            onClick={() => setFeedback(null)}
            aria-label={t("users.dismiss")}
            className="rounded p-1 opacity-75 transition-opacity hover:opacity-100"
          >
            <X aria-hidden className="size-4" />
          </button>
        </div>
      )}

      <Card className="p-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="flex min-w-0 items-center gap-4">
            <UserAvatar
              name={user.name}
              email={user.email}
              highlight={targetIsSuperAdmin || targetIsCantusAdmin}
            />
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <span className="truncate text-lg font-bold text-fg">
                  {user.name}
                </span>
                {isSelf && <Badge tone="neutral">{t("users.youBadge")}</Badge>}
                {fullUser?.status && (
                  <Badge
                    tone={fullUser.status === "active" ? "success" : "danger"}
                  >
                    {fullUser.status === "active"
                      ? t("users.statusActive")
                      : t("users.statusDisabled")}
                  </Badge>
                )}
                {targetIsSuperAdmin && (
                  <Badge tone="warning">{t("users.roleSuperAdmin")}</Badge>
                )}
                {targetIsCantusAdmin && (
                  <Badge tone="success">{t("users.roleCantusAdmin")}</Badge>
                )}
                {fullUser?.onboarded === false && (
                  <Badge tone="neutral">{t("users.notOnboarded")}</Badge>
                )}
              </div>
              {(lastLoginText || createdAtText) && (
                <p className="mt-1 text-xs text-muted">
                  {createdAtText
                    ? `${t("users.createdAt")}: ${createdAtText}`
                    : ""}
                  {createdAtText && lastLoginText ? " · " : ""}
                  {lastLoginText
                    ? `${t("users.lastLogin")}: ${lastLoginText}`
                    : ""}
                </p>
              )}
            </div>
          </div>
        </div>
      </Card>

      <div
        role="tablist"
        aria-label={t("users.detailTabsLabel")}
        className="flex gap-1 overflow-x-auto border-b border-line"
      >
        {tabs.map((tab) => {
          const Icon = tab.icon;
          const selected = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              type="button"
              role="tab"
              id={`user-tab-${tab.id}`}
              aria-selected={selected}
              aria-controls={`user-tabpanel-${tab.id}`}
              onClick={() => selectTab(tab.id)}
              className={`-mb-px inline-flex shrink-0 items-center gap-2 border-b-2 px-4 py-3 text-sm font-semibold transition-colors ${
                selected
                  ? "border-primary text-primary"
                  : "border-transparent text-muted hover:border-line hover:text-fg"
              }`}
            >
              <Icon aria-hidden className="size-4" />
              <span>{tab.label}</span>
              {tab.count !== undefined && (
                <span
                  className={`rounded-full px-2 py-0.5 text-xs font-semibold ${
                    selected
                      ? "bg-primary/15 text-primary"
                      : "bg-surface-2 text-muted"
                  }`}
                >
                  {tab.count}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {activeTab === "profile" && (
        <div
          role="tabpanel"
          id="user-tabpanel-profile"
          aria-labelledby="user-tab-profile"
          className="flex flex-col gap-6"
        >
          <Card className="p-6">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h2 className="flex items-center gap-2 text-lg font-bold text-fg">
                <UserCog aria-hidden className="size-5 text-primary" />
                {t("users.profileSection")}
              </h2>
              {!editingProfile && (
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setEditingProfile(true)}
                  className="h-9 px-3 text-xs"
                >
                  <Pencil aria-hidden className="size-3.5" />
                  {t("users.editUser")}
                </Button>
              )}
            </div>

            {editingProfile ? (
              <form
                onSubmit={handleSaveProfile}
                className="grid gap-4 sm:grid-cols-2"
              >
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  {t("users.nameLabel")}
                  <input
                    type="text"
                    required
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  {t("users.phoneLabel")}
                  <input
                    type="tel"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    placeholder="+628123456789"
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                  />
                </label>
                <div className="flex flex-col-reverse gap-2 sm:col-span-2 sm:flex-row sm:justify-end">
                  <Button
                    type="button"
                    variant="outline"
                    disabled={updateUser.isPending}
                    onClick={() => {
                      setName(user.name ?? "");
                      setPhone(user.phone ?? "");
                      setEditingProfile(false);
                    }}
                    className="w-full sm:w-auto"
                  >
                    {t("users.cancel")}
                  </Button>
                  <Button
                    type="submit"
                    disabled={updateUser.isPending}
                    className="w-full sm:w-auto"
                  >
                    {t("users.saveChanges")}
                  </Button>
                </div>
              </form>
            ) : (
              <dl className="grid gap-4 sm:grid-cols-2">
                <div>
                  <dt className="text-xs font-medium text-muted">
                    {t("users.nameLabel")}
                  </dt>
                  <dd className="mt-1 text-sm font-medium text-fg">
                    {user.name || "—"}
                  </dd>
                </div>
                <div>
                  <dt className="text-xs font-medium text-muted">
                    {t("users.emailLabel")}
                  </dt>
                  <dd className="mt-1 text-sm font-medium text-fg">
                    {user.email || "—"}
                  </dd>
                </div>
                <div>
                  <dt className="text-xs font-medium text-muted">
                    {t("users.phoneLabel")}
                  </dt>
                  <dd className="mt-1 text-sm font-medium text-fg">
                    {user.phone || "—"}
                  </dd>
                </div>
                <div>
                  <dt className="text-xs font-medium text-muted">
                    {t("users.lastLogin")}
                  </dt>
                  <dd className="mt-1 text-sm font-medium text-fg">
                    {lastLoginText || "—"}
                  </dd>
                </div>
              </dl>
            )}
          </Card>

          <Card className="p-6">
            <h2 className="mb-4 flex items-center gap-2 text-lg font-bold text-fg">
              <ShieldCheck aria-hidden className="size-5 text-primary" />
              {t("users.accessSection")}
            </h2>

            <div className="flex flex-wrap items-center gap-2">
              <Button
                variant="outline"
                disabled={
                  updateUser.isPending || (isSelf && targetIsSuperAdmin)
                }
                title={
                  isSelf && targetIsSuperAdmin
                    ? t("users.selfProtectedHint")
                    : undefined
                }
                onClick={async () => {
                  if (targetIsSuperAdmin) {
                    setPendingConfirm({ kind: "demoteSuperAdmin" });
                    return;
                  }
                  setFeedback(null);
                  try {
                    await updateUser.mutateAsync({
                      id: userId,
                      body: { isSuperAdmin: true },
                    });
                    setFeedback({
                      tone: "success",
                      message: t("users.userUpdated", {
                        name: userDisplayName,
                      }),
                    });
                  } catch (err) {
                    if (err instanceof ApiError) {
                      setFeedback({ tone: "error", message: err.message });
                    }
                  }
                }}
              >
                <ShieldCheck aria-hidden className="size-4" />
                {targetIsSuperAdmin
                  ? t("users.demoteSuperAdmin")
                  : t("users.promoteSuperAdmin")}
              </Button>

              <Button
                variant="outline"
                disabled={updateUser.isPending}
                onClick={async () => {
                  if (targetIsCantusAdmin) {
                    setPendingConfirm({ kind: "revokeCantusAdmin" });
                    return;
                  }
                  setFeedback(null);
                  try {
                    await updateUser.mutateAsync({
                      id: userId,
                      body: { isCantusAdmin: true },
                    });
                    setFeedback({
                      tone: "success",
                      message: t("users.userUpdated", {
                        name: userDisplayName,
                      }),
                    });
                  } catch (err) {
                    if (err instanceof ApiError) {
                      setFeedback({ tone: "error", message: err.message });
                    }
                  }
                }}
              >
                <ShieldCheck aria-hidden className="size-4" />
                {targetIsCantusAdmin
                  ? t("users.revokeCantusAdmin")
                  : t("users.grantCantusAdmin")}
              </Button>
            </div>
          </Card>

          <InductiveDangerZone
            title={t("users.dangerZoneTitle")}
            description={t("users.dangerZoneDescription")}
            action={
              <div className="flex flex-wrap items-center gap-2">
                <Button
                  variant="outline"
                  disabled={
                    updateUser.isPending ||
                    (isSelf && fullUser?.status === "active")
                  }
                  title={
                    isSelf && fullUser?.status === "active"
                      ? t("users.selfProtectedHint")
                      : undefined
                  }
                  onClick={async () => {
                    if (fullUser?.status === "active") {
                      setPendingConfirm({ kind: "disableUser" });
                      return;
                    }
                    setFeedback(null);
                    try {
                      await updateUser.mutateAsync({
                        id: userId,
                        body: { status: "active" },
                      });
                      setFeedback({
                        tone: "success",
                        message: t("users.userUpdated", {
                          name: userDisplayName,
                        }),
                      });
                    } catch (err) {
                      if (err instanceof ApiError) {
                        setFeedback({ tone: "error", message: err.message });
                      }
                    }
                  }}
                >
                  {fullUser?.status === "active"
                    ? t("users.disableUser")
                    : t("users.enableUser")}
                </Button>

                <Button
                  variant="danger"
                  disabled={deleteUser.isPending || isSelf}
                  title={isSelf ? t("users.selfProtectedHint") : undefined}
                  onClick={() => setPendingConfirm({ kind: "deleteUser" })}
                >
                  <Trash2 aria-hidden className="size-4" />
                  {t("users.deleteUser")}
                </Button>
              </div>
            }
          />
        </div>
      )}

      {activeTab === "paroki" && (
        <div
          role="tabpanel"
          id="user-tabpanel-paroki"
          aria-labelledby="user-tab-paroki"
        >
          <Card className="p-6">
            <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
              <div>
                <h2 className="flex items-center gap-2 text-lg font-bold text-fg">
                  <Church aria-hidden className="size-5 text-primary" />
                  {t("users.parishSection")}
                </h2>
                <p className="mt-0.5 text-sm text-muted">
                  {t("users.parishSectionHint")}
                </p>
              </div>
              {availableParoki.length > 0 && (
                <Link
                  to={USER_ROUTES.parokiNew(userId)}
                  className={buttonClass("primary")}
                >
                  <Plus aria-hidden className="size-4" />
                  {t("users.addToParish")}
                </Link>
              )}
            </div>

            <div className="divide-y divide-line rounded-lg border border-line">
              {parokiLoading ? (
                <p className="p-4 text-sm text-muted">{t("auth.loading")}</p>
              ) : parokiMemberships.length === 0 ? (
                <p className="p-4 text-sm text-muted">
                  {t("users.noParishesJoined")}
                </p>
              ) : (
                parokiMemberships.map(({ paroki, member }) => {
                  const isAdmin = member.role === "admin";
                  return (
                    <div
                      key={paroki.id}
                      className="flex flex-wrap items-center justify-between gap-3 p-4"
                    >
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="text-sm font-bold text-fg">
                          {paroki.nama}
                        </span>
                        <Badge tone={isAdmin ? "warning" : "neutral"}>
                          {isAdmin
                            ? t("users.roleParishAdmin")
                            : t("users.roleMember")}
                        </Badge>
                        {member.umatAdmin ? (
                          <Badge tone="success">
                            {t("users.roleUmatAdmin")}
                          </Badge>
                        ) : null}
                        {member.liturgiAdmin ? (
                          <Badge tone="success">
                            {t("users.roleLiturgiAdmin")}
                          </Badge>
                        ) : null}
                      </div>

                      <div className="flex flex-wrap items-center gap-2">
                        {member.umatAdmin ? (
                          <Button
                            variant="outline"
                            className="h-9 px-3 text-xs"
                            disabled={updateParokiMember.isPending}
                            onClick={() =>
                              paroki.id &&
                              setPendingConfirm({
                                kind: "revokeUmatAdmin",
                                parokiId: paroki.id,
                                parokiName: paroki.nama ?? "",
                              })
                            }
                          >
                            {t("users.revokeUmatAdmin")}
                          </Button>
                        ) : (
                          <Button
                            variant="outline"
                            className="h-9 px-3 text-xs"
                            disabled={updateParokiMember.isPending}
                            onClick={() =>
                              paroki.id &&
                              executeUpdateParokiRole(paroki.id, {
                                umatAdmin: true,
                              })
                            }
                          >
                            <ShieldCheck aria-hidden className="size-3.5" />
                            {t("users.grantUmatAdmin")}
                          </Button>
                        )}

                        {canDelegateParokiRole &&
                          (member.liturgiAdmin ? (
                            <Button
                              variant="outline"
                              className="h-9 px-3 text-xs"
                              disabled={updateParokiMember.isPending}
                              onClick={() =>
                                paroki.id &&
                                setPendingConfirm({
                                  kind: "revokeLiturgiAdmin",
                                  parokiId: paroki.id,
                                  parokiName: paroki.nama ?? "",
                                })
                              }
                            >
                              {t("users.revokeLiturgiAdmin")}
                            </Button>
                          ) : (
                            <Button
                              variant="outline"
                              className="h-9 px-3 text-xs"
                              disabled={updateParokiMember.isPending}
                              onClick={() =>
                                paroki.id &&
                                executeUpdateParokiRole(paroki.id, {
                                  liturgiAdmin: true,
                                })
                              }
                            >
                              <ShieldCheck aria-hidden className="size-3.5" />
                              {t("users.grantLiturgiAdmin")}
                            </Button>
                          ))}

                        {canDelegateParokiRole &&
                          (isAdmin ? (
                            <Button
                              variant="outline"
                              className="h-9 px-3 text-xs"
                              disabled={updateParokiMember.isPending}
                              onClick={() =>
                                paroki.id &&
                                setPendingConfirm({
                                  kind: "demoteParokiAdmin",
                                  parokiId: paroki.id,
                                  parokiName: paroki.nama ?? "",
                                })
                              }
                            >
                              {t("users.demoteParishAdmin")}
                            </Button>
                          ) : (
                            <Button
                              variant="outline"
                              className="h-9 px-3 text-xs"
                              disabled={updateParokiMember.isPending}
                              onClick={() =>
                                paroki.id &&
                                executeUpdateParokiRole(paroki.id, {
                                  role: "admin",
                                })
                              }
                            >
                              <ShieldCheck aria-hidden className="size-3.5" />
                              {t("users.promoteParishAdmin")}
                            </Button>
                          ))}

                        {canDelegateParokiRole && (
                          <Button
                            variant="danger"
                            className="h-9 px-3"
                            disabled={removeParokiMember.isPending}
                            onClick={() =>
                              paroki.id &&
                              setPendingConfirm({
                                kind: "removeParokiMember",
                                parokiId: paroki.id,
                                parokiName: paroki.nama ?? "",
                              })
                            }
                            aria-label={`${t("users.deleteUser")} ${paroki.nama ?? ""}`}
                          >
                            <Trash2 aria-hidden className="size-4" />
                          </Button>
                        )}
                      </div>
                    </div>
                  );
                })
              )}
            </div>

            {availableParoki.length === 0 && !parokiLoading && (
              <p className="mt-4 text-sm text-muted">
                {t("users.allParishesJoined")}
              </p>
            )}
          </Card>
        </div>
      )}

      {activeTab === "org" && (
        <div
          role="tabpanel"
          id="user-tabpanel-org"
          aria-labelledby="user-tab-org"
        >
          <UserOrgTab
            userId={userId}
            userDisplayName={userDisplayName}
            activeParokiId={activeParokiId}
            activeParokiName={activeParokiName}
            parokiAdminCaller={parokiAdminCaller}
            orgs={orgs}
            memberships={orgMemberships}
            isLoading={orgsLoading}
            onFeedback={setFeedback}
          />
        </div>
      )}

      <ConfirmDialog
        open={Boolean(pendingConfirm)}
        title={confirmProps?.title ?? ""}
        description={confirmProps?.description ?? ""}
        confirmLabel={confirmProps?.confirmLabel ?? ""}
        cancelLabel={t("users.cancel")}
        confirmVariant="danger"
        pending={
          updateUser.isPending ||
          deleteUser.isPending ||
          updateParokiMember.isPending ||
          removeParokiMember.isPending
        }
        onConfirm={executeConfirm}
        onCancel={() => setPendingConfirm(null)}
      />
    </div>
  );
}
