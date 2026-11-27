import { ArrowLeft, Plus, ShieldAlert } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useSearchParams } from "react-router";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Button, buttonClass } from "~/shared/ui/Button";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import {
  useAddParokiMember,
  useAddUserOrgMember,
  useAdminCreateUser,
} from "../api/actions";
import { useDirectoryOrgs, useDirectoryParoki, useMe } from "../api/queries";
import {
  canManageUsers,
  getAvailableUserScopes,
  isParokiAdmin,
  isSuperAdmin,
  USER_ROUTES,
  type UserDirectoryScope,
} from "../model";
import type { Route } from "./+types/UserCreatePage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("users.createTitle")} · Domus` }];
}

export default function UserCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { data: me } = useMe();

  const availableScopes = getAvailableUserScopes(me);
  const allowed = canManageUsers(me);
  const superAdminCaller = isSuperAdmin(me);
  const parokiAdminCaller = isParokiAdmin(me);
  const canDelegateParokiRole = superAdminCaller || parokiAdminCaller;
  const activeParokiId = me?.activeParokiId ?? "";

  const requestedScope = searchParams.get("scope") as UserDirectoryScope | null;
  const activeScope: UserDirectoryScope =
    requestedScope && availableScopes.includes(requestedScope)
      ? requestedScope
      : (availableScopes[0] ?? "global");

  const parokiQuery = useDirectoryParoki(
    Boolean(me && activeScope === "paroki" && superAdminCaller),
  );
  const orgsQuery = useDirectoryOrgs(
    activeParokiId,
    Boolean(me && activeScope === "org" && activeParokiId),
  );

  const manageableOrgs = (orgsQuery.data ?? []).filter(
    (o) => Boolean(o.id) && (o.myRole === "admin" || parokiAdminCaller),
  );

  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [grantSuperAdmin, setGrantSuperAdmin] = useState(false);
  const [grantCantusAdmin, setGrantCantusAdmin] = useState(false);

  const [selectedParokiId, setSelectedParokiId] = useState(activeParokiId);
  const [parokiRole, setParokiRole] = useState<"admin" | "member">("member");
  const [umatAdmin, setUmatAdmin] = useState(false);
  const [liturgiAdmin, setLiturgiAdmin] = useState(false);

  const [selectedOrgId, setSelectedOrgId] = useState("");
  const [orgRole, setOrgRole] = useState<"admin" | "member">("member");

  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const createGlobalUser = useAdminCreateUser();
  const addParokiMember = useAddParokiMember();
  const addOrgMember = useAddUserOrgMember(activeParokiId);

  useEffect(() => {
    if (superAdminCaller && parokiQuery.data?.length) {
      if (!parokiQuery.data.some((p) => p.id === selectedParokiId)) {
        setSelectedParokiId(parokiQuery.data[0]?.id ?? "");
      }
    } else if (!superAdminCaller && activeParokiId) {
      setSelectedParokiId(activeParokiId);
    }
  }, [superAdminCaller, parokiQuery.data, activeParokiId, selectedParokiId]);

  useEffect(() => {
    if (!manageableOrgs.length) {
      setSelectedOrgId("");
      return;
    }
    if (!manageableOrgs.some((o) => o.id === selectedOrgId)) {
      setSelectedOrgId(manageableOrgs[0]?.id ?? "");
    }
  }, [manageableOrgs, selectedOrgId]);

  const selectedOrg = manageableOrgs.find((o) => o.id === selectedOrgId);
  const canOnlyAppointOrgAdmin =
    Boolean(selectedOrg) &&
    selectedOrg?.myRole !== "admin" &&
    parokiAdminCaller;

  useEffect(() => {
    if (canOnlyAppointOrgAdmin) {
      setOrgRole("admin");
    }
  }, [canOnlyAppointOrgAdmin]);

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

  function selectScope(scope: UserDirectoryScope) {
    setErrorMsg(null);
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.set("scope", scope);
        return next;
      },
      { replace: true },
    );
  }

  const scopeLabels: Record<UserDirectoryScope, string> = {
    global: t("users.scopeGlobal"),
    paroki: t("users.scopeParish"),
    org: t("users.scopeOrg"),
  };

  const pending =
    createGlobalUser.isPending ||
    addParokiMember.isPending ||
    addOrgMember.isPending;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);

    const trimmedEmail = email.trim();
    const trimmedName = name.trim();
    const trimmedPhone = phone.trim();

    try {
      if (activeScope === "global") {
        const created = await createGlobalUser.mutateAsync({
          email: trimmedEmail,
          name: trimmedName,
          ...(trimmedPhone ? { phone: trimmedPhone } : {}),
          ...(grantSuperAdmin ? { isSuperAdmin: true } : {}),
          ...(grantCantusAdmin ? { isCantusAdmin: true } : {}),
        });
        navigate(
          created.id ? USER_ROUTES.detail(created.id) : USER_ROUTES.index,
          { replace: true },
        );
        return;
      }

      if (activeScope === "paroki") {
        const targetParokiId = superAdminCaller
          ? selectedParokiId
          : activeParokiId;
        if (!targetParokiId) return;
        const created = await addParokiMember.mutateAsync({
          parokiId: targetParokiId,
          body: {
            email: trimmedEmail,
            ...(trimmedName ? { name: trimmedName } : {}),
            ...(trimmedPhone ? { phone: trimmedPhone } : {}),
            role: canDelegateParokiRole ? parokiRole : "member",
            ...(umatAdmin ? { umatAdmin: true } : {}),
            ...(canDelegateParokiRole && liturgiAdmin
              ? { liturgiAdmin: true }
              : {}),
          },
        });
        navigate(
          created.userId
            ? `${USER_ROUTES.detail(created.userId)}?tab=paroki`
            : USER_ROUTES.index,
          { replace: true },
        );
        return;
      }

      if (!selectedOrgId) return;
      const created = await addOrgMember.mutateAsync({
        orgId: selectedOrgId,
        body: {
          email: trimmedEmail,
          ...(trimmedName ? { name: trimmedName } : {}),
          ...(trimmedPhone ? { phone: trimmedPhone } : {}),
          role: canOnlyAppointOrgAdmin ? "admin" : orgRole,
        },
      });
      navigate(
        created.userId
          ? `${USER_ROUTES.detail(created.userId)}?tab=org`
          : USER_ROUTES.index,
        { replace: true },
      );
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
          { label: t("users.createTitle") },
        ]}
        title={t("users.createTitle")}
        subtitle={t("users.createSubtitle")}
      />

      <InductiveFormSection
        title={scopeLabels[activeScope]}
        description={
          activeScope === "global"
            ? t("users.superAdminHint")
            : t("users.existingEmailHint")
        }
      >
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

        {errorMsg && (
          <div
            role="alert"
            className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
          >
            {errorMsg}
          </div>
        )}

        <form onSubmit={handleSubmit} className="grid gap-4 sm:grid-cols-2">
          {activeScope === "paroki" && superAdminCaller && (
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
              {t("users.selectParish")}
              <select
                value={selectedParokiId}
                onChange={(e) => setSelectedParokiId(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                {(parokiQuery.data ?? []).map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.nama}
                  </option>
                ))}
              </select>
            </label>
          )}

          {activeScope === "org" && (
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
              {t("users.selectOrg")}
              <select
                value={selectedOrgId}
                onChange={(e) => setSelectedOrgId(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                {manageableOrgs.map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.name}
                  </option>
                ))}
              </select>
            </label>
          )}

          <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
            {t("users.emailLabel")}
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="nama@paroki.id"
              className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
            />
          </label>

          <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
            {activeScope === "global"
              ? t("users.nameLabel")
              : t("users.nameOptionalLabel")}
            <input
              type="text"
              required={activeScope === "global"}
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

          {activeScope === "global" && (
            <div className="flex flex-wrap items-center gap-4 sm:col-span-2">
              <label className="flex h-11 items-center gap-2 text-sm font-medium text-fg">
                <input
                  type="checkbox"
                  checked={grantSuperAdmin}
                  onChange={(e) => setGrantSuperAdmin(e.target.checked)}
                  className="size-4 rounded border-line"
                />
                {t("users.superAdminLabel")}
              </label>
              <label className="flex h-11 items-center gap-2 text-sm font-medium text-fg">
                <input
                  type="checkbox"
                  checked={grantCantusAdmin}
                  onChange={(e) => setGrantCantusAdmin(e.target.checked)}
                  className="size-4 rounded border-line"
                />
                {t("users.roleCantusAdmin")}
              </label>
            </div>
          )}

          {activeScope === "paroki" && (
            <>
              {canDelegateParokiRole && (
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
                  {t("users.roleLabel")}
                  <select
                    value={parokiRole}
                    onChange={(e) =>
                      setParokiRole(e.target.value as "admin" | "member")
                    }
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                  >
                    <option value="member">{t("users.roleMember")}</option>
                    <option value="admin">{t("users.roleParishAdmin")}</option>
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
            </>
          )}

          {activeScope === "org" && (
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
              {t("users.roleLabel")}
              <select
                value={canOnlyAppointOrgAdmin ? "admin" : orgRole}
                disabled={canOnlyAppointOrgAdmin}
                onChange={(e) =>
                  setOrgRole(e.target.value as "admin" | "member")
                }
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg disabled:opacity-60"
              >
                {!canOnlyAppointOrgAdmin && (
                  <option value="member">{t("users.roleMember")}</option>
                )}
                <option value="admin">{t("users.roleOrgAdmin")}</option>
              </select>
            </label>
          )}

          <div className="flex flex-col-reverse gap-2 pt-2 sm:col-span-2 sm:flex-row sm:justify-end">
            <Link
              to={USER_ROUTES.index}
              className={`${buttonClass("outline")} w-full sm:w-auto`}
            >
              {t("users.cancel")}
            </Link>
            <Button
              type="submit"
              disabled={
                pending ||
                (activeScope === "paroki" && !selectedParokiId) ||
                (activeScope === "org" && !selectedOrgId)
              }
              className="w-full sm:w-auto"
            >
              <Plus aria-hidden className="size-4" />
              {activeScope === "global"
                ? t("users.createSubmit")
                : activeScope === "paroki"
                  ? t("users.saveParishMember")
                  : t("users.saveOrgMember")}
            </Button>
          </div>
        </form>
      </InductiveFormSection>
    </div>
  );
}
