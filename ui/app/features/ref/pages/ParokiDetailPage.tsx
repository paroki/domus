import { Trash2 } from "lucide-react";
import { useState } from "react";
import { useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import {
  InductiveDangerZone,
  InductiveFormSection,
  InductivePageHeader,
} from "~/shared/ui/iui";
import {
  useAddParokiMember,
  useDeleteParoki,
  useRemoveParokiMember,
  useUpdateParoki,
  useUpdateParokiMember,
} from "../api/actions";
import { useParokiDetail, useParokiMembersSearch } from "../api/queries";
import type { UpdateParokiRequest } from "../api/types";
import {
  ParokiDomisiliCard,
  ParokiIdentitasCard,
  ParokiKontakCard,
  type ParokiSectionKey,
  ParokiStrukturCard,
} from "../components/ParokiSectionCards";
import { isSuperAdmin } from "../model";

export function meta() {
  return [{ title: "Detail Paroki — Domus" }];
}

export default function ParokiDetailPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  const detail = useParokiDetail(id, canManage && Boolean(id));
  const membersQuery = useParokiMembersSearch(id, {}, canManage && Boolean(id));
  const members = membersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const updateMut = useUpdateParoki();
  const deleteMut = useDeleteParoki();
  const addMemberMut = useAddParokiMember();
  const updateMemberMut = useUpdateParokiMember();
  const removeMemberMut = useRemoveParokiMember();

  const [editingSection, setEditingSection] = useState<ParokiSectionKey | null>(
    null,
  );
  const [adminEmail, setAdminEmail] = useState("");
  const [adminName, setAdminName] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [statusMsg, setStatusMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  async function handleSaveSection(body: UpdateParokiRequest) {
    setErrorMsg(null);
    setStatusMsg(null);
    try {
      await updateMut.mutateAsync({ id, body });
      setEditingSection(null);
      setStatusMsg("Data paroki berhasil diperbarui.");
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal menyimpan.");
    }
  }

  async function handleAddAdmin(e: React.FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    setStatusMsg(null);
    try {
      await addMemberMut.mutateAsync({
        parokiId: id,
        body: {
          email: adminEmail.trim(),
          name: adminName.trim() || undefined,
          role: "admin",
        },
      });
      setAdminEmail("");
      setAdminName("");
      setStatusMsg("Admin paroki berhasil ditambahkan.");
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal menunjuk admin.");
    }
  }

  async function handleDelete() {
    setErrorMsg(null);
    try {
      await deleteMut.mutateAsync(id);
      navigate("/settings/gerejawi/paroki");
    } catch (err) {
      setConfirmDelete(false);
      setErrorMsg(err instanceof Error ? err.message : "Gagal menghapus.");
    }
  }

  const paroki = detail.data;

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Struktur Gerejawi", to: "/settings/gerejawi" },
          { label: "Paroki", to: "/settings/gerejawi/paroki" },
          { label: paroki?.nama ?? "Detail Paroki" },
        ]}
        title={paroki?.nama ?? "Detail Paroki"}
        subtitle="Kelola struktur gerejawi, identitas, domisili, kontak, dan admin paroki per bagian."
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola detail paroki.
        </div>
      ) : null}

      {errorMsg ? (
        <div
          role="alert"
          className="rounded-xl border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
        >
          {errorMsg}
        </div>
      ) : null}

      {statusMsg ? (
        <div
          role="status"
          className="rounded-xl border border-success-fg/30 bg-success-bg px-4 py-3 text-sm text-success-fg"
        >
          {statusMsg}
        </div>
      ) : null}

      {canManage && detail.isLoading ? (
        <p className="text-sm text-muted">Memuat data…</p>
      ) : null}

      {canManage && paroki ? (
        <>
          <ParokiStrukturCard
            paroki={paroki}
            editing={editingSection === "struktur"}
            pending={updateMut.isPending}
            onEdit={() => setEditingSection("struktur")}
            onCancel={() => setEditingSection(null)}
            onSave={handleSaveSection}
          />

          <ParokiIdentitasCard
            paroki={paroki}
            editing={editingSection === "identitas"}
            pending={updateMut.isPending}
            onEdit={() => setEditingSection("identitas")}
            onCancel={() => setEditingSection(null)}
            onSave={handleSaveSection}
          />

          <ParokiDomisiliCard
            paroki={paroki}
            editing={editingSection === "domisili"}
            pending={updateMut.isPending}
            onEdit={() => setEditingSection("domisili")}
            onCancel={() => setEditingSection(null)}
            onSave={handleSaveSection}
          />

          <ParokiKontakCard
            paroki={paroki}
            editing={editingSection === "kontak"}
            pending={updateMut.isPending}
            onEdit={() => setEditingSection("kontak")}
            onCancel={() => setEditingSection(null)}
            onSave={handleSaveSection}
          />

          <InductiveFormSection
            title="Admin &amp; Anggota Paroki"
            description="Tunjuk pengguna sebagai Admin Paroki atau ubah perannya."
          >
            <form
              onSubmit={handleAddAdmin}
              className="grid grid-cols-1 gap-3 sm:grid-cols-3"
            >
              <div className="flex flex-col gap-1">
                <label
                  htmlFor="admin-email"
                  className="text-xs font-semibold text-muted"
                >
                  Email Pengguna
                </label>
                <input
                  id="admin-email"
                  type="email"
                  required
                  value={adminEmail}
                  onChange={(e) => setAdminEmail(e.target.value)}
                  placeholder="admin@paroki.org"
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                />
              </div>
              <div className="flex flex-col gap-1">
                <label
                  htmlFor="admin-name"
                  className="text-xs font-semibold text-muted"
                >
                  Nama (opsional)
                </label>
                <input
                  id="admin-name"
                  value={adminName}
                  onChange={(e) => setAdminName(e.target.value)}
                  placeholder="Nama lengkap"
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                />
              </div>
              <div className="flex items-end">
                <Button
                  type="submit"
                  disabled={addMemberMut.isPending}
                  className="w-full"
                >
                  Tunjuk Admin
                </Button>
              </div>
            </form>

            {members.length === 0 ? (
              <p className="text-sm text-muted">Belum ada anggota terdaftar.</p>
            ) : (
              <div className="flex flex-col gap-2">
                {members.map((m) => (
                  <div
                    key={m.id}
                    className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-line bg-canvas px-4 py-3"
                  >
                    <div className="flex flex-col">
                      <span className="text-sm font-semibold text-fg">
                        {m.name || m.email}
                      </span>
                      <span className="text-xs text-muted">{m.email}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge tone={m.role === "admin" ? "success" : "neutral"}>
                        {m.role === "admin" ? "Admin Paroki" : "Anggota"}
                      </Badge>
                      <Button
                        variant="outline"
                        onClick={() => {
                          if (!m.userId) return;
                          updateMemberMut.mutate({
                            parokiId: id,
                            userId: m.userId,
                            body: {
                              role: m.role === "admin" ? "member" : "admin",
                            },
                          });
                        }}
                      >
                        {m.role === "admin"
                          ? "Jadikan Anggota"
                          : "Jadikan Admin"}
                      </Button>
                      <Button
                        variant="ghost"
                        aria-label="Hapus anggota"
                        onClick={() => {
                          if (!m.userId) return;
                          removeMemberMut.mutate({
                            parokiId: id,
                            userId: m.userId,
                          });
                        }}
                      >
                        <Trash2 aria-hidden className="size-4 text-danger-fg" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </InductiveFormSection>

          <InductiveDangerZone
            title="Hapus Paroki"
            description="Menghapus paroki akan menonaktifkannya. Tidak dapat dilakukan jika masih memiliki data turunan."
            action={
              <Button
                variant="danger"
                disabled={deleteMut.isPending}
                onClick={() => setConfirmDelete(true)}
              >
                Hapus Paroki
              </Button>
            }
          />

          <ConfirmDialog
            open={confirmDelete}
            title="Hapus paroki ini?"
            description="Pastikan tidak ada data turunan yang masih aktif di paroki ini."
            confirmLabel="Ya, Hapus"
            pending={deleteMut.isPending}
            onConfirm={handleDelete}
            onCancel={() => setConfirmDelete(false)}
          />
        </>
      ) : null}
    </div>
  );
}
