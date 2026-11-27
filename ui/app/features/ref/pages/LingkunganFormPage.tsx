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
  useCreateLingkungan,
  useDeleteLingkungan,
  useUpdateLingkungan,
} from "../api/actions";
import { useLingkunganDetail, useWilayahSearch } from "../api/queries";
import { isParokiAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Formulir Lingkungan — Domus" }];
}

export default function LingkunganFormPage() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isParokiAdmin(me?.roles);
  const hasTenant = Boolean(me?.activeParokiId);

  const wilayahQuery = useWilayahSearch({}, canManage && hasTenant);
  const wilayahList =
    wilayahQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const detail = useLingkunganDetail(
    id ?? "",
    isEdit && canManage && hasTenant,
  );
  const createMut = useCreateLingkungan();
  const updateMut = useUpdateLingkungan();
  const deleteMut = useDeleteLingkungan();

  const [wilayahId, setWilayahId] = useState(
    searchParams.get("wilayahId") ?? "",
  );
  const [kode, setKode] = useState("");
  const [kodeTouched, setKodeTouched] = useState(false);
  const [nama, setNama] = useState("");
  const [pelindung, setPelindung] = useState("");
  const [ketua, setKetua] = useState("");
  const [teleponKetua, setTeleponKetua] = useState("");
  const [active, setActive] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    if (!detail.data) return;
    setWilayahId(detail.data.wilayahId ?? "");
    setKode(detail.data.kode ?? "");
    setNama(detail.data.nama ?? "");
    setPelindung(detail.data.pelindung ?? "");
    setKetua(detail.data.ketua ?? "");
    setTeleponKetua(detail.data.teleponKetua ?? "");
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
            wilayahId,
            kode: kode.trim(),
            nama: nama.trim(),
            pelindung: pelindung.trim(),
            ketua: ketua.trim(),
            teleponKetua: teleponKetua.trim(),
            active,
          },
        });
      } else {
        await createMut.mutateAsync({
          wilayahId,
          kode: (kode || slugifyCode(nama)).trim(),
          nama: nama.trim(),
          pelindung: pelindung.trim() || undefined,
          ketua: ketua.trim() || undefined,
          teleponKetua: teleponKetua.trim() || undefined,
        });
      }
      navigate(
        wilayahId
          ? `/settings/paroki/lingkungan?wilayahId=${wilayahId}`
          : "/settings/paroki/lingkungan",
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
      navigate("/settings/paroki/lingkungan");
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
          { label: "Administrasi Paroki", to: "/settings/paroki" },
          { label: "Lingkungan", to: "/settings/paroki/lingkungan" },
          { label: isEdit ? "Ubah Lingkungan" : "Tambah Lingkungan" },
        ]}
        title={isEdit ? "Ubah Lingkungan" : "Tambah Lingkungan"}
        subtitle={
          isEdit
            ? "Perbarui data lingkungan atau kelola status aktifnya."
            : "Tambahkan lingkungan baru di bawah wilayah/stasi."
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Admin Paroki yang dapat mengelola lingkungan.
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
            <InductiveFormSection title="Data Lingkungan">
              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor="lingkungan-wilayah"
                  className="text-xs font-semibold text-muted"
                >
                  Wilayah / Stasi Induk
                </label>
                <select
                  id="lingkungan-wilayah"
                  required
                  value={wilayahId}
                  onChange={(e) => setWilayahId(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                >
                  <option value="">Pilih wilayah…</option>
                  {wilayahList.map((w) => (
                    <option key={w.id} value={w.id}>
                      {w.nama}
                    </option>
                  ))}
                </select>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="lingkungan-kode"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode Lingkungan
                  </label>
                  <input
                    id="lingkungan-kode"
                    required
                    value={kode}
                    onChange={(e) => {
                      setKodeTouched(true);
                      setKode(e.target.value);
                    }}
                    placeholder="mis. ling-01"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 font-mono text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="lingkungan-nama"
                    className="text-xs font-semibold text-muted"
                  >
                    Nama Lingkungan
                  </label>
                  <input
                    id="lingkungan-nama"
                    required
                    value={nama}
                    onChange={(e) => {
                      const next = e.target.value;
                      setNama(next);
                      if (!isEdit && !kodeTouched) {
                        setKode(slugifyCode(next));
                      }
                    }}
                    placeholder="mis. Lingkungan Petrus"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="lingkungan-pelindung"
                    className="text-xs font-semibold text-muted"
                  >
                    Pelindung (opsional)
                  </label>
                  <input
                    id="lingkungan-pelindung"
                    value={pelindung}
                    onChange={(e) => setPelindung(e.target.value)}
                    placeholder="mis. Santo Petrus"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="lingkungan-ketua"
                    className="text-xs font-semibold text-muted"
                  >
                    Ketua (opsional)
                  </label>
                  <input
                    id="lingkungan-ketua"
                    value={ketua}
                    onChange={(e) => setKetua(e.target.value)}
                    placeholder="mis. Ibu Maria"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="lingkungan-telp"
                    className="text-xs font-semibold text-muted"
                  >
                    Telepon Ketua (opsional)
                  </label>
                  <input
                    id="lingkungan-telp"
                    value={teleponKetua}
                    onChange={(e) => setTeleponKetua(e.target.value)}
                    placeholder="08123456789"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
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
                  Lingkungan Aktif
                </label>
              ) : null}

              <div className="flex flex-col-reverse items-stretch justify-end gap-2 pt-2 sm:flex-row sm:items-center">
                <Link
                  to="/settings/paroki/lingkungan"
                  className={`${buttonClass("outline")} w-full sm:w-auto`}
                >
                  Batal
                </Link>
                <Button
                  type="submit"
                  disabled={pending}
                  className="w-full sm:w-auto"
                >
                  {pending ? "Menyimpan…" : "Simpan Lingkungan"}
                </Button>
              </div>
            </InductiveFormSection>
          </form>

          {isEdit ? (
            <InductiveDangerZone
              title="Hapus Lingkungan"
              description="Menghapus lingkungan akan menonaktifkannya dari paroki aktif."
              action={
                <Button
                  variant="danger"
                  disabled={deleteMut.isPending}
                  onClick={() => setConfirmDelete(true)}
                >
                  Hapus Lingkungan
                </Button>
              }
            />
          ) : null}

          <ConfirmDialog
            open={confirmDelete}
            title="Hapus lingkungan ini?"
            description="Data lingkungan akan dinonaktifkan dari paroki."
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
