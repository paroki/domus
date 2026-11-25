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
  useAddParishMember,
  useAdminDeleteUser,
  useAdminUpdateUser,
  useRemoveParishMember,
  useUpdateParishMember,
} from "../api/actions";
import {
  useDirectoryManagedOrgUsers,
  useDirectoryParishMembers,
  useMe,
  useUserDetail,
  useUserOrgMemberships,
  useUserParishMemberships,
} from "../api/queries";
import { ConfirmDialog, UserAvatar } from "../components/DirectoryShared";
import { UserOrgTab } from "../components/UserOrgTab";
import {
  canManageUsers,
  getAvailableUserDetailTabs,
  isOrgAdmin,
  isParishAdmin,
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
  | { kind: "demoteParishAdmin"; parishId: string; parishName: string }
  | { kind: "removeParishMember"; parishId: string; parishName: string };

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
  const parishAdminCaller = isParishAdmin(me);
  const orgAdminCaller = isOrgAdmin(me);
  const activeParishId = me?.activeParokiId ?? "";

  const navPreview = (
    location.state as { userPreview?: UserDirectoryPreview } | null
  )?.userPreview;
  const matchedPreview =
    navPreview && navPreview.id === id ? navPreview : undefined;

  const userQuery = useUserDetail(id, superAdminCaller);

  const parishFallbackQuery = useDirectoryParishMembers(
    activeParishId,
    { limit: 100 },
    Boolean(
      !superAdminCaller &&
        !matchedPreview?.email &&
        parishAdminCaller &&
        activeParishId &&
        id,
    ),
  );
  const orgFallbackQuery = useDirectoryManagedOrgUsers(
    activeParishId,
    "",
    Boolean(
      !superAdminCaller &&
        !matchedPreview?.email &&
        !parishAdminCaller &&
        orgAdminCaller &&
        activeParishId &&
        id,
    ),
  );

  const resolvedPreview = (() => {
    if (matchedPreview) return matchedPreview;
    if (parishAdminCaller) {
      const m = parishFallbackQuery.data?.pages
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
      ((parishAdminCaller && parishFallbackQuery.isLoading) ||
        (orgAdminCaller && orgFallbackQuery.isLoading));

  const {
    parishes: allParishes,
    memberships: allParishMemberships,
    isLoading: parishesLoading,
  } = useUserParishMemberships(
    user?.id,
    user?.email,
    Boolean((superAdminCaller || parishAdminCaller) && user?.id && user?.email),
  );

  const parishes = superAdminCaller
    ? allParishes
    : allParishes.filter((p) => p.id === activeParishId);
  const parishMemberships = superAdminCaller
    ? allParishMemberships
    : allParishMemberships.filter((m) => m.parish.id === activeParishId);

  const {
    orgs,
    memberships: orgMemberships,
    isLoading: orgsLoading,
  } = useUserOrgMemberships(
    user?.id,
    user?.email,
    activeParishId,
    Boolean(allowedCaller && user?.id && user?.email && activeParishId),
  );

  const updateUser = useAdminUpdateUser();
  const deleteUser = useAdminDeleteUser();
  const addParishMember = useAddParishMember();
  const updateParishMember = useUpdateParishMember();
  const removeParishMember = useRemoveParishMember();

  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [selectedParishId, setSelectedParishId] = useState("");
  const [newParishRole, setNewParishRole] = useState<"admin" | "member">(
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

  const joinedParishIds = new Set(parishMemberships.map((m) => m.parish.id));
  const availableParishes = parishes.filter((p) => !joinedParishIds.has(p.id));

  useEffect(() => {
    if (!availableParishes.length) {
      setSelectedParishId("");
      return;
    }
    if (!availableParishes.some((p) => p.id === selectedParishId)) {
      setSelectedParishId(availableParishes[0]?.id ?? "");
    }
  }, [availableParishes, selectedParishId]);

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
  const activeParishName = parishes.find((p) => p.id === activeParishId)?.nama;

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

  async function handleAddToParish(e: FormEvent) {
    e.preventDefault();
    if (!selectedParishId) return;
    setFeedback(null);
    try {
      await addParishMember.mutateAsync({
        parishId: selectedParishId,
        body: {
          userId,
          role: newParishRole,
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

  async function executeUpdateParishRole(
    parishId: string,
    role: "admin" | "member",
  ) {
    setFeedback(null);
    try {
      await updateParishMember.mutateAsync({
        parishId,
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
      if (action.kind === "demoteParishAdmin") {
        await updateParishMember.mutateAsync({
          parishId: action.parishId,
          userId,
          body: { role: "member" },
        });
        setFeedback({
          tone: "success",
          message: t("users.parishMemberUpdated", { name: userDisplayName }),
        });
        return;
      }
      if (action.kind === "removeParishMember") {
        await removeParishMember.mutateAsync({
          parishId: action.parishId,
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
      case "demoteParishAdmin":
        return {
          title: t("users.confirmDemoteParishAdminTitle"),
          description: `${t("users.confirmDemoteParishAdminDescription", {
            name: userDisplayName,
          })} (${pendingConfirm.parishName})`,
          confirmLabel: t("users.demoteParishAdmin"),
        };
      case "removeParishMember":
        return {
          title: t("users.confirmDeleteTitle"),
          description: `${t("users.confirmDeleteDescription", {
            name: userDisplayName,
          })} (${pendingConfirm.parishName})`,
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
      id: "parish",
      label: t("users.tabParish"),
      icon: Church,
      count: parishesLoading ? undefined : parishMemberships.length,
    },
    {
      id: "org",
      label: t("users.tabOrg"),
      icon: Building2,
      count: activeParishId && !orgsLoading ? orgMemberships.length : undefined,
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

      {activeTab === "parish" && (
        <div
          role="tabpanel"
          id="user-tabpanel-parish"
          aria-labelledby="user-tab-parish"
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
              {parishesLoading ? (
                <p className="p-4 text-sm text-muted">{t("auth.loading")}</p>
              ) : parishMemberships.length === 0 ? (
                <p className="p-4 text-sm text-muted">
                  {t("users.noParishesJoined")}
                </p>
              ) : (
                parishMemberships.map(({ parish, member }) => {
                  const isAdmin = member.role === "admin";
                  return (
                    <div
                      key={parish.id}
                      className="flex flex-wrap items-center justify-between gap-3 p-4"
                    >
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="text-sm font-bold text-fg">
                          {parish.nama}
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
                            disabled={updateParishMember.isPending}
                            onClick={() =>
                              parish.id &&
                              setPendingConfirm({
                                kind: "demoteParishAdmin",
                                parishId: parish.id,
                                parishName: parish.nama ?? "",
                              })
                            }
                          >
                            {t("users.demoteParishAdmin")}
                          </Button>
                        ) : (
                          <Button
                            variant="outline"
                            className="h-9 px-3 text-xs"
                            disabled={updateParishMember.isPending}
                            onClick={() =>
                              parish.id &&
                              executeUpdateParishRole(parish.id, "admin")
                            }
                          >
                            <ShieldCheck aria-hidden className="size-3.5" />
                            {t("users.promoteParishAdmin")}
                          </Button>
                        )}

                        <Button
                          variant="danger"
                          className="h-9 px-3"
                          disabled={removeParishMember.isPending}
                          onClick={() =>
                            parish.id &&
                            setPendingConfirm({
                              kind: "removeParishMember",
                              parishId: parish.id,
                              parishName: parish.nama ?? "",
                            })
                          }
                          aria-label={`${t("users.deleteUser")} ${parish.nama ?? ""}`}
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
              onSubmit={handleAddToParish}
              className="mt-5 flex flex-wrap items-end gap-3 border-t border-line pt-5"
            >
              {availableParishes.length === 0 && !parishesLoading ? (
                <p className="text-sm text-muted">
                  {t("users.allParishesJoined")}
                </p>
              ) : (
                <>
                  <label className="flex min-w-52 flex-1 flex-col gap-1.5 text-sm font-medium text-fg">
                    {t("users.selectParish")}
                    <select
                      value={selectedParishId}
                      onChange={(e) => setSelectedParishId(e.target.value)}
                      className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                    >
                      {availableParishes.map((p) => (
                        <option key={p.id} value={p.id}>
                          {p.nama}
                        </option>
                      ))}
                    </select>
                  </label>

                  <label className="flex min-w-44 flex-col gap-1.5 text-sm font-medium text-fg">
                    {t("users.roleLabel")}
                    <select
                      value={newParishRole}
                      onChange={(e) =>
                        setNewParishRole(e.target.value as "admin" | "member")
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
                    disabled={!selectedParishId || addParishMember.isPending}
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
            activeParishId={activeParishId}
            activeParishName={activeParishName}
            parishAdminCaller={parishAdminCaller}
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
          updateParishMember.isPending ||
          removeParishMember.isPending
        }
        onConfirm={executeConfirm}
        onCancel={() => setPendingConfirm(null)}
      />
    </div>
  );
}
