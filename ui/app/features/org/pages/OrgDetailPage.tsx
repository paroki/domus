import {
  Building2,
  CalendarRange,
  Check,
  Copy,
  Link2,
  LogOut,
  Network,
  Pencil,
  Plus,
  Search,
  Trash2,
} from "lucide-react";
import {
  type FormEvent,
  type ReactNode,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import {
  InductiveDangerZone,
  InductiveEmptyState,
  InductivePageHeader,
} from "~/shared/ui/iui";
import {
  useDeleteOrg,
  useDeleteOrgJabatan,
  useDeleteOrgPeriode,
  useRemoveOrgMember,
  useRevokeOrgInvite,
  useUpdateOrg,
  useUpdateOrgPeriode,
} from "../api/actions";
import { resolveOrgError } from "../api/errors";
import {
  useOrg,
  useOrgInvites,
  useOrgJabatanList,
  useOrgMembers,
  useOrgPeriodeList,
} from "../api/queries";
import type {
  OrgInviteResponse,
  OrgJabatanResponse,
  OrgPeriodeResponse,
} from "../api/types";
import {
  buildInviteJoinUrl,
  buildJabatanTree,
  computeInviteStatus,
  inviteStatusMeta,
  isParokiAdmin,
  kindLabel,
  type OrgJabatanNode,
} from "../model";
import type { Route } from "./+types/OrgDetailPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("org.title")} · Domus` }];
}

