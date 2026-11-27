import { ArrowLeft, Plus, ShieldAlert } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Button, buttonClass } from "~/shared/ui/Button";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useAddUserOrgMember } from "../api/actions";
import { useMe, useUserDetail, useUserOrgMemberships } from "../api/queries";
import {
  canManageUsers,
  isParokiAdmin,
  isSuperAdmin,
  USER_ROUTES,
} from "../model";
import type { Route } from "./+types/UserOrgAddPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("users.addToOrg")} · Domus` }];
}

export default function UserOrgAddPage() {
  const { id = "" } = useParams();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: me } = useMe();

  const allowed = canManageUsers(me);
  const superAdminCaller = isSuperAdmin(me);
  const parokiAdminCaller = isParokiAdmin(me);
  const activeParokiId = me?.activeParokiId ?? "";

  const userQuery = useUserDetail(id, Boolean(id && superAdminCaller));
  const user = userQuery.data;

  const { orgs, memberships, isLoading } = useUserOrgMemberships(
    id,
    user?.email,
    activeParokiId,
    Boolean(
      id &&
        allowed &&
        activeParokiId &&
        (superAdminCaller ? user?.email : true),
    ),
  );

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
  const [role, setRole] = useState<"admin" | "member">("member");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const addOrgMember = useAddUserOrgMember(activeParokiId);

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
      setRole("admin");
    }
  }, [canOnlyAppointAdmin]);

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
        <Link to={USER_ROUTES.index} className={buttonClass("outline")}>
          <ArrowLeft aria-hidden className="size-4" />
          {t("users.backToUsers")}
        </Link>
      </div>
    );
  }

  const detailUrl = `${USER_ROUTES.detail(id)}?tab=org`;
  const userLabel = user?.name || user?.email || t("users.detailFallbackTitle");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!id || !selectedOrgId) return;
    setErrorMsg(null);
    try {
      await addOrgMember.mutateAsync({
        orgId: selectedOrgId,
        body: {
          userId: id,
          role: canOnlyAppointAdmin ? "admin" : role,
        },
      });
      navigate(detailUrl, { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        setErrorMsg(err.details[0]?.message ?? err.message);
      }
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("me.homeBreadcrumb"), to: "/" },
          { label: t("users.title"), to: USER_ROUTES.index },
          { label: userLabel, to: detailUrl },
          { label: t("users.addToOrg") },
        ]}
        title={t("users.addToOrg")}
        subtitle={t("users.addToOrgSubtitle")}
      />

      <InductiveFormSection
        title={t("users.orgSection")}
        description={t("users.orgSectionHint")}
      >
        {errorMsg && (
          <div
            role="alert"
            className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
          >
            {errorMsg}
          </div>
        )}

        {!activeParokiId ? (
          <p className="text-sm text-muted">{t("users.noActiveParish")}</p>
        ) : isLoading ? (
          <p className="text-sm text-muted">{t("auth.loading")}</p>
        ) : availableOrgs.length === 0 ? (
          <div className="flex flex-col gap-4">
            <p className="text-sm text-muted">
              {hasAnyManageableOrg
                ? t("users.allOrgsJoined")
                : t("users.orgManagePermissionHint")}
            </p>
            <div>
              <Link to={detailUrl} className={buttonClass("outline")}>
                <ArrowLeft aria-hidden className="size-4" />
                {t("users.cancel")}
              </Link>
            </div>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="grid gap-4 sm:grid-cols-2">
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
                value={canOnlyAppointAdmin ? "admin" : role}
                disabled={canOnlyAppointAdmin}
                onChange={(e) => setRole(e.target.value as "admin" | "member")}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg disabled:opacity-60"
              >
                {!canOnlyAppointAdmin && (
                  <option value="member">{t("users.roleMember")}</option>
                )}
                <option value="admin">{t("users.roleOrgAdmin")}</option>
              </select>
            </label>

            {canOnlyAppointAdmin && (
              <p className="text-xs text-muted sm:col-span-2">
                {t("users.parishAdminAppointHint")}
              </p>
            )}

            <div className="flex flex-col-reverse gap-2 pt-2 sm:col-span-2 sm:flex-row sm:justify-end">
              <Link
                to={detailUrl}
                className={`${buttonClass("outline")} w-full sm:w-auto`}
              >
                {t("users.cancel")}
              </Link>
              <Button
                type="submit"
                disabled={!selectedOrgId || addOrgMember.isPending}
                className="w-full sm:w-auto"
              >
                <Plus aria-hidden className="size-4" />
                {t("users.addToOrg")}
              </Button>
            </div>
          </form>
        )}
      </InductiveFormSection>
    </div>
  );
}
