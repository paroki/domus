import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button, buttonClass } from "~/shared/ui/Button";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import {
  InductiveDangerZone,
  InductiveFormSection,
  InductivePageHeader,
} from "~/shared/ui/iui";
import {
  useCreateWilayah,
  useDeleteWilayah,
  useUpdateWilayah,
} from "../api/actions";
import { useWilayahDetail } from "../api/queries";
import { isParokiAdmin } from "../model";

export function meta() {
  return [{ title: "Formulir Wilayah — Domus" }];
}

export default function WilayahFormPage() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isParokiAdmin(me?.roles);
  const hasTenant = Boolean(me?.activeParokiId);

  const detail = useWilayahDetail(id ?? "", isEdit && canManage && hasTenant);
  const createMut = useCreateWilayah();
  const updateMut = useUpdateWilayah();
  const deleteMut = useDeleteWilayah();

  const [jenis, setJenis] = useState<"wilayah" | "stasi">("wilayah");
  const [kode, setKode] = useState("");
  const [nama, setNama] = useState("");
  const [koordinator, setKoordinator] = useState("");
  const [teleponKoordinator, setTeleponKoordinator] = useState("");
  const [active, setActive] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    if (!detail.data) return;
    setJenis(detail.data.jenis === "stasi" ? "stasi" : "wilayah");
    setKode(detail.data.kode ?? "");
    setNama(detail.data.nama ?? "");
    setKoordinator(detail.data.koordinator ?? "");
    setTeleponKoordinator(detail.data.teleponKoordinator ?? "");
    setActive(detail.data.active ?? true);
  }, [detail.data]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      if (isEdit && id) {
        await updateMut.mutateAsync({
          id,
          body: {
            jenis,
            kode: kode.trim(),
            nama: nama.trim(),
            koordinator: koordinator.trim(),
            teleponKoordinator: teleponKoordinator.trim(),
            active,
          },
        });
      } else {
        await createMut.mutateAsync({
          jenis,
          kode: kode.trim(),
          nama: nama.trim(),
          koordinator: koordinator.trim() || undefined,
          teleponKoordinator: teleponKoordinator.trim() || undefined,
        });
      }
      navigate("/settings/paroki/wilayah");
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal menyimpan.");
    }
  }

  async function handleDelete() {
    if (!id) return;
    setErrorMsg(null);
    try {
      await deleteMut.mutateAsync(id);
      navigate("/settings/paroki/wilayah");
    } catch (err) {
      setConfirmDelete(false);
      setErrorMsg(err instanceof Error ? err.message : "Gagal menghapus.");
    }
  }

  const pending = createMut.isPending || updateMut.isPending;

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Administrasi Paroki", to: "/settings/paroki" },
          { label: "Wilayah & Stasi", to: "/settings/paroki/wilayah" },
          { label: isEdit ? "Ubah Wilayah" : "Tambah Wilayah" },
        ]}
        title={isEdit ? "Ubah Wilayah / Stasi" : "Tambah Wilayah / Stasi"}
        subtitle={
          isEdit
            ? "Perbarui data wilayah/stasi atau kelola status aktifnya."
            : "Tambahkan wilayah atau stasi baru pada paroki aktif."
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Admin Paroki yang dapat mengelola wilayah dan stasi.
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

      {canManage && hasTenant ? (
        <>
          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            <InductiveFormSection title="Data Wilayah / Stasi">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wilayah-jenis"
                    className="text-xs font-semibold text-muted"
                  >
                    Jenis Wilayah
                  </label>
                  <select
                    id="wilayah-jenis"
                    value={jenis}
                    onChange={(e) =>
                      setJenis(e.target.value as "wilayah" | "stasi")
                    }
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  >
                    <option value="wilayah">Wilayah</option>
                    <option value="stasi">Stasi</option>
                  </select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wilayah-kode"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode
                  </label>
                  <input
                    id="wilayah-kode"
                    required
                    value={kode}
                    onChange={(e) => setKode(e.target.value)}
                    placeholder="mis. wil-01"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 font-mono text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wilayah-nama"
                    className="text-xs font-semibold text-muted"
                  >
                    Nama Wilayah / Stasi
                  </label>
                  <input
                    id="wilayah-nama"
                    required
                    value={nama}
                    onChange={(e) => setNama(e.target.value)}
                    placeholder="mis. Wilayah Santo Lukas"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wilayah-koordinator"
                    className="text-xs font-semibold text-muted"
                  >
                    Nama Koordinator (opsional)
                  </label>
                  <input
                    id="wilayah-koordinator"
                    value={koordinator}
                    onChange={(e) => setKoordinator(e.target.value)}
                    placeholder="mis. Bapak Anton"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wilayah-telp"
                    className="text-xs font-semibold text-muted"
                  >
                    Telepon Koordinator (opsional)
                  </label>
                  <input
                    id="wilayah-telp"
                    value={teleponKoordinator}
                    onChange={(e) => setTeleponKoordinator(e.target.value)}
                    placeholder="08123456789"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
              </div>

              {isEdit ? (
                <label className="inline-flex items-center gap-2 text-sm text-fg">
                  <input
                    type="checkbox"
                    checked={active}
                    onChange={(e) => setActive(e.target.checked)}
                    className="size-4 accent-primary"
                  />
                  Wilayah Aktif
                </label>
              ) : null}

              <div className="flex items-center justify-end gap-2 pt-2">
                <Link
                  to="/settings/paroki/wilayah"
                  className={buttonClass("outline")}
                >
                  Batal
                </Link>
                <Button type="submit" disabled={pending}>
                  {pending ? "Menyimpan…" : "Simpan Wilayah"}
                </Button>
              </div>
            </InductiveFormSection>
          </form>

          {isEdit ? (
            <InductiveDangerZone
              title="Hapus Wilayah"
              description="Menghapus wilayah tidak dapat dilakukan jika masih memiliki lingkungan aktif."
              action={
                <Button
                  variant="danger"
                  disabled={deleteMut.isPending}
                  onClick={() => setConfirmDelete(true)}
                >
                  Hapus Wilayah
                </Button>
              }
            />
          ) : null}

          <ConfirmDialog
            open={confirmDelete}
            title="Hapus wilayah ini?"
            description="Pastikan tidak ada lingkungan yang masih terhubung dengan wilayah ini."
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
