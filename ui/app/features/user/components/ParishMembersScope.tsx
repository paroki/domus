import { Plus, Search, Trash2, UserPlus } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  useDirectoryAddParishMember,
  useDirectoryRemoveParishMember,
  useDirectoryUpdateParishMember,
} from "../api/actions";
import {
  useDirectoryParishes,
  useDirectoryParishMembers,
} from "../api/queries";
import type { User } from "../api/types";
import { isSuperAdmin } from "../model";

export function ParishMembersScope({
  me,
  onError,
}: {
  me: User;
  onError: (msg: string | null) => void;
}) {
  const { t } = useTranslation();
  const superAdmin = isSuperAdmin(me);
  const parishesQuery = useDirectoryParishes(superAdmin);
  const parishes = parishesQuery.data ?? [];

  const [selectedParishId, setSelectedParishId] = useState<string>(
    me.activeParishId ?? "",
  );

  useEffect(() => {
    if (!superAdmin) {
      setSelectedParishId(me.activeParishId ?? "");
      return;
    }
    if (selectedParishId) return;
    if (me.activeParishId) {
      setSelectedParishId(me.activeParishId);
    } else if (parishes[0]?.id) {
      setSelectedParishId(parishes[0].id);
    }
  }, [superAdmin, me.activeParishId, parishes, selectedParishId]);

  const effectiveParishId = superAdmin
    ? selectedParishId
    : (me.activeParishId ?? "");

  const [q, setQ] = useState("");
  const [roleFilter, setRoleFilter] = useState<"all" | "admin" | "member">(
    "all",
  );
  const [showCreate, setShowCreate] = useState(false);
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [role, setRole] = useState<"member" | "admin">("member");

  const membersQuery = useDirectoryParishMembers(
    effectiveParishId,
    {
      q: q.trim() || undefined,
      filter: roleFilter === "all" ? undefined : { role: roleFilter },
    },
    Boolean(effectiveParishId),
  );

  const addMember = useDirectoryAddParishMember(effectiveParishId);
  const updateMember = useDirectoryUpdateParishMember(effectiveParishId);
  const removeMember = useDirectoryRemoveParishMember(effectiveParishId);

  if (!effectiveParishId && !superAdmin) {
    return (
      <Card className="p-6 text-sm text-muted">
        {t("users.noActiveParish")}
      </Card>
    );
  }

  const members = membersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    if (!effectiveParishId) return;
    onError(null);
    try {
      await addMember.mutateAsync({
        email: email.trim(),
        name: name.trim() || undefined,
        phone: phone.trim() || undefined,
        role,
      });
      setEmail("");
      setName("");
      setPhone("");
      setRole("member");
      setShowCreate(false);
    } catch (err) {
      if (err instanceof ApiError) {
        onError(err.details[0]?.message ?? err.message);
      }
    }
  }

  async function handleRoleChange(
    userId: string,
    nextRole: "admin" | "member",
  ) {
    onError(null);
    try {
      await updateMember.mutateAsync({
        userId,
        body: { role: nextRole },
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
        {superAdmin ? (
          <select
            value={effectiveParishId}
            onChange={(e) => setSelectedParishId(e.target.value)}
            aria-label={t("users.selectParish")}
            className="h-11 rounded-lg border border-line bg-surface px-3.5 text-sm font-medium text-fg"
          >
            {parishes.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        ) : (
          <div />
        )}

        {effectiveParishId && (
          <Button onClick={() => setShowCreate((v) => !v)}>
            <Plus aria-hidden className="size-4" />
            {t("users.addUser")}
          </Button>
        )}
      </div>

      {showCreate && effectiveParishId && (
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
                value={role}
                onChange={(e) => setRole(e.target.value as "member" | "admin")}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="member">{t("users.roleMember")}</option>
                <option value="admin">{t("users.roleParishAdmin")}</option>
              </select>
            </label>
            <div className="flex gap-2 sm:col-span-2">
              <Button type="submit" disabled={addMember.isPending}>
                {t("users.saveParishMember")}
              </Button>
              <Button variant="outline" onClick={() => setShowCreate(false)}>
                {t("users.cancel")}
              </Button>
            </div>
          </form>
        </Card>
      )}

      {effectiveParishId && (
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
              <option value="admin">{t("users.roleParishAdmin")}</option>
              <option value="member">{t("users.roleMember")}</option>
            </select>
          </div>

          <Card className="divide-y divide-line">
            {membersQuery.isLoading ? (
              <p className="p-6 text-sm text-muted">{t("auth.loading")}</p>
            ) : members.length === 0 ? (
              <p className="p-6 text-sm text-muted">{t("users.empty")}</p>
            ) : (
              members.map((m) => (
                <div
                  key={m.userId}
                  className="flex flex-wrap items-center justify-between gap-4 p-4 sm:px-6"
                >
                  <div>
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="text-sm font-bold text-fg">
                        {m.name}
                      </span>
                      <Badge tone={m.role === "admin" ? "warning" : "neutral"}>
                        {m.role === "admin"
                          ? t("users.roleParishAdmin")
                          : t("users.roleMember")}
                      </Badge>
                    </div>
                    <p className="mt-0.5 text-xs text-muted">{m.email}</p>
                  </div>

                  <div className="flex flex-wrap items-center gap-2">
                    <select
                      value={m.role ?? "member"}
                      disabled={updateMember.isPending}
                      onChange={(e) =>
                        m.userId &&
                        handleRoleChange(
                          m.userId,
                          e.target.value as "admin" | "member",
                        )
                      }
                      aria-label={`${t("users.roleLabel")} ${m.name ?? ""}`}
                      className="h-9 rounded-lg border border-line bg-surface px-2.5 text-xs font-medium text-fg"
                    >
                      <option value="member">{t("users.roleMember")}</option>
                      <option value="admin">
                        {t("users.roleParishAdmin")}
                      </option>
                    </select>
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
                </div>
              ))
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
