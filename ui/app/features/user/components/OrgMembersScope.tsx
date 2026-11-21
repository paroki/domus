import {
  Pencil,
  Plus,
  Search,
  ShieldCheck,
  Trash2,
  UserPlus,
} from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  useDirectoryAddOrgMember,
  useDirectoryRemoveOrgMember,
  useDirectoryUpdateOrgMember,
} from "../api/actions";
import {
  useDirectoryLookups,
  useDirectoryOrgMembers,
  useDirectoryOrgs,
  useDirectoryPeriods,
} from "../api/queries";
import type { User } from "../api/types";
import { isParishAdmin } from "../model";

export function OrgMembersScope({
  me,
  onError,
}: {
  me: User;
  onError: (msg: string | null) => void;
}) {
  const { t } = useTranslation();
  const parishId = me.activeParishId ?? "";
  const parishAdmin = isParishAdmin(me);

  const orgsQuery = useDirectoryOrgs(parishId, Boolean(parishId));
  const allOrgs = orgsQuery.data ?? [];
  const manageableOrgs = parishAdmin
    ? allOrgs
    : allOrgs.filter((o) => o.myRole === "admin");

  const [selectedOrgId, setSelectedOrgId] = useState("");

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
  const isSelectedOrgAdmin = selectedOrg?.myRole === "admin";
  const canOnlyAppointAdmin = parishAdmin && !isSelectedOrgAdmin;

  const [q, setQ] = useState("");
  const [roleFilter, setRoleFilter] = useState<"all" | "admin" | "member">(
    "all",
  );
  const [statusFilter, setStatusFilter] = useState<
    "all" | "active" | "inactive"
  >("all");
  const [periodFilter, setPeriodFilter] = useState("");

  const [showCreate, setShowCreate] = useState(false);
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [role, setRole] = useState<"member" | "admin">("member");
  const [position, setPosition] = useState("");
  const [periodId, setPeriodId] = useState("");

  const [editingUserId, setEditingUserId] = useState<string | null>(null);
  const [editPosition, setEditPosition] = useState("");
  const [editPeriodId, setEditPeriodId] = useState("");

  useEffect(() => {
    if (canOnlyAppointAdmin) {
      setRole("admin");
    }
  }, [canOnlyAppointAdmin]);

  const lookupsQuery = useDirectoryLookups("org_position", Boolean(parishId));
  const periodsQuery = useDirectoryPeriods(parishId, Boolean(parishId));
  const membersQuery = useDirectoryOrgMembers(
    selectedOrgId,
    {
      q: q.trim() || undefined,
      filter: {
        ...(roleFilter !== "all" ? { role: roleFilter } : {}),
        ...(statusFilter !== "all" ? { status: statusFilter } : {}),
        ...(periodFilter ? { periodId: periodFilter } : {}),
      },
    },
    Boolean(selectedOrgId),
  );

  const addMember = useDirectoryAddOrgMember(selectedOrgId, parishId);
  const updateMember = useDirectoryUpdateOrgMember(selectedOrgId, parishId);
  const removeMember = useDirectoryRemoveOrgMember(selectedOrgId, parishId);

  if (!parishId) {
    return (
      <Card className="p-6 text-sm text-muted">
        {t("users.noActiveParish")}
      </Card>
    );
  }

  if (!orgsQuery.isLoading && manageableOrgs.length === 0) {
    return (
      <Card className="p-6 text-sm text-muted">
        {t("users.noOrgsAvailable")}
      </Card>
    );
  }

  const members = membersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const positionLookups = lookupsQuery.data ?? [];
  const periods = periodsQuery.data ?? [];

  async function handleMakeMeAdmin() {
    if (!selectedOrgId || !me.email) return;
    onError(null);
    try {
      await addMember.mutateAsync({
        email: me.email,
        name: me.name,
        role: "admin",
      });
    } catch (err) {
      if (err instanceof ApiError) onError(err.message);
    }
  }

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    if (!selectedOrgId) return;
    onError(null);
    try {
      await addMember.mutateAsync({
        email: email.trim(),
        name: name.trim() || undefined,
        phone: phone.trim() || undefined,
        role: canOnlyAppointAdmin ? "admin" : role,
        position: position.trim() || undefined,
        periodId: periodId || undefined,
      });
      setEmail("");
      setName("");
      setPhone("");
      setRole(canOnlyAppointAdmin ? "admin" : "member");
      setPosition("");
      setPeriodId("");
      setShowCreate(false);
    } catch (err) {
      if (err instanceof ApiError) {
        onError(err.details[0]?.message ?? err.message);
      }
    }
  }

  async function handleSavePosition(userId: string) {
    onError(null);
    try {
      await updateMember.mutateAsync({
        userId,
        body: {
          position: editPosition.trim() || undefined,
          periodId: editPeriodId || undefined,
        },
      });
      setEditingUserId(null);
    } catch (err) {
      if (err instanceof ApiError) onError(err.message);
    }
  }

  async function handleToggleRole(userId: string, currentRole: string) {
    onError(null);
    const nextRole = currentRole === "admin" ? "member" : "admin";
    try {
      await updateMember.mutateAsync({
        userId,
        body: { role: nextRole },
      });
    } catch (err) {
      if (err instanceof ApiError) onError(err.message);
    }
  }

  async function handleToggleStatus(userId: string, currentStatus: string) {
    onError(null);
    const nextStatus = currentStatus === "active" ? "inactive" : "active";
    try {
      await updateMember.mutateAsync({
        userId,
        body: { status: nextStatus },
      });
    } catch (err) {
      if (err instanceof ApiError) onError(err.message);
    }
  }

  async function handleRemove(userId: string, memberName: string) {
    if (!window.confirm(t("users.confirmDelete", { name: memberName }))) return;
    onError(null);
    try {
      await removeMember.mutateAsync(userId);
    } catch (err) {
      if (err instanceof ApiError) onError(err.message);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <select
          value={selectedOrgId}
          onChange={(e) => {
            setSelectedOrgId(e.target.value);
            setEditingUserId(null);
          }}
          aria-label={t("users.selectOrg")}
          className="h-11 rounded-lg border border-line bg-surface px-3.5 text-sm font-medium text-fg"
        >
          {manageableOrgs.map((org) => (
            <option key={org.id} value={org.id}>
              {org.name}
            </option>
          ))}
        </select>

        <div className="flex flex-wrap items-center gap-2">
          {canOnlyAppointAdmin && (
            <Button
              variant="outline"
              disabled={addMember.isPending}
              onClick={handleMakeMeAdmin}
            >
              <ShieldCheck aria-hidden className="size-4" />
              {t("users.makeMeOrgAdmin")}
            </Button>
          )}
          {selectedOrgId && (
            <Button onClick={() => setShowCreate((v) => !v)}>
              <Plus aria-hidden className="size-4" />
              {t("users.addUser")}
            </Button>
          )}
        </div>
      </div>

      {showCreate && selectedOrgId && (
        <Card className="p-6">
          <h2 className="mb-4 flex items-center gap-2 text-lg font-bold text-fg">
            <UserPlus aria-hidden className="size-5" />
            {t("users.createTitle")}
          </h2>
          <form onSubmit={handleCreate} className="grid gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("users.emailLabel")}
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
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
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("users.roleLabel")}
              <select
                value={canOnlyAppointAdmin ? "admin" : role}
                disabled={canOnlyAppointAdmin}
                onChange={(e) => setRole(e.target.value as "member" | "admin")}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg disabled:opacity-60"
              >
                {!canOnlyAppointAdmin && (
                  <option value="member">{t("users.roleMember")}</option>
                )}
                <option value="admin">{t("users.roleOrgAdmin")}</option>
              </select>
            </label>
            {!canOnlyAppointAdmin && (
              <>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  {t("users.positionLabel")}
                  <input
                    type="text"
                    list="user-org-position-options"
                    value={position}
                    onChange={(e) => setPosition(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                  />
                  <datalist id="user-org-position-options">
                    {positionLookups.map((item) => (
                      <option key={item.id} value={item.label} />
                    ))}
                  </datalist>
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  {t("users.periodLabel")}
                  <select
                    value={periodId}
                    onChange={(e) => setPeriodId(e.target.value)}
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
              </>
            )}
            <div className="flex gap-2 sm:col-span-2">
              <Button type="submit" disabled={addMember.isPending}>
                {t("users.saveOrgMember")}
              </Button>
              <Button variant="outline" onClick={() => setShowCreate(false)}>
                {t("users.cancel")}
              </Button>
            </div>
          </form>
        </Card>
      )}

      {selectedOrgId && (
        <>
          <div className="flex flex-wrap gap-3">
            <div className="relative min-w-56 flex-1">
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
            <select
              value={roleFilter}
              onChange={(e) =>
                setRoleFilter(e.target.value as "all" | "admin" | "member")
              }
              aria-label={t("users.roleAll")}
              className="h-11 rounded-lg border border-line bg-surface px-3.5 text-sm font-medium text-fg"
            >
              <option value="all">{t("users.roleAll")}</option>
              <option value="admin">{t("users.roleOrgAdmin")}</option>
              <option value="member">{t("users.roleMember")}</option>
            </select>
            <select
              value={statusFilter}
              onChange={(e) =>
                setStatusFilter(e.target.value as "all" | "active" | "inactive")
              }
              aria-label={t("users.statusAll")}
              className="h-11 rounded-lg border border-line bg-surface px-3.5 text-sm font-medium text-fg"
            >
              <option value="all">{t("users.statusAll")}</option>
              <option value="active">{t("users.statusActive")}</option>
              <option value="inactive">{t("users.statusInactive")}</option>
            </select>
            <select
              value={periodFilter}
              onChange={(e) => setPeriodFilter(e.target.value)}
              aria-label={t("users.allPeriods")}
              className="h-11 rounded-lg border border-line bg-surface px-3.5 text-sm font-medium text-fg"
            >
              <option value="">{t("users.allPeriods")}</option>
              {periods.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </div>

          <Card className="divide-y divide-line">
            {membersQuery.isLoading ? (
              <p className="p-6 text-sm text-muted">{t("auth.loading")}</p>
            ) : members.length === 0 ? (
              <p className="p-6 text-sm text-muted">{t("users.empty")}</p>
            ) : (
              members.map((m) => {
                const isEditing = editingUserId === m.userId;
                const periodName = periods.find(
                  (p) => p.id === m.periodId,
                )?.name;
                return (
                  <div
                    key={m.userId}
                    className="flex flex-col gap-3 p-4 sm:px-6"
                  >
                    <div className="flex flex-wrap items-center justify-between gap-4">
                      <div>
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="text-sm font-bold text-fg">
                            {m.name}
                          </span>
                          <Badge
                            tone={m.role === "admin" ? "warning" : "neutral"}
                          >
                            {m.role === "admin"
                              ? t("users.roleOrgAdmin")
                              : t("users.roleMember")}
                          </Badge>
                          <Badge
                            tone={m.status === "active" ? "success" : "danger"}
                          >
                            {m.status === "active"
                              ? t("users.statusActive")
                              : t("users.statusInactive")}
                          </Badge>
                        </div>
                        <p className="mt-0.5 text-xs text-muted">
                          {m.email}
                          {m.position ? ` · ${m.position}` : ""}
                          {periodName ? ` · ${periodName}` : ""}
                        </p>
                      </div>

                      {isSelectedOrgAdmin && (
                        <div className="flex flex-wrap items-center gap-2">
                          <Button
                            variant="outline"
                            onClick={() => {
                              setEditingUserId(
                                isEditing ? null : (m.userId ?? null),
                              );
                              setEditPosition(m.position ?? "");
                              setEditPeriodId(m.periodId ?? "");
                            }}
                          >
                            <Pencil aria-hidden className="size-3.5" />
                            {t("users.editPosition")}
                          </Button>
                          <Button
                            variant="outline"
                            disabled={updateMember.isPending}
                            onClick={() =>
                              m.userId &&
                              handleToggleRole(m.userId, m.role ?? "member")
                            }
                          >
                            {m.role === "admin"
                              ? t("users.demoteOrgAdmin")
                              : t("users.promoteOrgAdmin")}
                          </Button>
                          <Button
                            variant="ghost"
                            disabled={updateMember.isPending}
                            onClick={() =>
                              m.userId &&
                              handleToggleStatus(m.userId, m.status ?? "active")
                            }
                          >
                            {m.status === "active"
                              ? t("users.disableUser")
                              : t("users.enableUser")}
                          </Button>
                          <Button
                            variant="danger"
                            disabled={removeMember.isPending}
                            onClick={() =>
                              m.userId && handleRemove(m.userId, m.name ?? "")
                            }
                            aria-label={`${t("users.deleteUser")} ${m.name ?? ""}`}
                          >
                            <Trash2 aria-hidden className="size-4" />
                          </Button>
                        </div>
                      )}
                    </div>

                    {isEditing && m.userId && (
                      <div className="flex flex-wrap items-end gap-2 pt-1">
                        <input
                          type="text"
                          list="user-org-position-options"
                          value={editPosition}
                          onChange={(e) => setEditPosition(e.target.value)}
                          placeholder={t("users.positionLabel")}
                          aria-label={t("users.positionLabel")}
                          className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                        />
                        <select
                          value={editPeriodId}
                          onChange={(e) => setEditPeriodId(e.target.value)}
                          aria-label={t("users.periodLabel")}
                          className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                        >
                          <option value="">{t("users.noPeriod")}</option>
                          {periods.map((p) => (
                            <option key={p.id} value={p.id}>
                              {p.name}
                            </option>
                          ))}
                        </select>
                        <Button
                          disabled={updateMember.isPending}
                          onClick={() =>
                            m.userId && handleSavePosition(m.userId)
                          }
                        >
                          {t("users.save")}
                        </Button>
                        <Button
                          variant="outline"
                          onClick={() => setEditingUserId(null)}
                        >
                          {t("users.cancel")}
                        </Button>
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </Card>

          {membersQuery.hasNextPage && (
            <div className="flex justify-center">
              <Button
                variant="outline"
                disabled={membersQuery.isFetchingNextPage}
                onClick={() => membersQuery.fetchNextPage()}
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
