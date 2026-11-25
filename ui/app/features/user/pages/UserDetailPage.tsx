import {
  ArrowLeft,
  Building2,
  CheckCircle2,
  Church,
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
import {
  useAddParokiMember,
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
import { ConfirmDialog, UserAvatar } from "../components/DirectoryShared";
import { UserOrgTab } from "../components/UserOrgTab";
import {
  canManageUsers,
  getAvailableUserDetailTabs,
  isOrgAdmin,
  isParokiAdmin,
  isSuperAdmin,
  parseUserDetailTab,
  providerLabel,
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
  | { kind: "disableUser" }
  | { kind: "demoteParokiAdmin"; parokiId: string; parokiName: string }
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
        parokiAdminCaller &&
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
        orgAdminCaller &&
        activeParokiId &&
        id,
    ),
  );

  const resolvedPreview = (() => {
    if (matchedPreview) return matchedPreview;
    if (parokiAdminCaller) {
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
      ((parokiAdminCaller && parokiFallbackQuery.isLoading) ||
        (orgAdminCaller && orgFallbackQuery.isLoading));

  const {
    paroki: allParoki,
    memberships: allParokiMemberships,
    isLoading: parokiLoading,
  } = useUserParokiMemberships(
    user?.id,
    user?.email,
    Boolean((superAdminCaller || parokiAdminCaller) && user?.id && user?.email),
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
  const addParokiMember = useAddParokiMember();
  const updateParokiMember = useUpdateParokiMember();
  const removeParokiMember = useRemoveParokiMember();

  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [selectedParokiId, setSelectedParokiId] = useState("");
  const [newParokiRole, setNewParokiRole] = useState<"admin" | "member">(
    "admin",
  );
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
    if (!availableParoki.length) {
      setSelectedParokiId("");
      return;
    }
    if (!availableParoki.some((p) => p.id === selectedParokiId)) {
      setSelectedParokiId(availableParoki[0]?.id ?? "");
    }
  }, [availableParoki, selectedParokiId]);

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
        <Link to="/users" className={buttonClass("outline")}>
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
        <Link to="/users" className={buttonClass("outline")}>
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

  async function handleAddToParoki(e: FormEvent) {
    e.preventDefault();
    if (!selectedParokiId) return;
    setFeedback(null);
    try {
      await addParokiMember.mutateAsync({
        parokiId: selectedParokiId,
        body: {
          userId,
          role: newParokiRole,
        },
      });
      setFeedback({
        tone: "success",
        message: t("users.parishMemberAdded"),
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
    role: "admin" | "member",
  ) {
    setFeedback(null);
    try {
      await updateParokiMember.mutateAsync({
        parokiId,
        userId,
        body: { role },
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
        navigate("/users", { replace: true });
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

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-8 sm:px-6">
      <div>
        <Link
          to="/users"
          className="inline-flex items-center gap-1.5 text-sm font-semibold text-muted transition-colors hover:text-primary"
        >
          <ArrowLeft aria-hidden className="size-4" />
          {t("users.backToUsers")}
        </Link>
      </div>

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
              highlight={targetIsSuperAdmin}
            />
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <h1 className="truncate text-xl font-bold text-fg">
                  {user.name}
                </h1>
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
                {fullUser?.onboarded === false && (
                  <Badge tone="neutral">{t("users.notOnboarded")}</Badge>
                )}
              </div>
              <p className="mt-1 text-sm text-muted">
                {user.email}
                {user.phone ? ` · ${user.phone}` : ""}
                {providers.length > 0 ? ` · ${providers.join(", ")}` : ""}
              </p>
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
            <h2 className="mb-4 flex items-center gap-2 text-lg font-bold text-fg">
              <UserCog aria-hidden className="size-5 text-primary" />
              {t("users.profileSection")}
            </h2>

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
              <div className="sm:col-span-2">
                <Button type="submit" disabled={updateUser.isPending}>
                  {t("users.saveChanges")}
                </Button>
              </div>
            </form>
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
          </Card>
        </div>
      )}

      {activeTab === "paroki" && (
        <div
          role="tabpanel"
          id="user-tabpanel-paroki"
          aria-labelledby="user-tab-paroki"
        >
          <Card className="p-6">
            <div className="mb-4">
              <h2 className="flex items-center gap-2 text-lg font-bold text-fg">
                <Church aria-hidden className="size-5 text-primary" />
                {t("users.parishSection")}
              </h2>
              <p className="mt-0.5 text-sm text-muted">
                {t("users.parishSectionHint")}
              </p>
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
                      </div>

                      <div className="flex flex-wrap items-center gap-2">
                        {isAdmin ? (
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
                              executeUpdateParokiRole(paroki.id, "admin")
                            }
                          >
                            <ShieldCheck aria-hidden className="size-3.5" />
                            {t("users.promoteParishAdmin")}
                          </Button>
                        )}

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
                      </div>
                    </div>
                  );
                })
              )}
            </div>

            <form
              onSubmit={handleAddToParoki}
              className="mt-5 flex flex-wrap items-end gap-3 border-t border-line pt-5"
            >
              {availableParoki.length === 0 && !parokiLoading ? (
                <p className="text-sm text-muted">
                  {t("users.allParishesJoined")}
                </p>
              ) : (
                <>
                  <label className="flex min-w-52 flex-1 flex-col gap-1.5 text-sm font-medium text-fg">
                    {t("users.selectParish")}
                    <select
                      value={selectedParokiId}
                      onChange={(e) => setSelectedParokiId(e.target.value)}
                      className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                    >
                      {availableParoki.map((p) => (
                        <option key={p.id} value={p.id}>
                          {p.nama}
                        </option>
                      ))}
                    </select>
                  </label>

                  <label className="flex min-w-44 flex-col gap-1.5 text-sm font-medium text-fg">
                    {t("users.roleLabel")}
                    <select
                      value={newParokiRole}
                      onChange={(e) =>
                        setNewParokiRole(e.target.value as "admin" | "member")
                      }
                      className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                    >
                      <option value="admin">
                        {t("users.roleParishAdmin")}
                      </option>
                      <option value="member">{t("users.roleMember")}</option>
                    </select>
                  </label>

                  <Button
                    type="submit"
                    disabled={!selectedParokiId || addParokiMember.isPending}
                  >
                    <Plus aria-hidden className="size-4" />
                    {t("users.addToParish")}
                  </Button>
                </>
              )}
            </form>
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
        tone="danger"
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
