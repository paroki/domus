import { ArrowLeft, Plus, ShieldAlert } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Button, buttonClass } from "~/shared/ui/Button";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useAddParokiMember } from "../api/actions";
import { useMe, useUserDetail, useUserParokiMemberships } from "../api/queries";
import {
  canManageUsers,
  isParokiAdmin,
  isSuperAdmin,
  isUmatAdmin,
  USER_ROUTES,
} from "../model";
import type { Route } from "./+types/UserParokiAddPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("users.addToParish")} · Domus` }];
}

export default function UserParokiAddPage() {
  const { id = "" } = useParams();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: me } = useMe();

  const allowed = canManageUsers(me);
  const superAdminCaller = isSuperAdmin(me);
  const parokiAdminCaller = isParokiAdmin(me);
  const umatAdminCaller = isUmatAdmin(me);
  const canAccessParokiDirectory =
    superAdminCaller || parokiAdminCaller || umatAdminCaller;
  const canDelegateParokiRole = superAdminCaller || parokiAdminCaller;
  const activeParokiId = me?.activeParokiId ?? "";

  const userQuery = useUserDetail(id, Boolean(id && superAdminCaller));
  const user = userQuery.data;

  const {
    paroki: allParoki,
    memberships: allMemberships,
    isLoading,
  } = useUserParokiMemberships(
    id,
    user?.email,
    Boolean(
      id && canAccessParokiDirectory && (superAdminCaller ? user?.email : true),
    ),
  );

  const parokiList = superAdminCaller
    ? allParoki
    : allParoki.filter((p) => p.id === activeParokiId);
  const memberships = superAdminCaller
    ? allMemberships
    : allMemberships.filter((m) => m.paroki.id === activeParokiId);
  const joinedParokiIds = new Set(memberships.map((m) => m.paroki.id));
  const availableParoki = parokiList.filter((p) => !joinedParokiIds.has(p.id));

  const [selectedParokiId, setSelectedParokiId] = useState("");
  const [role, setRole] = useState<"admin" | "member">("member");
  const [umatAdmin, setUmatAdmin] = useState(false);
  const [liturgiAdmin, setLiturgiAdmin] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const addParokiMember = useAddParokiMember();

  useEffect(() => {
    if (!availableParoki.length) {
      setSelectedParokiId("");
      return;
    }
    if (!availableParoki.some((p) => p.id === selectedParokiId)) {
      setSelectedParokiId(availableParoki[0]?.id ?? "");
    }
  }, [availableParoki, selectedParokiId]);

  if (me && (!allowed || !canAccessParokiDirectory)) {
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

  const detailUrl = `${USER_ROUTES.detail(id)}?tab=paroki`;
  const userLabel = user?.name || user?.email || t("users.detailFallbackTitle");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!id || !selectedParokiId) return;
    setErrorMsg(null);
    try {
      await addParokiMember.mutateAsync({
        parokiId: selectedParokiId,
        body: {
          userId: id,
          role: canDelegateParokiRole ? role : "member",
          ...(umatAdmin ? { umatAdmin: true } : {}),
          ...(canDelegateParokiRole && liturgiAdmin
            ? { liturgiAdmin: true }
            : {}),
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
          { label: t("users.addToParish") },
        ]}
        title={t("users.addToParish")}
        subtitle={t("users.addToParishSubtitle")}
      />

      <InductiveFormSection
        title={t("users.parishSection")}
        description={t("users.parishSectionHint")}
      >
        {errorMsg && (
          <div
            role="alert"
            className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
          >
            {errorMsg}
          </div>
        )}

        {isLoading ? (
          <p className="text-sm text-muted">{t("auth.loading")}</p>
        ) : availableParoki.length === 0 ? (
          <div className="flex flex-col gap-4">
            <p className="text-sm text-muted">{t("users.allParishesJoined")}</p>
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

            {canDelegateParokiRole && (
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("users.roleLabel")}
                <select
                  value={role}
                  onChange={(e) =>
                    setRole(e.target.value as "admin" | "member")
                  }
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  <option value="admin">{t("users.roleParishAdmin")}</option>
                  <option value="member">{t("users.roleMember")}</option>
                </select>
              </label>
            )}

            <div className="flex flex-wrap items-center gap-4 sm:col-span-2">
              <label className="flex h-11 items-center gap-2 text-sm font-medium text-fg">
                <input
                  type="checkbox"
                  checked={umatAdmin}
                  onChange={(e) => setUmatAdmin(e.target.checked)}
                  className="size-4 rounded border-line"
                />
                {t("users.roleUmatAdmin")}
              </label>

              {canDelegateParokiRole && (
                <label className="flex h-11 items-center gap-2 text-sm font-medium text-fg">
                  <input
                    type="checkbox"
                    checked={liturgiAdmin}
                    onChange={(e) => setLiturgiAdmin(e.target.checked)}
                    className="size-4 rounded border-line"
                  />
                  {t("users.roleLiturgiAdmin")}
                </label>
              )}
            </div>

            <div className="flex flex-col-reverse gap-2 pt-2 sm:col-span-2 sm:flex-row sm:justify-end">
              <Link
                to={detailUrl}
                className={`${buttonClass("outline")} w-full sm:w-auto`}
              >
                {t("users.cancel")}
              </Link>
              <Button
                type="submit"
                disabled={!selectedParokiId || addParokiMember.isPending}
                className="w-full sm:w-auto"
              >
                <Plus aria-hidden className="size-4" />
                {t("users.addToParish")}
              </Button>
            </div>
          </form>
        )}
      </InductiveFormSection>
    </div>
  );
}
