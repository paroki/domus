import { Building2, Pencil, Plus, ShieldCheck, Trash2 } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  useAddUserOrgMember,
  useRemoveUserOrgMember,
  useUpdateUserOrgMember,
} from "../api/actions";
import { useDirectoryLookups, useDirectoryPeriods } from "../api/queries";
import type { Organization, OrgMember } from "../api/types";
import type { DirectoryFeedback } from "../pages/UserPage";
import { ConfirmDialog } from "./DirectoryShared";

export function UserOrgTab({
  userId,
  userDisplayName,
  activeParokiId,
  activeParokiName,
  parokiAdminCaller,
  orgs,
  memberships,
  isLoading,
  onFeedback,
}: {
  userId: string;
  userDisplayName: string;
  activeParokiId: string;
  activeParokiName?: string;
  parokiAdminCaller: boolean;
  orgs: Organization[];
  memberships: Array<{ org: Organization; member: OrgMember }>;
  isLoading: boolean;
  onFeedback: (fb: DirectoryFeedback | null) => void;
}) {
  const { t } = useTranslation();

  const lookupsQuery = useDirectoryLookups(
    "org_position",
    Boolean(activeParokiId),
  );
  const periodsQuery = useDirectoryPeriods(
    activeParokiId,
    Boolean(activeParokiId),
  );
  const positionLookups = lookupsQuery.data ?? [];
  const periods = periodsQuery.data ?? [];

  const addOrgMember = useAddUserOrgMember(activeParokiId);
  const updateOrgMember = useUpdateUserOrgMember(activeParokiId);
  const removeOrgMember = useRemoveUserOrgMember(activeParokiId);

  const joinedOrgIds = new Set(memberships.map((m) => m.org.id));
  const availableOrgs = orgs.filter(
    (o) =>
      o.id &&
      !joinedOrgIds.has(o.id) &&
      (o.myRole === "admin" || parokiAdminCaller),
  );
  const hasAnyManageableOrg =
    parokiAdminCaller || orgs.some((o) => o.myRole === "admin");

  const [selectedOrgId, setSelectedOrgId] = useState("");
  const [newRole, setNewRole] = useState<"admin" | "member">("member");
  const [newPosition, setNewPosition] = useState("");
  const [newPeriodId, setNewPeriodId] = useState("");

  const [editingOrgId, setEditingOrgId] = useState<string | null>(null);
  const [editPosition, setEditPosition] = useState("");
  const [editPeriodId, setEditPeriodId] = useState("");

  const [pendingRemove, setPendingRemove] = useState<{
    orgId: string;
    orgName: string;
  } | null>(null);

  useEffect(() => {
    if (!availableOrgs.length) {
      setSelectedOrgId("");
      return;
    }
    if (!availableOrgs.some((o) => o.id === selectedOrgId)) {
      setSelectedOrgId(availableOrgs[0]?.id ?? "");
    }
  }, [availableOrgs, selectedOrgId]);

  const selectedOrg = availableOrgs.find((o) => o.id === selectedOrgId);
  const canOnlyAppointAdmin =
    Boolean(selectedOrg) &&
    selectedOrg?.myRole !== "admin" &&
    parokiAdminCaller;

  useEffect(() => {
    if (canOnlyAppointAdmin) {
      setNewRole("admin");
    }
  }, [canOnlyAppointAdmin]);

  if (!activeParokiId) {
    return (
      <Card className="p-6 text-sm text-muted">
        {t("users.noActiveParish")}
      </Card>
    );
  }

  async function handleAddToOrg(e: FormEvent) {
    e.preventDefault();
    if (!selectedOrgId) return;
    onFeedback(null);
    try {
      await addOrgMember.mutateAsync({
        orgId: selectedOrgId,
        body: {
          userId,
          role: canOnlyAppointAdmin ? "admin" : newRole,
          position: newPosition.trim() || undefined,
          periodId: newPeriodId || undefined,
        },
      });
      setNewPosition("");
      setNewPeriodId("");
      onFeedback({
        tone: "success",
        message: t("users.orgMemberAdded"),
      });
    } catch (err) {
      if (err instanceof ApiError) {
        onFeedback({
          tone: "error",
          message: err.details[0]?.message ?? err.message,
        });
      }
    }
  }

  async function handleSavePosition(orgId: string) {
    onFeedback(null);
    try {
      await updateOrgMember.mutateAsync({
        orgId,
        userId,
        body: {
          position: editPosition.trim() || undefined,
          periodId: editPeriodId || undefined,
        },
      });
      setEditingOrgId(null);
      onFeedback({
        tone: "success",
        message: t("users.orgMemberUpdated", { name: userDisplayName }),
      });
    } catch (err) {
      if (err instanceof ApiError) {
        onFeedback({ tone: "error", message: err.message });
      }
    }
  }

  async function handleUpdateRole(orgId: string, nextRole: "admin" | "member") {
    onFeedback(null);
    try {
      await updateOrgMember.mutateAsync({
        orgId,
        userId,
        body: { role: nextRole },
      });
      onFeedback({
        tone: "success",
        message: t("users.orgMemberUpdated", { name: userDisplayName }),
      });
    } catch (err) {
      if (err instanceof ApiError) {
        onFeedback({ tone: "error", message: err.message });
      }
    }
  }

  async function handleToggleStatus(orgId: string, currentStatus: string) {
    onFeedback(null);
    const nextStatus = currentStatus === "active" ? "inactive" : "active";
    try {
      await updateOrgMember.mutateAsync({
        orgId,
        userId,
        body: { status: nextStatus },
      });
      onFeedback({
        tone: "success",
        message: t("users.orgMemberUpdated", { name: userDisplayName }),
      });
    } catch (err) {
      if (err instanceof ApiError) {
        onFeedback({ tone: "error", message: err.message });
      }
    }
  }

  async function executeRemove() {
    if (!pendingRemove) return;
    const target = pendingRemove;
    setPendingRemove(null);
    onFeedback(null);
    try {
      await removeOrgMember.mutateAsync({
        orgId: target.orgId,
        userId,
      });
      onFeedback({
        tone: "success",
        message: t("users.orgMemberRemoved", { name: userDisplayName }),
      });
    } catch (err) {
      if (err instanceof ApiError) {
        onFeedback({ tone: "error", message: err.message });
      }
    }
  }

  return (
    <>
      <Card className="p-6">
        <div className="mb-4 flex flex-wrap items-start justify-between gap-2">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-bold text-fg">
              <Building2 aria-hidden className="size-5 text-primary" />
              {t("users.orgSection")}
            </h2>
            <p className="mt-0.5 text-sm text-muted">
              {t("users.orgSectionHint")}
            </p>
          </div>
          {activeParokiName && (
            <Badge tone="neutral">
              {t("users.activeParishContext", { name: activeParokiName })}
            </Badge>
          )}
        </div>

        <datalist id="user-detail-org-position-options">
          {positionLookups.map((item) => (
            <option key={item.id} value={item.label} />
          ))}
        </datalist>

        <div className="divide-y divide-line rounded-lg border border-line">
          {isLoading ? (
            <p className="p-4 text-sm text-muted">{t("auth.loading")}</p>
          ) : memberships.length === 0 ? (
            <p className="p-4 text-sm text-muted">{t("users.noOrgsJoined")}</p>
          ) : (
            memberships.map(({ org, member }) => {
              const orgId = org.id ?? "";
              const isOrgAdmin = org.myRole === "admin";
              const isMemberAdmin = member.role === "admin";
              const isEditing = editingOrgId === orgId;
              const periodName = periods.find(
                (p) => p.id === member.periodId,
              )?.name;

              return (
                <div key={orgId} className="flex flex-col gap-3 p-4">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="text-sm font-bold text-fg">
                          {org.name}
                        </span>
                        <Badge tone={isMemberAdmin ? "warning" : "neutral"}>
                          {isMemberAdmin
                            ? t("users.roleOrgAdmin")
                            : t("users.roleMember")}
                        </Badge>
                        <Badge
                          tone={
                            member.status === "active" ? "success" : "danger"
                          }
                        >
                          {member.status === "active"
                            ? t("users.statusActive")
                            : t("users.statusInactive")}
                        </Badge>
                      </div>
                      {(member.position || periodName) && (
                        <p className="mt-1 text-xs text-muted">
                          {member.position ?? ""}
                          {member.position && periodName ? " · " : ""}
                          {periodName ?? ""}
                        </p>
                      )}
                    </div>

                    <div className="flex flex-wrap items-center gap-2">
                      {isOrgAdmin ? (
                        <>
                          <Button
                            variant="outline"
                            className="h-9 px-3 text-xs"
                            onClick={() => {
                              setEditingOrgId(isEditing ? null : orgId);
                              setEditPosition(member.position ?? "");
                              setEditPeriodId(member.periodId ?? "");
                            }}
                          >
                            <Pencil aria-hidden className="size-3.5" />
                            {t("users.editPosition")}
                          </Button>
                          <Button
                            variant="outline"
                            className="h-9 px-3 text-xs"
                            disabled={updateOrgMember.isPending}
                            onClick={() =>
                              handleUpdateRole(
                                orgId,
                                isMemberAdmin ? "member" : "admin",
                              )
                            }
                          >
                            {isMemberAdmin
                              ? t("users.demoteOrgAdmin")
                              : t("users.promoteOrgAdmin")}
                          </Button>
                          <Button
                            variant="ghost"
                            className="h-9 px-3 text-xs"
                            disabled={updateOrgMember.isPending}
                            onClick={() =>
                              handleToggleStatus(
                                orgId,
                                member.status ?? "active",
                              )
                            }
                          >
                            {member.status === "active"
                              ? t("users.disableUser")
                              : t("users.enableUser")}
                          </Button>
                          <Button
                            variant="danger"
                            className="h-9 px-3"
                            disabled={removeOrgMember.isPending}
                            onClick={() =>
                              setPendingRemove({
                                orgId,
                                orgName: org.name ?? "",
                              })
                            }
                            aria-label={`${t("users.deleteUser")} ${org.name ?? ""}`}
                          >
                            <Trash2 aria-hidden className="size-4" />
                          </Button>
                        </>
                      ) : (
                        parokiAdminCaller &&
                        !isMemberAdmin && (
                          <Button
                            variant="outline"
                            className="h-9 px-3 text-xs"
                            disabled={updateOrgMember.isPending}
                            onClick={() => handleUpdateRole(orgId, "admin")}
                          >
                            <ShieldCheck aria-hidden className="size-3.5" />
                            {t("users.promoteOrgAdmin")}
                          </Button>
                        )
                      )}
                    </div>
                  </div>

                  {isEditing && (
                    <div className="flex flex-wrap items-end gap-2 rounded-lg border border-line bg-canvas p-3">
                      <input
                        type="text"
                        list="user-detail-org-position-options"
                        value={editPosition}
                        onChange={(e) => setEditPosition(e.target.value)}
                        placeholder={t("users.positionLabel")}
                        aria-label={t("users.positionLabel")}
                        className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                      />
                      <select
                        value={editPeriodId}
                        onChange={(e) => setEditPeriodId(e.target.value)}
                        aria-label={t("users.periodLabel")}
                        className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                      >
                        <option value="">{t("users.noPeriod")}</option>
                        {periods.map((p) => (
                          <option key={p.id} value={p.id}>
                            {p.name}
                          </option>
                        ))}
                      </select>
                      <Button
                        className="h-9 px-4 text-xs"
                        disabled={updateOrgMember.isPending}
                        onClick={() => handleSavePosition(orgId)}
                      >
                        {t("users.save")}
                      </Button>
                      <Button
                        variant="outline"
                        className="h-9 px-4 text-xs"
                        onClick={() => setEditingOrgId(null)}
                      >
                        {t("users.cancel")}
                      </Button>
                    </div>
                  )}
                </div>
              );
            })
          )}
        </div>

        <form
          onSubmit={handleAddToOrg}
          className="mt-5 border-t border-line pt-5"
        >
          {availableOrgs.length === 0 && !isLoading ? (
            <p className="text-sm text-muted">
              {hasAnyManageableOrg
                ? t("users.allOrgsJoined")
                : t("users.orgManagePermissionHint")}
            </p>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("users.selectOrg")}
                <select
                  value={selectedOrgId}
                  onChange={(e) => setSelectedOrgId(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  {availableOrgs.map((o) => (
                    <option key={o.id} value={o.id}>
                      {o.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("users.roleLabel")}
                <select
                  value={canOnlyAppointAdmin ? "admin" : newRole}
                  disabled={canOnlyAppointAdmin}
                  onChange={(e) =>
                    setNewRole(e.target.value as "admin" | "member")
                  }
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg disabled:opacity-60"
                >
                  {!canOnlyAppointAdmin && (
                    <option value="member">{t("users.roleMember")}</option>
                  )}
                  <option value="admin">{t("users.roleOrgAdmin")}</option>
                </select>
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("users.positionLabel")}
                <input
                  type="text"
                  list="user-detail-org-position-options"
                  value={newPosition}
                  onChange={(e) => setNewPosition(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("users.periodLabel")}
                <select
                  value={newPeriodId}
                  onChange={(e) => setNewPeriodId(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  <option value="">{t("users.noPeriod")}</option>
                  {periods.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
              </label>

              {canOnlyAppointAdmin && (
                <p className="text-xs text-muted sm:col-span-2">
                  {t("users.parishAdminAppointHint")}
                </p>
              )}

              <div className="sm:col-span-2">
                <Button
                  type="submit"
                  disabled={!selectedOrgId || addOrgMember.isPending}
                >
                  <Plus aria-hidden className="size-4" />
                  {t("users.addToOrg")}
                </Button>
              </div>
            </div>
          )}
        </form>
      </Card>

      <ConfirmDialog
        open={Boolean(pendingRemove)}
        title={t("users.confirmDeleteTitle")}
        description={`${t("users.confirmDeleteDescription", {
          name: userDisplayName,
        })} (${pendingRemove?.orgName ?? ""})`}
        confirmLabel={t("users.deleteUser")}
        cancelLabel={t("users.cancel")}
        tone="danger"
        pending={removeOrgMember.isPending}
        onConfirm={executeRemove}
        onCancel={() => setPendingRemove(null)}
      />
    </>
  );
}