export default function OrgDetailPage() {
  const { t } = useTranslation();
  const { orgId = "" } = useParams();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const hasParoki = Boolean(me?.activeParokiId);
  const parokiAdmin = isParokiAdmin(me?.roles);

  const orgQuery = useOrg(orgId);
  const org = orgQuery.data;
  const isOrgAdmin = org?.myRole === "admin";
  const isOrgMember = org?.myRole === "member" || isOrgAdmin;
  const canReadOrgInternal = isOrgMember || parokiAdmin;
  const canManageOrg = isOrgAdmin || parokiAdmin;

  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState("");
  const [editDesc, setEditDesc] = useState("");
  const [editActive, setEditActive] = useState(true);
  const [pageError, setPageError] = useState<string | null>(null);

  const [memberQ, setMemberQ] = useState("");
  const [memberRoleFilter, setMemberRoleFilter] = useState("");
  const [memberStatusFilter, setMemberStatusFilter] = useState("");
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

  const updateOrg = useUpdateOrg(orgId);
  const deleteOrg = useDeleteOrg();
  const removeMember = useRemoveOrgMember(orgId);
  const revokeInvite = useRevokeOrgInvite(orgId);
  const updatePeriode = useUpdateOrgPeriode(orgId);
  const deletePeriode = useDeleteOrgPeriode(orgId);
  const deleteJabatan = useDeleteOrgJabatan(orgId);

  const periodeQuery = useOrgPeriodeList(
    orgId,
    {},
    Boolean(org) && canReadOrgInternal,
  );
  const periods = periodeQuery.data ?? [];

  const [selectedPeriodeId, setSelectedPeriodeId] = useState("");
  useEffect(() => {
    if (periods.length === 0) {
      setSelectedPeriodeId("");
      return;
    }
    if (!periods.some((p) => p.id === selectedPeriodeId)) {
      const current = periods.find((p) => p.isCurrent) ?? periods[0];
      setSelectedPeriodeId(current?.id ?? "");
    }
  }, [periods, selectedPeriodeId]);

  const jabatanQuery = useOrgJabatanList(
    orgId,
    { filter: { periodeId: selectedPeriodeId || undefined } },
    Boolean(org) && canReadOrgInternal && Boolean(selectedPeriodeId),
  );
  const jabatanTree = useMemo(
    () => buildJabatanTree(jabatanQuery.data ?? []),
    [jabatanQuery.data],
  );

  const [pendingDeletePeriode, setPendingDeletePeriode] =
    useState<OrgPeriodeResponse | null>(null);
  const [pendingDeleteJabatan, setPendingDeleteJabatan] =
    useState<OrgJabatanResponse | null>(null);
  const [pendingRevokeInvite, setPendingRevokeInvite] =
    useState<OrgInviteResponse | null>(null);
  const [pendingLeaveOrg, setPendingLeaveOrg] = useState(false);
  const [pendingDeleteOrg, setPendingDeleteOrg] = useState(false);

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

  async function handleCopyInvite(inviteId: string, code: string) {
    setPageError(null);
    try {
      const url = buildInviteJoinUrl(window.location.origin, code);
      await navigator.clipboard.writeText(url);
      setCopiedId(inviteId);
      setTimeout(() => setCopiedId(null), 2000);
    } catch (err) {
      setPageError(resolveOrgError(err));
    }
  }

  if (!meLoading && !hasParoki) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-6 sm:px-6 sm:py-8">
        <InductiveEmptyState
          icon={Building2}
          title={t("org.noParishTitle")}
          description={t("org.noParishDescription")}
        />
      </div>
    );
  }

  if (meLoading || orgQuery.isLoading) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-6 text-sm text-muted sm:px-6 sm:py-8">
        Memuat detail organisasi…
      </div>
    );
  }

  if (orgQuery.isError || !org) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-6 sm:px-6 sm:py-8">
        <InductiveEmptyState
          icon={Building2}
          title="Organisasi tidak ditemukan"
          description={
            orgQuery.isError
              ? resolveOrgError(orgQuery.error)
              : "Organisasi yang Anda cari tidak tersedia."
          }
          action={
            <Link to="/org" className={buttonClass("outline")}>
              {t("org.backToDirectory")}
            </Link>
          }
        />
      </div>
    );
  }

  const members = membersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const invites = invitesQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const showDangerZone = (isOrgMember && Boolean(me?.id)) || parokiAdmin;

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Organisasi", to: "/org" },
          { label: org.name ?? "Detail Organisasi" },
        ]}
        title={org.name ?? ""}
        subtitle={org.code ? `Kode: ${org.code}` : ""}
      />

      {pageError && (
        <div
          role="alert"
          className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
        >
          {pageError}
        </div>
      )}

      <Card className="p-4 sm:p-6">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line pb-4">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-lg font-semibold text-fg">Profil Organisasi</h2>
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

          {canManageOrg && !editing && (
            <Button
              variant="outline"
              onClick={startEdit}
              className="min-h-10 w-full sm:w-auto"
            >
              <Pencil aria-hidden className="size-4" />
              Ubah
            </Button>
          )}
        </div>

        {!editing ? (
          <dl className="mt-4 grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-xs font-medium text-muted">
                Nama Organisasi
              </dt>
              <dd className="mt-1 text-sm font-medium text-fg">{org.name}</dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">Kode</dt>
              <dd className="mt-1 font-mono text-sm text-fg">{org.code}</dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">
                Jenis Organisasi
              </dt>
              <dd className="mt-1 text-sm text-fg">
                {kindLabel(org.kind ?? "")}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">Status</dt>
              <dd className="mt-1 text-sm text-fg">
                {org.active ? "Aktif" : "Nonaktif"}
              </dd>
            </div>
            <div className="sm:col-span-2">
              <dt className="text-xs font-medium text-muted">Deskripsi</dt>
              <dd className="mt-1 text-sm text-fg">{org.description || "—"}</dd>
            </div>
          </dl>
        ) : (
          <form
            onSubmit={handleSaveOrg}
            className="mt-4 grid gap-4 sm:grid-cols-2"
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
            <label className="flex min-h-11 cursor-pointer items-center gap-2.5 text-sm font-medium text-fg sm:col-span-2">
              <input
                type="checkbox"
                checked={editActive}
                onChange={(e) => setEditActive(e.target.checked)}
                className="size-4 accent-primary"
              />
              Organisasi Aktif
            </label>
            <div className="flex flex-col-reverse gap-2 sm:col-span-2 sm:flex-row sm:justify-end">
              <Button
                variant="outline"
                onClick={() => setEditing(false)}
                className="w-full sm:w-auto"
              >
                Batal
              </Button>
              <Button
                type="submit"
                disabled={updateOrg.isPending}
                className="w-full sm:w-auto"
              >
                Simpan Perubahan
              </Button>
            </div>
          </form>
        )}
      </Card>

      {!canReadOrgInternal && (
        <InductiveEmptyState
          icon={Building2}
          title={t("org.rosterRestrictedTitle")}
          description={t("org.rosterRestrictedDescription")}
        />
      )}

      {canReadOrgInternal && (
        <Card className="flex flex-col gap-4 p-4 sm:p-6">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <CalendarRange aria-hidden className="size-5 text-primary" />
              <h2 className="text-lg font-semibold text-fg">
                Periode Kepengurusan
              </h2>
            </div>
            {canManageOrg && (
              <Link
                to={`/org/${orgId}/periode/new`}
                className={`${buttonClass("primary")} min-h-11 w-full sm:w-auto`}
              >
                <Plus aria-hidden className="size-4" />
                Tambah Periode
              </Link>
            )}
          </div>

          {periodeQuery.isLoading ? (
            <p className="text-sm text-muted">Memuat periode kepengurusan…</p>
          ) : periods.length === 0 ? (
            <p className="text-sm text-muted">
              Belum ada periode kepengurusan yang terdaftar.
            </p>
          ) : (
            <div className="divide-y divide-line">
              {periods.map((p) => (
                <div
                  key={p.id}
                  className="flex flex-wrap items-center justify-between gap-3 py-3"
                >
                  <div className="flex min-w-0 flex-1 flex-col gap-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-semibold text-fg">{p.nama}</span>
                      <span className="font-mono text-xs text-muted">
                        ({p.kode})
                      </span>
                      <Badge tone={p.status === "open" ? "success" : "neutral"}>
                        {p.status === "open" ? "Terbuka" : "Ditutup"}
                      </Badge>
                      {p.isCurrent && (
                        <Badge tone="warning">Periode Aktif</Badge>
                      )}
                    </div>
                    <p className="text-xs text-muted">
                      {p.tanggalMulai} – {p.tanggalSelesai}
                    </p>
                  </div>

                  {canManageOrg && (
                    <div className="flex flex-wrap items-center gap-2">
                      {!p.isCurrent && (
                        <Button
                          variant="outline"
                          className="min-h-10 px-3 text-xs"
                          disabled={updatePeriode.isPending}
                          onClick={async () => {
                            setPageError(null);
                            try {
                              await updatePeriode.mutateAsync({
                                periodeId: p.id ?? "",
                                body: { isCurrent: true },
                              });
                            } catch (err) {
                              setPageError(resolveOrgError(err));
                            }
                          }}
                        >
                          Jadikan Aktif
                        </Button>
                      )}
                      <Button
                        variant="outline"
                        className="min-h-10 px-3 text-xs"
                        disabled={updatePeriode.isPending}
                        onClick={async () => {
                          setPageError(null);
                          try {
                            await updatePeriode.mutateAsync({
                              periodeId: p.id ?? "",
                              body: {
                                status: p.status === "open" ? "closed" : "open",
                              },
                            });
                          } catch (err) {
                            setPageError(resolveOrgError(err));
                          }
                        }}
                      >
                        {p.status === "open" ? "Tutup" : "Buka"}
                      </Button>
                      <Link
                        to={`/org/${orgId}/periode/${p.id}/edit`}
                        className={`${buttonClass("outline")} min-h-10 px-3 text-xs`}
                      >
                        <Pencil aria-hidden className="size-3.5" />
                        Ubah
                      </Link>
                      <Button
                        variant="danger"
                        className="min-h-10 px-3 text-xs"
                        disabled={deletePeriode.isPending}
                        onClick={() => setPendingDeletePeriode(p)}
                      >
                        <Trash2 aria-hidden className="size-3.5" />
                        Hapus
                      </Button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </Card>
      )}

      {canReadOrgInternal && (
        <Card className="flex flex-col gap-4 p-4 sm:p-6">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <Network aria-hidden className="size-5 text-primary" />
              <h2 className="text-lg font-semibold text-fg">
                Struktur &amp; Jabatan Organisasi
              </h2>
            </div>
            <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
              {periods.length > 0 && (
                <select
                  aria-label="Pilih periode struktur jabatan"
                  value={selectedPeriodeId}
                  onChange={(e) => setSelectedPeriodeId(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg sm:w-auto"
                >
                  {periods.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.nama}
                      {p.isCurrent ? " (Aktif)" : ""}
                    </option>
                  ))}
                </select>
              )}
              {canManageOrg && selectedPeriodeId && (
                <Link
                  to={`/org/${orgId}/periode/${selectedPeriodeId}/jabatan/new`}
                  className={`${buttonClass("primary")} min-h-11 w-full sm:w-auto`}
                >
                  <Plus aria-hidden className="size-4" />
                  Tambah Jabatan
                </Link>
              )}
            </div>
          </div>

          {!selectedPeriodeId ? (
            <p className="text-sm text-muted">
              Buat periode kepengurusan terlebih dahulu untuk menyusun struktur
              jabatan.
            </p>
          ) : jabatanQuery.isLoading ? (
            <p className="text-sm text-muted">Memuat struktur jabatan…</p>
          ) : jabatanTree.length === 0 ? (
            <p className="text-sm text-muted">
              Belum ada jabatan pada periode ini.
            </p>
          ) : (
            <div className="divide-y divide-line">
              {renderJabatanNodes(jabatanTree)}
            </div>
          )}
        </Card>
      )}

      {isOrgMember && (
        <Card className="flex flex-col gap-4 p-4 sm:p-6">
          <h2 className="text-lg font-semibold text-fg">
            {t("org.rosterTitle")}
          </h2>

          <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap">
            <div className="relative min-w-0 flex-1">
              <Search
                aria-hidden
                className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted"
              />
              <input
                type="search"
                value={memberQ}
                onChange={(e) => setMemberQ(e.target.value)}
                placeholder="Cari nama atau email…"
                className="h-11 w-full rounded-lg border border-line bg-canvas pl-9 pr-3 text-sm text-fg"
              />
            </div>
            <select
              value={memberRoleFilter}
              onChange={(e) => setMemberRoleFilter(e.target.value)}
              className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg sm:w-auto"
            >
              <option value="">Semua Peran</option>
              <option value="admin">Admin</option>
              <option value="member">Anggota</option>
            </select>
            <select
              value={memberStatusFilter}
              onChange={(e) => setMemberStatusFilter(e.target.value)}
              className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg sm:w-auto"
            >
              <option value="">Semua Status</option>
              <option value="active">Aktif</option>
              <option value="inactive">Nonaktif</option>
            </select>
          </div>

          {membersQuery.isLoading ? (
            <p className="text-sm text-muted">{t("org.loadingMembers")}</p>
          ) : members.length === 0 ? (
            <p className="text-sm text-muted">{t("org.noMembers")}</p>
          ) : (
            <div className="divide-y divide-line">
              {members.map((m) => (
                <div key={m.id} className="flex flex-col gap-3 py-3">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-medium text-fg">{m.name}</span>
                        <Badge
                          tone={m.role === "admin" ? "warning" : "neutral"}
                        >
                          {m.role === "admin"
                            ? t("org.orgAdminBadge")
                            : t("org.orgMemberBadge")}
                        </Badge>
                        <Badge
                          tone={m.status === "active" ? "success" : "danger"}
                        >
                          {m.status === "active" ? "Aktif" : "Nonaktif"}
                        </Badge>
                      </div>
                      <p className="break-all text-xs text-muted">{m.email}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {membersQuery.hasNextPage && (
            <div className="flex justify-center pt-2">
              <Button
                variant="outline"
                disabled={membersQuery.isFetchingNextPage}
                onClick={() => membersQuery.fetchNextPage()}
                className="w-full sm:w-auto"
              >
                {t("org.loadMore")}
              </Button>
            </div>
          )}
        </Card>
      )}

      {isOrgAdmin && (
        <Card id="invites" className="flex flex-col gap-4 p-4 sm:p-6">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <Link2 aria-hidden className="size-5 text-primary" />
              <h2 className="text-lg font-semibold text-fg">
                {t("org.invitesTitle")}
              </h2>
            </div>
            <Link
              to={`/org/${orgId}/invites/new`}
              className={`${buttonClass("primary")} min-h-11 w-full sm:w-auto`}
            >
              <Plus aria-hidden className="size-4" />
              {t("org.createInvite")}
            </Link>
          </div>

          {invitesQuery.isLoading ? (
            <p className="text-sm text-muted">{t("org.loadingInvites")}</p>
          ) : invites.length === 0 ? (
            <p className="text-sm text-muted">{t("org.noInvites")}</p>
          ) : (
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
                    <div className="flex min-w-0 flex-1 flex-col gap-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-mono text-sm font-semibold text-fg">
                          {invCode}
                        </span>
                        <Badge tone={meta.tone}>{meta.label}</Badge>
                        <Badge tone="neutral">{inv.role}</Badge>
                      </div>
                      <p className="min-w-0 break-all text-xs text-muted">
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
                        className="min-h-10"
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
                          className="min-h-10"
                          disabled={revokeInvite.isPending}
                          onClick={() => setPendingRevokeInvite(inv)}
                        >
                          {t("org.revokeInvite")}
                        </Button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {invitesQuery.hasNextPage && (
            <div className="flex justify-center pt-2">
              <Button
                variant="outline"
                disabled={invitesQuery.isFetchingNextPage}
                onClick={() => invitesQuery.fetchNextPage()}
                className="w-full sm:w-auto"
              >
                {t("org.loadMore")}
              </Button>
            </div>
          )}
        </Card>
      )}

      {showDangerZone && (
        <InductiveDangerZone
          title="Zona Berbahaya"
          description="Tindakan di bawah ini mengubah keanggotaan Anda atau menonaktifkan organisasi dari direktori."
          action={
            <div className="flex flex-col gap-2 sm:flex-row">
              {isOrgMember && me?.id && (
                <Button
                  variant="outline"
                  disabled={removeMember.isPending}
                  onClick={() => setPendingLeaveOrg(true)}
                  className="w-full sm:w-auto"
                >
                  <LogOut aria-hidden className="size-4" />
                  {t("org.leaveOrg")}
                </Button>
              )}
              {parokiAdmin && (
                <Button
                  variant="danger"
                  disabled={deleteOrg.isPending}
                  onClick={() => setPendingDeleteOrg(true)}
                  className="w-full sm:w-auto"
                >
                  <Trash2 aria-hidden className="size-4" />
                  {t("org.deleteOrg")}
                </Button>
              )}
            </div>
          }
        />
      )}

      <ConfirmDialog
        open={Boolean(pendingDeletePeriode)}
        title="Hapus periode kepengurusan?"
        description={
          pendingDeletePeriode
            ? `Periode "${pendingDeletePeriode.nama}" akan dihapus dari organisasi ini.`
            : ""
        }
        pending={deletePeriode.isPending}
        onCancel={() => setPendingDeletePeriode(null)}
        onConfirm={async () => {
          if (!pendingDeletePeriode?.id) return;
          setPageError(null);
          try {
            await deletePeriode.mutateAsync(pendingDeletePeriode.id);
            setPendingDeletePeriode(null);
          } catch (err) {
            setPendingDeletePeriode(null);
            setPageError(resolveOrgError(err));
          }
        }}
      />

      <ConfirmDialog
        open={Boolean(pendingDeleteJabatan)}
        title="Hapus jabatan organisasi?"
        description={
          pendingDeleteJabatan
            ? `Jabatan "${pendingDeleteJabatan.jabatan}" akan dihapus dari periode ini.`
            : ""
        }
        pending={deleteJabatan.isPending}
        onCancel={() => setPendingDeleteJabatan(null)}
        onConfirm={async () => {
          if (!pendingDeleteJabatan?.id) return;
          setPageError(null);
          try {
            await deleteJabatan.mutateAsync(pendingDeleteJabatan.id);
            setPendingDeleteJabatan(null);
          } catch (err) {
            setPendingDeleteJabatan(null);
            setPageError(resolveOrgError(err));
          }
        }}
      />

      <ConfirmDialog
        open={Boolean(pendingRevokeInvite)}
        title="Cabut tautan undangan?"
        description={
          pendingRevokeInvite
            ? `Tautan undangan "${pendingRevokeInvite.code}" tidak akan dapat digunakan lagi setelah dicabut.`
            : ""
        }
        confirmLabel={t("org.revokeInvite")}
        pending={revokeInvite.isPending}
        onCancel={() => setPendingRevokeInvite(null)}
        onConfirm={async () => {
          if (!pendingRevokeInvite?.id) return;
          setPageError(null);
          try {
            await revokeInvite.mutateAsync(pendingRevokeInvite.id);
            setPendingRevokeInvite(null);
          } catch (err) {
            setPendingRevokeInvite(null);
            setPageError(resolveOrgError(err));
          }
        }}
      />

      <ConfirmDialog
        open={pendingLeaveOrg}
        title={t("org.leaveOrg")}
        description={t("org.confirmLeaveOrg")}
        confirmLabel={t("org.leaveOrg")}
        pending={removeMember.isPending}
        onCancel={() => setPendingLeaveOrg(false)}
        onConfirm={async () => {
          if (!me?.id) return;
          setPageError(null);
          try {
            await removeMember.mutateAsync(me.id);
            setPendingLeaveOrg(false);
            navigate("/org", { replace: true });
          } catch (err) {
            setPendingLeaveOrg(false);
            setPageError(resolveOrgError(err));
          }
        }}
      />

      <ConfirmDialog
        open={pendingDeleteOrg}
        title={t("org.deleteOrg")}
        description={t("org.confirmDeleteOrg")}
        confirmLabel={t("org.deleteOrg")}
        pending={deleteOrg.isPending}
        onCancel={() => setPendingDeleteOrg(false)}
        onConfirm={async () => {
          setPageError(null);
          try {
            await deleteOrg.mutateAsync(orgId);
            setPendingDeleteOrg(false);
            navigate("/org", { replace: true });
          } catch (err) {
            setPendingDeleteOrg(false);
            setPageError(resolveOrgError(err));
          }
        }}
      />
    </div>
  );

  function renderJabatanNodes(
    nodes: OrgJabatanNode<OrgJabatanResponse>[],
  ): ReactNode[] {
    return nodes.flatMap((node) => {
      const j = node.item;
      // ponytail: cap visual tree indentation at 48px (3 levels) to prevent horizontal overflow on 360px screens
      const indentPx = Math.min(node.depth, 3) * 16;
      const row = (
        <div
          key={j.id}
          className="flex flex-wrap items-center justify-between gap-3 py-3"
          style={{ paddingLeft: `${indentPx}px` }}
        >
          <div className="flex min-w-0 flex-1 flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
              {node.depth > 0 && <span className="text-xs text-muted">↳</span>}
              <span className="font-semibold text-fg">{j.jabatan}</span>
              {j.nama ? (
                <span className="text-sm text-fg">· {j.nama}</span>
              ) : (
                <Badge tone="neutral">Belum ditentukan</Badge>
              )}
              {j.userId && <Badge tone="neutral">Akun Pengguna</Badge>}
              {j.umatId && <Badge tone="neutral">Data Umat</Badge>}
              {!j.active && <Badge tone="danger">Nonaktif</Badge>}
            </div>
            {(j.startDate || j.endDate) && (
              <p className="text-xs text-muted">
                Masa jabatan: {j.startDate ?? "—"} – {j.endDate ?? "—"}
              </p>
            )}
          </div>

          {canManageOrg && (
            <div className="flex flex-wrap items-center gap-2">
              <Link
                to={`/org/${orgId}/periode/${selectedPeriodeId}/jabatan/new?parentId=${j.id}`}
                className={`${buttonClass("outline")} min-h-10 px-3 text-xs`}
              >
                <Plus aria-hidden className="size-3.5" />
                Tambah Sub-Jabatan
              </Link>
              <Link
                to={`/org/${orgId}/periode/${selectedPeriodeId}/jabatan/${j.id}/edit`}
                className={`${buttonClass("outline")} min-h-10 px-3 text-xs`}
              >
                <Pencil aria-hidden className="size-3.5" />
                Ubah
              </Link>
              <Button
                variant="danger"
                className="min-h-10 px-3 text-xs"
                disabled={deleteJabatan.isPending}
                onClick={() => setPendingDeleteJabatan(j)}
              >
                <Trash2 aria-hidden className="size-3.5" />
                Hapus
              </Button>
            </div>
          )}
        </div>
      );
      return [row, ...renderJabatanNodes(node.children)];
    });
  }
}
