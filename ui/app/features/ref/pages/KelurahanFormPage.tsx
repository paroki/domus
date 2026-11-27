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
  useCreateKelurahan,
  useDeleteKelurahan,
  useUpdateKelurahan,
} from "../api/actions";
import {
  useKecamatanSearch,
  useKelurahanDetail,
  useKotaSearch,
} from "../api/queries";
import { isSuperAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Formulir Kelurahan / Desa — Domus" }];
}

export default function KelurahanFormPage() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  const [kotaId, setKotaId] = useState(searchParams.get("kotaId") ?? "");
  const [kecamatanId, setKecamatanId] = useState(
    searchParams.get("kecamatanId") ?? "",
  );
  const [nama, setNama] = useState("");
  const [kode, setKode] = useState("");
  const [kodeTouched, setKodeTouched] = useState(false);
  const [kodePos, setKodePos] = useState("");
  const [active, setActive] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  const kotaQuery = useKotaSearch({}, canManage);
  const kotaList = kotaQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const kecQuery = useKecamatanSearch(
    { filter: kotaId ? { kotaId } : undefined },
    canManage,
  );
  const kecList = kecQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const detail = useKelurahanDetail(id ?? "", isEdit && canManage);

  const createMut = useCreateKelurahan();
  const updateMut = useUpdateKelurahan();
  const deleteMut = useDeleteKelurahan();

  useEffect(() => {
    if (!detail.data) return;
    setKecamatanId(detail.data.kecamatanId ?? "");
    setNama(detail.data.nama ?? "");
    setKode(detail.data.kode ?? "");
    setKodePos(detail.data.kodePos ?? "");
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
            kecamatanId,
            kode: kode.trim(),
            nama: nama.trim(),
            kodePos: kodePos.trim(),
            active,
          },
        });
      } else {
        await createMut.mutateAsync({
          kecamatanId,
          kode: (kode || slugifyCode(nama)).trim(),
          nama: nama.trim(),
          kodePos: kodePos.trim() || undefined,
        });
      }
      navigate(
        kecamatanId
          ? `/settings/sipil/kelurahan?kecamatanId=${kecamatanId}`
          : "/settings/sipil/kelurahan",
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
      navigate("/settings/sipil/kelurahan");
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
          { label: "Kelurahan / Desa", to: "/settings/sipil/kelurahan" },
          { label: isEdit ? "Ubah Kelurahan" : "Tambah Kelurahan" },
        ]}
        title={isEdit ? "Ubah Kelurahan / Desa" : "Tambah Kelurahan / Desa"}
        subtitle={
          isEdit
            ? "Perbarui data kelurahan/desa atau status aktifnya."
            : "Tambahkan kelurahan atau desa baru di bawah kecamatan."
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
            <InductiveFormSection title="Data Kelurahan / Desa">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kelurahan-kota"
                    className="text-xs font-semibold text-muted"
                  >
                    Filter Kota / Kabupaten
                  </label>
                  <select
                    id="kelurahan-kota"
                    value={kotaId}
                    onChange={(e) => {
                      setKotaId(e.target.value);
                      setKecamatanId("");
                    }}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  >
                    <option value="">Semua Kota</option>
                    {kotaList.map((k) => (
                      <option key={k.id} value={k.id}>
                        {k.nama}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kelurahan-kecamatan"
                    className="text-xs font-semibold text-muted"
                  >
                    Kecamatan Induk
                  </label>
                  <select
                    id="kelurahan-kecamatan"
                    required
                    value={kecamatanId}
                    onChange={(e) => setKecamatanId(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  >
                    <option value="">Pilih kecamatan…</option>
                    {kecList.map((k) => (
                      <option key={k.id} value={k.id}>
                        {k.nama}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kelurahan-nama"
                    className="text-xs font-semibold text-muted"
                  >
                    Nama Kelurahan / Desa
                  </label>
                  <input
                    id="kelurahan-nama"
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
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kelurahan-kode"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode Kemendagri
                  </label>
                  <input
                    id="kelurahan-kode"
                    required
                    value={kode}
                    onChange={(e) => {
                      setKodeTouched(true);
                      setKode(e.target.value);
                    }}
                    placeholder="mis. 31710101"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 font-mono text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kelurahan-kodepos"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode Pos (opsional)
                  </label>
                  <input
                    id="kelurahan-kodepos"
                    maxLength={10}
                    value={kodePos}
                    onChange={(e) => setKodePos(e.target.value)}
                    placeholder="mis. 10110"
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
                  Kelurahan / Desa Aktif
                </label>
              ) : null}

              <div className="flex flex-col-reverse items-stretch justify-end gap-2 pt-2 sm:flex-row sm:items-center">
                <Link
                  to="/settings/sipil/kelurahan"
                  className={`${buttonClass("outline")} w-full sm:w-auto`}
                >
                  Batal
                </Link>
                <Button
                  type="submit"
                  disabled={pending}
                  className="w-full sm:w-auto"
                >
                  {pending ? "Menyimpan…" : "Simpan Kelurahan"}
                </Button>
              </div>
            </InductiveFormSection>
          </form>

          {isEdit ? (
            <InductiveDangerZone
              title="Hapus Kelurahan / Desa"
              description="Tidak dapat dihapus jika masih digunakan oleh paroki."
              action={
                <Button
                  variant="danger"
                  disabled={deleteMut.isPending}
                  onClick={() => setConfirmDelete(true)}
                >
                  Hapus Kelurahan
                </Button>
              }
            />
          ) : null}

          <ConfirmDialog
            open={confirmDelete}
            title="Hapus kelurahan/desa ini?"
            description="Pastikan tidak ada paroki yang masih terhubung dengan kelurahan ini."
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
