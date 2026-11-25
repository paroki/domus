import { useEffect, useState } from "react";
import { Link, useNavigate, useParams, useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button, buttonClass } from "~/shared/ui/Button";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import {
  InductiveDangerZone,
  InductiveFormSection,
  InductivePageHeader,
} from "~/shared/ui/iui";
import {
  useCreateKecamatan,
  useDeleteKecamatan,
  useUpdateKecamatan,
} from "../api/actions";
import { useKecamatanDetail, useKotaSearch } from "../api/queries";
import { isSuperAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Formulir Kecamatan — Domus" }];
}

export default function KecamatanFormPage() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  const kotaQuery = useKotaSearch({}, canManage);
  const kotaList = kotaQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const detail = useKecamatanDetail(id ?? "", isEdit && canManage);

  const createMut = useCreateKecamatan();
  const updateMut = useUpdateKecamatan();
  const deleteMut = useDeleteKecamatan();

  const [kotaId, setKotaId] = useState(searchParams.get("kotaId") ?? "");
  const [nama, setNama] = useState("");
  const [kode, setKode] = useState("");
  const [kodeTouched, setKodeTouched] = useState(false);
  const [active, setActive] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    if (!detail.data) return;
    setKotaId(detail.data.kotaId ?? "");
    setNama(detail.data.nama ?? "");
    setKode(detail.data.kode ?? "");
    setActive(detail.data.active ?? true);
  }, [detail.data]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      if (isEdit && id) {
        await updateMut.mutateAsync({
          id,
          body: { kotaId, kode: kode.trim(), nama: nama.trim(), active },
        });
      } else {
        await createMut.mutateAsync({
          kotaId,
          kode: (kode || slugifyCode(nama)).trim(),
          nama: nama.trim(),
        });
      }
      navigate(
        kotaId
          ? `/settings/sipil/kecamatan?kotaId=${kotaId}`
          : "/settings/sipil/kecamatan",
      );
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal menyimpan.");
    }
  }

  async function handleDelete() {
    if (!id) return;
    setErrorMsg(null);
    try {
      await deleteMut.mutateAsync(id);
      navigate("/settings/sipil/kecamatan");
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
          { label: "Wilayah Sipil", to: "/settings/sipil" },
          { label: "Kecamatan", to: "/settings/sipil/kecamatan" },
          { label: isEdit ? "Ubah Kecamatan" : "Tambah Kecamatan" },
        ]}
        title={isEdit ? "Ubah Kecamatan" : "Tambah Kecamatan"}
        subtitle={
          isEdit
            ? "Perbarui data kecamatan atau status aktifnya."
            : "Tambahkan kecamatan baru di bawah kota/kabupaten."
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola wilayah sipil.
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

      {canManage ? (
        <>
          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            <InductiveFormSection title="Data Kecamatan">
              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor="kecamatan-kota"
                  className="text-xs font-semibold text-muted"
                >
                  Kota / Kabupaten Induk
                </label>
                <select
                  id="kecamatan-kota"
                  required
                  value={kotaId}
                  onChange={(e) => setKotaId(e.target.value)}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                >
                  <option value="">Pilih kota/kabupaten…</option>
                  {kotaList.map((k) => (
                    <option key={k.id} value={k.id}>
                      {k.nama}
                    </option>
                  ))}
                </select>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kecamatan-nama"
                    className="text-xs font-semibold text-muted"
                  >
                    Nama Kecamatan
                  </label>
                  <input
                    id="kecamatan-nama"
                    required
                    value={nama}
                    onChange={(e) => {
                      const next = e.target.value;
                      setNama(next);
                      if (!isEdit && !kodeTouched) {
                        setKode(slugifyCode(next));
                      }
                    }}
                    placeholder="mis. Gambir"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kecamatan-kode"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode Kemendagri
                  </label>
                  <input
                    id="kecamatan-kode"
                    required
                    value={kode}
                    onChange={(e) => {
                      setKodeTouched(true);
                      setKode(e.target.value);
                    }}
                    placeholder="mis. 317101"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 font-mono text-sm text-fg focus:border-primary focus:outline-none"
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
                  Kecamatan Aktif
                </label>
              ) : null}

              <div className="flex items-center justify-end gap-2 pt-2">
                <Link
                  to="/settings/sipil/kecamatan"
                  className={buttonClass("outline")}
                >
                  Batal
                </Link>
                <Button type="submit" disabled={pending}>
                  {pending ? "Menyimpan…" : "Simpan Kecamatan"}
                </Button>
              </div>
            </InductiveFormSection>
          </form>

          {isEdit ? (
            <InductiveDangerZone
              title="Hapus Kecamatan"
              description="Tidak dapat dihapus jika masih memiliki kelurahan atau paroki."
              action={
                <Button
                  variant="danger"
                  disabled={deleteMut.isPending}
                  onClick={() => setConfirmDelete(true)}
                >
                  Hapus Kecamatan
                </Button>
              }
            />
          ) : null}

          <ConfirmDialog
            open={confirmDelete}
            title="Hapus kecamatan ini?"
            description="Pastikan tidak ada kelurahan atau paroki yang masih terhubung."
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
