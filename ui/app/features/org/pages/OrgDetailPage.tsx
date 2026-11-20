import {
  ArrowLeft,
  Check,
  Copy,
  Link2,
  LogOut,
  Pencil,
  Search,
  ShieldAlert,
  Trash2,
  UserPlus,
} from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  useAddOrgMember,
  useCreateOrgInvite,
  useDeleteOrg,
  useQuickSelfAssignOrgAdmin,
  useRemoveOrgMember,
  useRevokeOrgInvite,
  useUpdateOrg,
  useUpdateOrgMember,
} from "../api/actions";
import { resolveOrgError } from "../api/errors";
import {
  useOrg,
  useOrgInvites,
  useOrgMembers,
  useOrgPeriods,
  useOrgPositionLookups,
} from "../api/queries";
import {
  buildInviteJoinUrl,
  computeInviteStatus,
  DEFAULT_POSITION_SUGGESTIONS,
  INVITE_TTL_OPTIONS,
  inviteStatusMeta,
  isParishAdmin,
  kindLabel,
} from "../model";
import type { Route } from "./+types/OrgDetailPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("org.title")} · Domus` }];
}

export default function OrgDetailPage() {
  const { t } = useTranslation();
  const { orgId = "" } = useParams();
  const navigate = useNavigate();
  const { data: me } = useMe();
  const parishAdmin = isParishAdmin(me?.roles);

  const orgQuery = useOrg(orgId);
  const org = orgQuery.data;
  const isOrgAdmin = org?.myRole === "admin";
  const isOrgMember = org?.myRole === "member" || isOrgAdmin;

  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState("");
  const [editDesc, setEditDesc] = useState("");
  const [editActive, setEditActive] = useState(true);
  const [pageError, setPageError] = useState<string | null>(null);

  const [appointEmail, setAppointEmail] = useState("");
  const [appointName, setAppointName] = useState("");

  const [memberQ, setMemberQ] = useState("");
  const [memberRoleFilter, setMemberRoleFilter] = useState("");
  const [memberStatusFilter, setMemberStatusFilter] = useState("");

  const [addEmail, setAddEmail] = useState("");
  const [addName, setAddName] = useState("");
  const [addPhone, setAddPhone] = useState("");
  const [addRole, setAddRole] = useState<"member" | "admin">("member");
  const [addPosition, setAddPosition] = useState("");
  const [addPeriodId, setAddPeriodId] = useState("");

  const [invRole, setInvRole] = useState<"member" | "admin">("member");
  const [invPosition, setInvPosition] = useState("");
  const [invTTL, setInvTTL] = useState(86400);
  const [invMaxUses, setInvMaxUses] = useState("");
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const membersQuery = useOrgMembers(
    orgId,
    {
      q: memberQ.trim() || undefined,
      filter: {
        role: (memberRoleFilter as "admin" | "member") || undefined,
        status: (memberStatusFilter as "active" | "inactive") || undefined,
      },
    },
    Boolean(org) && isOrgMember,
  );
  const invitesQuery = useOrgInvites(orgId, {}, Boolean(org) && isOrgAdmin);
  const lookupsQuery = useOrgPositionLookups(isOrgAdmin);
  const periodsQuery = useOrgPeriods(isOrgAdmin);

  const updateOrg = useUpdateOrg(orgId);
  const deleteOrg = useDeleteOrg();
  const selfAssignAdmin = useQuickSelfAssignOrgAdmin();
  const addMember = useAddOrgMember(orgId);
  const updateMember = useUpdateOrgMember(orgId);
  const removeMember = useRemoveOrgMember(orgId);
  const createInvite = useCreateOrgInvite(orgId);
  const revokeInvite = useRevokeOrgInvite(orgId);

  const positionSuggestions = Array.from(
    new Set([
      ...(lookupsQuery.data ?? []).map((l) => l.label),
      ...DEFAULT_POSITION_SUGGESTIONS,
    ]),
  );

  function startEdit() {
    if (!org) return;
    setEditName(org.name ?? "");
    setEditDesc(org.description ?? "");
    setEditActive(org.active ?? true);
    setEditing(true);
  }

  async function handleSaveOrg(e: FormEvent) {
    e.preventDefault();
    setPageError(null);
    try {
      await updateOrg.mutateAsync({
        name: editName.trim(),
        description: editDesc.trim(),
        active: editActive,
      });
      setEditing(false);
    } catch (err) {
      setPageError(resolveOrgError(err));
    }
  }

  async function handleDeleteOrg() {
    setPageError(null);
    try {
      await deleteOrg.mutateAsync(orgId);
      navigate("/org", { replace: true });
    } catch (err) {
      setPageError(resolveOrgError(err));
    }
  }

  async function handleSelfAssignAdmin() {
    if (!me?.id) return;
    setPageError(null);
    try {
      await selfAssignAdmin.mutateAsync({ orgId, userId: me.id });
    } catch (err) {
      setPageError(resolveOrgError(err));
    }
  }

  async function handleAppointAdmin(e: FormEvent) {
    e.preventDefault();
    setPageError(null);
    try {
      await addMember.mutateAsync({
        email: appointEmail.trim(),
        name: appointName.trim(),
        role: "admin",
      });
      setAppointEmail("");
      setAppointName("");
    } catch (err) {
      setPageError(resolveOrgError(err));
    }
  }

  async function handleAddMember(e: FormEvent) {
    e.preventDefault();
    setPageError(null);
    try {
      await addMember.mutateAsync({
        email: addEmail.trim(),
        name: addName.trim() || undefined,
        phone: addPhone.trim() || undefined,
        role: addRole,
        position: addPosition.trim() || undefined,
        periodId: addPeriodId || undefined,
      });
      setAddEmail("");
      setAddName("");
      setAddPhone("");
      setAddPosition("");
      setAddPeriodId("");
    } catch (err) {
      setPageError(resolveOrgError(err));
    }
  }

  async function handleToggleRole(userId: string, currentRole: string) {
    setPageError(null);
    try {
      await updateMember.mutateAsync({
        userId,
        body: { role: currentRole === "admin" ? "member" : "admin" },
      });
    } catch (err) {
      setPageError(resolveOrgError(err));
    }
  }

  async function handleToggleStatus(userId: string, currentStatus: string) {
    setPageError(null);
    try {
      await updateMember.mutateAsync({
        userId,
        body: { status: currentStatus === "active" ? "inactive" : "active" },
      });
    } catch (err) {
      setPageError(resolveOrgError(err));
    }
  }

  async function handleRemoveMember(userId: string) {
    setPageError(null);
    try {
      await removeMember.mutateAsync(userId);
    } catch (err) {
      setPageError(resolveOrgError(err));
    }
  }

  async function handleCreateInvite(e: FormEvent) {
    e.preventDefault();
    setPageError(null);
    try {
      const parsedMax = invMaxUses.trim()
        ? Number.parseInt(invMaxUses.trim(), 10)
        : undefined;
      await createInvite.mutateAsync({
        role: invRole,
        defaultPosition: invPosition.trim() || undefined,
        ttlSeconds: invTTL,
        maxUses: parsedMax && parsedMax > 0 ? parsedMax : undefined,
      });
      setInvPosition("");
      setInvMaxUses("");
    } catch (err) {
      setPageError(resolveOrgError(err));
    }
  }

  async function handleCopyInvite(inviteId: string, code: string) {
    const url = buildInviteJoinUrl(window.location.origin, code);
    await navigator.clipboard.writeText(url);
    setCopiedId(inviteId);
    setTimeout(() => setCopiedId(null), 2000);
  }

  if (orgQuery.isLoading) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-8 text-sm text-muted">
        Memuat detail organisasi…
      </div>
    );
  }

  if (!org) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8">
        <Card className="p-6 text-center">
          <p className="text-sm text-muted">Organisasi tidak ditemukan.</p>
          <Link to="/org" className="mt-3 inline-block text-sm text-primary">
            ← {t("org.backToDirectory")}
          </Link>
        </Card>
      </div>
    );
  }

  const members = membersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const invites = invitesQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 px-4 py-6">
      <div>
        <Link
          to="/org"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:underline"
        >
          <ArrowLeft aria-hidden className="size-4" />
          {t("org.backToDirectory")}
        </Link>
      </div>

      {pageError && (
        <div
          role="alert"
          className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
        >
          {pageError}
        </div>
      )}

      <Card className="p-5">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="flex flex-col gap-1.5">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="text-2xl font-bold text-fg">{org.name}</h1>
              <Badge tone={org.active ? "success" : "neutral"}>
                {org.active ? "Aktif" : "Nonaktif"}
              </Badge>
              <Badge tone="neutral">{kindLabel(org.kind ?? "")}</Badge>
              {org.myRole === "admin" && (
                <Badge tone="warning">{t("org.orgAdminBadge")}</Badge>
              )}
              {org.myRole === "member" && (
                <Badge tone="success">{t("org.orgMemberBadge")}</Badge>
              )}
            </div>
            <p className="font-mono text-xs text-muted">{org.code}</p>
            {org.description && (
              <p className="mt-1 text-sm text-muted">{org.description}</p>
            )}
          </div>

          <div className="flex flex-wrap gap-2">
            {(isOrgAdmin || parishAdmin) && (
              <Button variant="outline" onClick={startEdit}>
                <Pencil aria-hidden className="size-4" />
                {t("org.editOrg")}
              </Button>
            )}
            {parishAdmin && (
              <Button
                variant="danger"
                disabled={deleteOrg.isPending}
                onClick={handleDeleteOrg}
              >
                <Trash2 aria-hidden className="size-4" />
                {t("org.deleteOrg")}
              </Button>
            )}
          </div>
        </div>

        {editing && (
          <form
            onSubmit={handleSaveOrg}
            className="mt-5 grid gap-4 border-t border-line pt-4 sm:grid-cols-2"
          >
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Nama Organisasi
              <input
                type="text"
                required
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Deskripsi
              <input
                type="text"
                value={editDesc}
                onChange={(e) => setEditDesc(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex items-center gap-2 text-sm font-medium text-fg">
              <input
                type="checkbox"
                checked={editActive}
                onChange={(e) => setEditActive(e.target.checked)}
                className="size-4 accent-primary"
              />
              Organisasi Aktif
            </label>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setEditing(false)}>
                Batal
              </Button>
              <Button type="submit" disabled={updateOrg.isPending}>
                Simpan Perubahan
              </Button>
            </div>
          </form>
        )}
      </Card>

      {parishAdmin && !isOrgAdmin && (
        <Card className="border-primary/40 bg-primary/5 p-5">
          <div className="flex flex-col gap-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-2 text-sm font-medium text-fg">
                <ShieldAlert aria-hidden className="size-5 text-primary" />
                {t("org.parishAdminBanner")}
              </div>
              <Button
                disabled={selfAssignAdmin.isPending}
                onClick={handleSelfAssignAdmin}
              >
                {t("org.joinAsOrgAdmin")}
              </Button>
            </div>

            <form
              onSubmit={handleAppointAdmin}
              className="flex flex-wrap items-end gap-3 border-t border-line pt-3"
            >
              <label className="flex flex-1 flex-col gap-1 text-xs font-medium text-fg">
                Email Admin
                <input
                  type="email"
                  required
                  value={appointEmail}
                  onChange={(e) => setAppointEmail(e.target.value)}
                  placeholder="pengurus@paroki.org"
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                />
              </label>
              <label className="flex flex-1 flex-col gap-1 text-xs font-medium text-fg">
                Nama Admin
                <input
                  type="text"
                  required
                  value={appointName}
                  onChange={(e) => setAppointName(e.target.value)}
                  placeholder="Nama lengkap"
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                />
              </label>
              <Button
                type="submit"
                variant="outline"
                disabled={addMember.isPending}
              >
                {t("org.appointOrgAdmin")}
              </Button>
            </form>
          </div>
        </Card>
      )}

      {!isOrgMember && !parishAdmin && (
        <Card className="p-6 text-center">
          <h2 className="text-base font-semibold text-fg">
            {t("org.rosterRestrictedTitle")}
          </h2>
          <p className="mt-1 text-sm text-muted">
            {t("org.rosterRestrictedDescription")}
          </p>
        </Card>
      )}

      {isOrgMember && (
        <Card className="flex flex-col gap-4 p-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h2 className="text-lg font-semibold text-fg">
              {t("org.rosterTitle")}
            </h2>
            {me?.id && (
              <Button
                variant="outline"
                disabled={removeMember.isPending}
                onClick={() => handleRemoveMember(me.id ?? "")}
              >
                <LogOut aria-hidden className="size-4" />
                {t("org.leaveOrg")}
              </Button>
            )}
          </div>

          <datalist id="org-position-suggestions">
            {positionSuggestions.map((pos) => (
              <option key={pos} value={pos} />
            ))}
          </datalist>

          {isOrgAdmin && (
            <form
              onSubmit={handleAddMember}
              className="grid gap-3 rounded-lg border border-line bg-canvas p-4 sm:grid-cols-3"
            >
              <input
                type="email"
                required
                value={addEmail}
                onChange={(e) => setAddEmail(e.target.value)}
                placeholder="Email anggota *"
                className="h-10 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
              />
              <input
                type="text"
                value={addName}
                onChange={(e) => setAddName(e.target.value)}
                placeholder="Nama lengkap (opsional)"
                className="h-10 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
              />
              <input
                type="tel"
                value={addPhone}
                onChange={(e) => setAddPhone(e.target.value)}
                placeholder="No. HP (opsional)"
                className="h-10 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
              />
              <select
                value={addRole}
                onChange={(e) =>
                  setAddRole(e.target.value as "member" | "admin")
                }
                className="h-10 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
              >
                <option value="member">Anggota (member)</option>
                <option value="admin">Admin Organisasi (admin)</option>
              </select>
              <input
                type="text"
                list="org-position-suggestions"
                value={addPosition}
                onChange={(e) => setAddPosition(e.target.value)}
                placeholder="Jabatan (mis. Sekretaris)"
                className="h-10 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
              />
              <select
                value={addPeriodId}
                onChange={(e) => setAddPeriodId(e.target.value)}
                className="h-10 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
              >
                <option value="">Periode (opsional)</option>
                {(periodsQuery.data ?? []).map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
              <div className="sm:col-span-3 flex justify-end">
                <Button type="submit" disabled={addMember.isPending}>
                  <UserPlus aria-hidden className="size-4" />
                  Tambah Anggota
                </Button>
              </div>
            </form>
          )}

          <div className="flex flex-wrap gap-2">
            <div className="relative min-w-[200px] flex-1">
              <Search
                aria-hidden
                className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted"
              />
              <input
                type="search"
                value={memberQ}
                onChange={(e) => setMemberQ(e.target.value)}
                placeholder="Cari nama, email, atau jabatan…"
                className="h-10 w-full rounded-lg border border-line bg-canvas pl-9 pr-3 text-sm text-fg"
              />
            </div>
            <select
              value={memberRoleFilter}
              onChange={(e) => setMemberRoleFilter(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            >
              <option value="">Semua Peran</option>
              <option value="admin">Admin</option>
              <option value="member">Anggota</option>
            </select>
            <select
              value={memberStatusFilter}
              onChange={(e) => setMemberStatusFilter(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            >
              <option value="">Semua Status</option>
              <option value="active">Aktif</option>
              <option value="inactive">Nonaktif</option>
            </select>
          </div>

          <div className="divide-y divide-line">
            {members.map((m) => (
              <div
                key={m.id}
                className="flex flex-wrap items-center justify-between gap-3 py-3"
              >
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-medium text-fg">{m.name}</span>
                    <Badge tone={m.role === "admin" ? "warning" : "neutral"}>
                      {m.role === "admin"
                        ? t("org.orgAdminBadge")
                        : t("org.orgMemberBadge")}
                    </Badge>
                    <Badge tone={m.status === "active" ? "success" : "danger"}>
                      {m.status === "active" ? "Aktif" : "Nonaktif"}
                    </Badge>
                  </div>
                  <p className="text-xs text-muted">
                    {m.email}
                    {m.position ? ` · ${m.position}` : ""}
                  </p>
                </div>

                {isOrgAdmin && m.userId && (
                  <div className="flex flex-wrap gap-2">
                    <Button
                      variant="outline"
                      onClick={() =>
                        handleToggleRole(m.userId ?? "", m.role ?? "member")
                      }
                    >
                      {m.role === "admin" ? "Jadikan Anggota" : "Jadikan Admin"}
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() =>
                        handleToggleStatus(m.userId ?? "", m.status ?? "active")
                      }
                    >
                      {m.status === "active" ? "Nonaktifkan" : "Aktifkan"}
                    </Button>
                    <Button
                      variant="danger"
                      onClick={() => handleRemoveMember(m.userId ?? "")}
                    >
                      Hapus
                    </Button>
                  </div>
                )}
              </div>
            ))}
          </div>
        </Card>
      )}

      {isOrgAdmin && (
        <Card className="flex flex-col gap-4 p-5">
          <div className="flex items-center gap-2">
            <Link2 aria-hidden className="size-5 text-primary" />
            <h2 className="text-lg font-semibold text-fg">
              {t("org.invitesTitle")}
            </h2>
          </div>

          <form
            onSubmit={handleCreateInvite}
            className="grid gap-3 rounded-lg border border-line bg-canvas p-4 sm:grid-cols-4"
          >
            <select
              value={invRole}
              onChange={(e) => setInvRole(e.target.value as "member" | "admin")}
              className="h-10 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
            >
              <option value="member">Anggota (member)</option>
              <option value="admin">Admin Organisasi (admin)</option>
            </select>
            <input
              type="text"
              list="org-position-suggestions"
              value={invPosition}
              onChange={(e) => setInvPosition(e.target.value)}
              placeholder="Jabatan default (opsional)"
              className="h-10 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
            />
            <select
              value={invTTL}
              onChange={(e) => setInvTTL(Number(e.target.value))}
              className="h-10 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
            >
              {INVITE_TTL_OPTIONS.map((opt) => (
                <option key={opt.seconds} value={opt.seconds}>
                  Berlaku {opt.label}
                </option>
              ))}
            </select>
            <input
              type="number"
              min={1}
              value={invMaxUses}
              onChange={(e) => setInvMaxUses(e.target.value)}
              placeholder="Maks. pakai (opsional)"
              className="h-10 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
            />
            <div className="sm:col-span-4 flex justify-end">
              <Button type="submit" disabled={createInvite.isPending}>
                {t("org.createInvite")}
              </Button>
            </div>
          </form>

          <div className="divide-y divide-line">
            {invites.map((inv) => {
              const invId = inv.id ?? "";
              const invCode = inv.code ?? "";
              const status = computeInviteStatus({
                revokedAt: inv.revokedAt,
                expiresAt: inv.expiresAt ?? "",
                maxUses: inv.maxUses,
                usedCount: inv.usedCount ?? 0,
              });
              const meta = inviteStatusMeta(status);
              const shareUrl =
                typeof window !== "undefined"
                  ? buildInviteJoinUrl(window.location.origin, invCode)
                  : `/org/join/${invCode}`;
              return (
                <div
                  key={invId}
                  className="flex flex-wrap items-center justify-between gap-3 py-3"
                >
                  <div className="flex flex-col gap-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-mono text-sm font-semibold text-fg">
                        {invCode}
                      </span>
                      <Badge tone={meta.tone}>{meta.label}</Badge>
                      <Badge tone="neutral">{inv.role}</Badge>
                      {inv.defaultPosition && (
                        <span className="text-xs text-muted">
                          · {inv.defaultPosition}
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-muted">
                      {shareUrl} · Terpakai: {inv.usedCount ?? 0}
                      {inv.maxUses != null ? ` / ${inv.maxUses}` : ""} ·
                      Kedaluwarsa:{" "}
                      {inv.expiresAt
                        ? new Date(inv.expiresAt).toLocaleString("id-ID")
                        : "-"}
                    </p>
                  </div>

                  <div className="flex flex-wrap gap-2">
                    <Button
                      variant="outline"
                      onClick={() => handleCopyInvite(invId, invCode)}
                    >
                      {copiedId === invId ? (
                        <>
                          <Check aria-hidden className="size-4" />
                          {t("org.copied")}
                        </>
                      ) : (
                        <>
                          <Copy aria-hidden className="size-4" />
                          {t("org.copyLink")}
                        </>
                      )}
                    </Button>
                    {status === "active" && (
                      <Button
                        variant="danger"
                        onClick={() => revokeInvite.mutateAsync(invId)}
                      >
                        {t("org.revokeInvite")}
                      </Button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </Card>
      )}
    </div>
  );
}
