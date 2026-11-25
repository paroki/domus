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
  useCreateKevikepan,
  useDeleteKevikepan,
  useUpdateKevikepan,
} from "../api/actions";
import { useKeuskupanSearch, useKevikepanDetail } from "../api/queries";
import { isSuperAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Formulir Kevikepan — Domus" }];
}

export default function KevikepanFormPage() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  const keuskupanQuery = useKeuskupanSearch({}, canManage);
  const keuskupanList =
    keuskupanQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const detail = useKevikepanDetail(id ?? "", isEdit && canManage);

  const createMut = useCreateKevikepan();
  const updateMut = useUpdateKevikepan();
  const deleteMut = useDeleteKevikepan();

  const [keuskupanId, setKeuskupanId] = useState(
    searchParams.get("keuskupanId") ?? "",
  );
  const [nama, setNama] = useState("");
  const [kode, setKode] = useState("");
  const [kodeTouched, setKodeTouched] = useState(false);
  const [vikep, setVikep] = useState("");
  const [active, setActive] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    if (!detail.data) return;
    setKeuskupanId(detail.data.keuskupanId ?? "");
    setNama(detail.data.nama ?? "");
    setKode(detail.data.kode ?? "");
    setVikep(detail.data.vikep ?? "");
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
            keuskupanId,
            kode: kode.trim(),
            nama: nama.trim(),
            vikep: vikep.trim(),
            active,
          },
        });
      } else {
        await createMut.mutateAsync({
          keuskupanId,
          kode: (kode || slugifyCode(nama)).trim(),
          nama: nama.trim(),
          vikep: vikep.trim() || undefined,
        });
      }
      navigate(
        keuskupanId
          ? `/settings/gerejawi/kevikepan?keuskupanId=${keuskupanId}`
          : "/settings/gerejawi/kevikepan",
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
      navigate("/settings/gerejawi/kevikepan");
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
          { label: "Struktur Gerejawi", to: "/settings/gerejawi" },
          { label: "Kevikepan", to: "/settings/gerejawi/kevikepan" },
          { label: isEdit ? "Ubah Kevikepan" : "Tambah Kevikepan" },
        ]}
        title={isEdit ? "Ubah Kevikepan" : "Tambah Kevikepan"}
        subtitle={
          isEdit
            ? "Perbarui data kevikepan atau kelola status aktifnya."
            : "Tambahkan kevikepan baru di bawah keuskupan."
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola kevikepan.
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
            <InductiveFormSection title="Identitas Kevikepan">
              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor="kevikepan-keuskupan"
                  className="text-xs font-semibold text-muted"
                >
                  Keuskupan Induk
                </label>
                <select
                  id="kevikepan-keuskupan"
                  required
                  value={keuskupanId}
                  onChange={(e) => setKeuskupanId(e.target.value)}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                >
                  <option value="">Pilih keuskupan…</option>
                  {keuskupanList.map((k) => (
                    <option key={k.id} value={k.id}>
                      {k.nama}
                    </option>
                  ))}
                </select>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kevikepan-nama"
                    className="text-xs font-semibold text-muted"
                  >
                    Nama Kevikepan
                  </label>
                  <input
                    id="kevikepan-nama"
                    required
                    value={nama}
                    onChange={(e) => {
                      const next = e.target.value;
                      setNama(next);
                      if (!isEdit && !kodeTouched) {
                        setKode(slugifyCode(next));
                      }
                    }}
                    placeholder="mis. Kevikepan Jakarta Pusat"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="kevikepan-kode"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode Kevikepan
                  </label>
                  <input
                    id="kevikepan-kode"
                    required
                    value={kode}
                    onChange={(e) => {
                      setKodeTouched(true);
                      setKode(e.target.value);
                    }}
                    placeholder="mis. pusat"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 font-mono text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
              </div>

              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor="kevikepan-vikep"
                  className="text-xs font-semibold text-muted"
                >
                  Vikaris Episkopal (opsional)
                </label>
                <input
                  id="kevikepan-vikep"
                  value={vikep}
                  onChange={(e) => setVikep(e.target.value)}
                  placeholder="mis. RD. Antonius"
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                />
              </div>

              {isEdit ? (
                <label className="inline-flex items-center gap-2 text-sm text-fg">
                  <input
                    type="checkbox"
                    checked={active}
                    onChange={(e) => setActive(e.target.checked)}
                    className="size-4 accent-primary"
                  />
                  Kevikepan Aktif
                </label>
              ) : null}

              <div className="flex items-center justify-end gap-2 pt-2">
                <Link
                  to="/settings/gerejawi/kevikepan"
                  className={buttonClass("outline")}
                >
                  Batal
                </Link>
                <Button type="submit" disabled={pending}>
                  {pending ? "Menyimpan…" : "Simpan Kevikepan"}
                </Button>
              </div>
            </InductiveFormSection>
          </form>

          {isEdit ? (
            <InductiveDangerZone
              title="Hapus Kevikepan"
              description="Menghapus kevikepan tidak dapat dilakukan jika masih memiliki paroki."
              action={
                <Button
                  variant="danger"
                  disabled={deleteMut.isPending}
                  onClick={() => setConfirmDelete(true)}
                >
                  Hapus Kevikepan
                </Button>
              }
            />
          ) : null}

          <ConfirmDialog
            open={confirmDelete}
            title="Hapus kevikepan ini?"
            description="Pastikan tidak ada paroki yang masih terhubung dengan kevikepan ini."
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
