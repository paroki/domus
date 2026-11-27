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
import { useCreateKota, useDeleteKota, useUpdateKota } from "../api/actions";
import { useKotaDetail } from "../api/queries";
import { isSuperAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Formulir Kota / Kabupaten — Domus" }];
}

export default function KotaFormPage() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  const detail = useKotaDetail(id ?? "", isEdit && canManage);
  const createMut = useCreateKota();
  const updateMut = useUpdateKota();
  const deleteMut = useDeleteKota();

  const [nama, setNama] = useState("");
  const [kode, setKode] = useState("");
  const [kodeTouched, setKodeTouched] = useState(false);
  const [active, setActive] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    if (!detail.data) return;
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
          body: { kode: kode.trim(), nama: nama.trim(), active },
        });
      } else {
        await createMut.mutateAsync({
          kode: (kode || slugifyCode(nama)).trim(),
          nama: nama.trim(),
        });
      }
      navigate("/settings/sipil/kota");
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal menyimpan.");
    }
  }

  async function handleDelete() {
    if (!id) return;
    setErrorMsg(null);
    try {
      await deleteMut.mutateAsync(id);
      navigate("/settings/sipil/kota");
    } catch (err) {
      setConfirmDelete(false);
      setErrorMsg(err instanceof Error ? err.message : "Gagal menghapus.");
    }
  }

  const pending = createMut.isPending || updateMut.isPending;

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Wilayah Sipil", to: "/settings/sipil" },
          { label: "Kota / Kabupaten", to: "/settings/sipil/kota" },
          { label: isEdit ? "Ubah Kota" : "Tambah Kota" },
        ]}
        title={isEdit ? "Ubah Kota / Kabupaten" : "Tambah Kota / Kabupaten"}
        subtitle={
          isEdit
            ? "Perbarui data kota/kabupaten atau status aktifnya."
            : "Tambahkan kota atau kabupaten baru ke referensi wilayah sipil."
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
            <InductiveFormSection title="Data Kota / Kabupaten">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kota-nama"
                    className="text-xs font-semibold text-muted"
                  >
                    Nama Kota / Kabupaten
                  </label>
                  <input
                    id="kota-nama"
                    required
                    value={nama}
                    onChange={(e) => {
                      const next = e.target.value;
                      setNama(next);
                      if (!isEdit && !kodeTouched) {
                        setKode(slugifyCode(next));
                      }
                    }}
                    placeholder="mis. Kota Jakarta Pusat"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kota-kode"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode Kemendagri
                  </label>
                  <input
                    id="kota-kode"
                    required
                    value={kode}
                    onChange={(e) => {
                      setKodeTouched(true);
                      setKode(e.target.value);
                    }}
                    placeholder="mis. 3171"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 font-mono text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
              </div>

              {isEdit ? (
                <label className="inline-flex min-h-11 items-center gap-2 text-sm text-fg">
                  <input
                    type="checkbox"
                    checked={active}
                    onChange={(e) => setActive(e.target.checked)}
                    className="size-4 accent-primary"
                  />
                  Kota / Kabupaten Aktif
                </label>
              ) : null}

              <div className="flex flex-col-reverse items-stretch justify-end gap-2 pt-2 sm:flex-row sm:items-center">
                <Link
                  to="/settings/sipil/kota"
                  className={`${buttonClass("outline")} w-full sm:w-auto`}
                >
                  Batal
                </Link>
                <Button
                  type="submit"
                  disabled={pending}
                  className="w-full sm:w-auto"
                >
                  {pending ? "Menyimpan…" : "Simpan Kota"}
                </Button>
              </div>
            </InductiveFormSection>
          </form>

          {isEdit ? (
            <InductiveDangerZone
              title="Hapus Kota / Kabupaten"
              description="Tidak dapat dihapus jika masih memiliki kecamatan atau paroki."
              action={
                <Button
                  variant="danger"
                  disabled={deleteMut.isPending}
                  onClick={() => setConfirmDelete(true)}
                >
                  Hapus Kota
                </Button>
              }
            />
          ) : null}

          <ConfirmDialog
            open={confirmDelete}
            title="Hapus kota/kabupaten ini?"
            description="Pastikan tidak ada kecamatan atau paroki yang masih terhubung."
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
