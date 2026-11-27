import { Users } from "lucide-react";
import { useState } from "react";
import { useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button } from "~/shared/ui/Button";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import {
  InductiveDangerZone,
  InductivePageHeader,
  InductiveStepCard,
} from "~/shared/ui/iui";
import { useDeleteParoki, useUpdateParoki } from "../api/actions";
import { useParokiDetail } from "../api/queries";
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

  const updateMut = useUpdateParoki();
  const deleteMut = useDeleteParoki();

  const [editingSection, setEditingSection] = useState<ParokiSectionKey | null>(
    null,
  );
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
    <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Struktur Gerejawi", to: "/settings/gerejawi" },
          { label: "Paroki", to: "/settings/gerejawi/paroki" },
          { label: paroki?.nama ?? "Detail Paroki" },
        ]}
        title={paroki?.nama ?? "Detail Paroki"}
        subtitle="Kelola struktur gerejawi, identitas, domisili, dan kontak paroki per bagian."
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

          <InductiveStepCard
            to="/users"
            title="Kelola Admin & Anggota Paroki di Direktori Pengguna"
            description="Penunjukan Admin Paroki, perubahan peran, dan pengelolaan keanggotaan paroki dilakukan terpusat di Direktori Pengguna."
            icon={Users}
            plate={4}
          />

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
