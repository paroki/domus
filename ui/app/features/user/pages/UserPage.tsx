import { Plus, Search, ShieldAlert, Trash2, UserPlus } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  useAdminCreateUser,
  useAdminDeleteUser,
  useAdminUpdateUser,
} from "../api/actions";
import { useMe, useUsersSearch } from "../api/queries";
import { isSuperAdmin } from "../model";
import type { Route } from "./+types/UserPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.user.name")} · Domus` }];
}

export default function UserPage() {
  const { t } = useTranslation();
  const { data: me } = useMe();
  const [q, setQ] = useState("");
  const [statusFilter, setStatusFilter] = useState<
    "all" | "active" | "disabled"
  >("all");
  const [showCreate, setShowCreate] = useState(false);
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [makeSuperAdmin, setMakeSuperAdmin] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const allowed = isSuperAdmin(me);
  const searchQuery = useUsersSearch({
    q: q.trim() || undefined,
    filter: statusFilter === "all" ? undefined : { status: statusFilter },
  });
  const createUser = useAdminCreateUser();
  const updateUser = useAdminUpdateUser();
  const deleteUser = useAdminDeleteUser();

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
      </div>
    );
  }

  const users = searchQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      await createUser.mutateAsync({
        email: email.trim(),
        name: name.trim(),
        phone: phone.trim() || undefined,
        isSuperAdmin: makeSuperAdmin,
      });
      setEmail("");
      setName("");
      setPhone("");
      setMakeSuperAdmin(false);
      setShowCreate(false);
    } catch (err) {
      if (err instanceof ApiError) {
        setErrorMsg(err.details[0]?.message ?? err.message);
      }
    }
  }

  async function handleToggleSuperAdmin(id: string, current: boolean) {
    setErrorMsg(null);
    try {
      await updateUser.mutateAsync({
        id,
        body: { isSuperAdmin: !current },
      });
    } catch (err) {
      if (err instanceof ApiError) setErrorMsg(err.message);
    }
  }

  async function handleToggleStatus(id: string, currentStatus: string) {
    setErrorMsg(null);
    const nextStatus = currentStatus === "active" ? "disabled" : "active";
    try {
      await updateUser.mutateAsync({ id, body: { status: nextStatus } });
    } catch (err) {
      if (err instanceof ApiError) setErrorMsg(err.message);
    }
  }

  async function handleDelete(id: string, userName: string) {
    if (!window.confirm(t("users.confirmDelete", { name: userName }))) return;
    setErrorMsg(null);
    try {
      await deleteUser.mutateAsync(id);
    } catch (err) {
      if (err instanceof ApiError) setErrorMsg(err.message);
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-8 sm:px-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg">{t("users.title")}</h1>
          <p className="text-sm text-muted">{t("users.subtitle")}</p>
        </div>
        <Button onClick={() => setShowCreate((v) => !v)}>
          <Plus aria-hidden className="size-4" />
          {t("users.addUser")}
        </Button>
      </div>

      {errorMsg && (
        <div
          role="alert"
          className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
        >
          {errorMsg}
        </div>
      )}

      {showCreate && (
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
            <label className="flex items-center gap-2.5 self-end pb-2.5 text-sm font-medium text-fg">
              <input
                type="checkbox"
                checked={makeSuperAdmin}
                onChange={(e) => setMakeSuperAdmin(e.target.checked)}
                className="size-4 rounded border-line"
              />
              {t("users.superAdminLabel")}
            </label>
            <div className="flex gap-2 sm:col-span-2">
              <Button type="submit" disabled={createUser.isPending}>
                {t("users.createSubmit")}
              </Button>
              <Button variant="outline" onClick={() => setShowCreate(false)}>
                {t("users.cancel")}
              </Button>
            </div>
          </form>
        </Card>
      )}

      <div className="flex flex-wrap gap-3">
        <div className="relative flex-1 min-w-56">
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
          value={statusFilter}
          onChange={(e) =>
            setStatusFilter(e.target.value as "all" | "active" | "disabled")
          }
          aria-label={t("users.statusAll")}
          className="h-11 rounded-lg border border-line bg-surface px-3.5 text-sm font-medium text-fg"
        >
          <option value="all">{t("users.statusAll")}</option>
          <option value="active">{t("users.statusActive")}</option>
          <option value="disabled">{t("users.statusDisabled")}</option>
        </select>
      </div>

      <Card className="divide-y divide-line">
        {searchQuery.isLoading ? (
          <p className="p-6 text-sm text-muted">{t("auth.loading")}</p>
        ) : users.length === 0 ? (
          <p className="p-6 text-sm text-muted">{t("users.empty")}</p>
        ) : (
          users.map((u) => {
            const superAdmin = isSuperAdmin(u);
            return (
              <div
                key={u.id}
                className="flex flex-wrap items-center justify-between gap-4 p-4 sm:px-6"
              >
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-bold text-fg">{u.name}</span>
                    <Badge tone={u.status === "active" ? "success" : "danger"}>
                      {u.status === "active"
                        ? t("users.statusActive")
                        : t("users.statusDisabled")}
                    </Badge>
                    {superAdmin && (
                      <Badge tone="warning">{t("users.roleSuperAdmin")}</Badge>
                    )}
                  </div>
                  <p className="mt-0.5 text-xs text-muted">
                    {u.email}
                    {u.phone ? ` · ${u.phone}` : ""}
                  </p>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  <Button
                    variant="outline"
                    disabled={updateUser.isPending}
                    onClick={() =>
                      u.id && handleToggleSuperAdmin(u.id, superAdmin)
                    }
                  >
                    {superAdmin
                      ? t("users.demoteSuperAdmin")
                      : t("users.promoteSuperAdmin")}
                  </Button>
                  <Button
                    variant="ghost"
                    disabled={updateUser.isPending}
                    onClick={() =>
                      u.id && handleToggleStatus(u.id, u.status ?? "active")
                    }
                  >
                    {u.status === "active"
                      ? t("users.disableUser")
                      : t("users.enableUser")}
                  </Button>
                  <Button
                    variant="danger"
                    disabled={deleteUser.isPending}
                    onClick={() => u.id && handleDelete(u.id, u.name ?? "")}
                    aria-label={`${t("users.deleteUser")} ${u.name ?? ""}`}
                  >
                    <Trash2 aria-hidden className="size-4" />
                  </Button>
                </div>
              </div>
            );
          })
        )}
      </Card>

      {searchQuery.hasNextPage && (
        <div className="flex justify-center">
          <Button
            variant="outline"
            disabled={searchQuery.isFetchingNextPage}
            onClick={() => searchQuery.fetchNextPage()}
          >
            {t("users.loadMore")}
          </Button>
        </div>
      )}
    </div>
  );
}
