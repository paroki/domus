import { Building2, Plus, ShieldCheck, Trash2 } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import { useRemoveUserOrgMember, useUpdateUserOrgMember } from "../api/actions";
import type { Organization, OrgMember } from "../api/types";
import { USER_ROUTES } from "../model";
import type { DirectoryFeedback } from "../pages/UserPage";

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

  const [pendingRemove, setPendingRemove] = useState<{
    orgId: string;
    orgName: string;
  } | null>(null);

  if (!activeParokiId) {
    return (
      <Card className="p-6 text-sm text-muted">
        {t("users.noActiveParish")}
      </Card>
    );
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
        <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-bold text-fg">
              <Building2 aria-hidden className="size-5 text-primary" />
              {t("users.orgSection")}
            </h2>
            <p className="mt-0.5 text-sm text-muted">
              {t("users.orgSectionHint")}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {activeParokiName && (
              <Badge tone="neutral">
                {t("users.activeParishContext", { name: activeParokiName })}
              </Badge>
            )}
            {availableOrgs.length > 0 && (
              <Link
                to={USER_ROUTES.orgNew(userId)}
                className={buttonClass("primary")}
              >
                <Plus aria-hidden className="size-4" />
                {t("users.addToOrg")}
              </Link>
            )}
          </div>
        </div>

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
                    </div>

                    <div className="flex flex-wrap items-center gap-2">
                      {isOrgAdmin ? (
                        <>
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
                </div>
              );
            })
          )}
        </div>

        {availableOrgs.length === 0 && !isLoading && (
          <p className="mt-4 text-sm text-muted">
            {hasAnyManageableOrg
              ? t("users.allOrgsJoined")
              : t("users.orgManagePermissionHint")}
          </p>
        )}
      </Card>

      <ConfirmDialog
        open={Boolean(pendingRemove)}
        title={t("users.confirmDeleteTitle")}
        description={`${t("users.confirmDeleteDescription", {
          name: userDisplayName,
        })} (${pendingRemove?.orgName ?? ""})`}
        confirmLabel={t("users.deleteUser")}
        cancelLabel={t("users.cancel")}
        confirmVariant="danger"
        pending={removeOrgMember.isPending}
        onConfirm={executeRemove}
        onCancel={() => setPendingRemove(null)}
      />
    </>
  );
}
