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
  useCreateKeuskupan,
  useDeleteKeuskupan,
  useUpdateKeuskupan,
} from "../api/actions";
import { useKeuskupanDetail } from "../api/queries";
import { isSuperAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Formulir Keuskupan — Domus" }];
}

export default function KeuskupanFormPage() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  const detail = useKeuskupanDetail(id ?? "", isEdit && canManage);
  const createMut = useCreateKeuskupan();
  const updateMut = useUpdateKeuskupan();
  const deleteMut = useDeleteKeuskupan();

  const [kode, setKode] = useState("");
  const [kodeTouched, setKodeTouched] = useState(false);
  const [nama, setNama] = useState("");
  const [uskup, setUskup] = useState("");
  const [telepon, setTelepon] = useState("");
  const [email, setEmail] = useState("");
  const [website, setWebsite] = useState("");
  const [alamat, setAlamat] = useState("");
  const [active, setActive] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    if (!detail.data) return;
    setKode(detail.data.kode ?? "");
    setNama(detail.data.nama ?? "");
    setUskup(detail.data.uskup ?? "");
    setTelepon(detail.data.telepon ?? "");
    setEmail(detail.data.email ?? "");
    setWebsite(detail.data.website ?? "");
    setAlamat(detail.data.alamat ?? "");
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
            kode: kode.trim(),
            nama: nama.trim(),
            uskup: uskup.trim(),
            telepon: telepon.trim(),
            email: email.trim(),
            website: website.trim(),
            alamat: alamat.trim(),
            active,
          },
        });
      } else {
        await createMut.mutateAsync({
          kode: (kode || slugifyCode(nama)).trim(),
          nama: nama.trim(),
          uskup: uskup.trim() || undefined,
          telepon: telepon.trim() || undefined,
          email: email.trim() || undefined,
          website: website.trim() || undefined,
          alamat: alamat.trim() || undefined,
        });
      }
      navigate("/settings/gerejawi/keuskupan");
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal menyimpan.");
    }
  }

  async function handleDelete() {
    if (!id) return;
    setErrorMsg(null);
    try {
      await deleteMut.mutateAsync(id);
      navigate("/settings/gerejawi/keuskupan");
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
          { label: "Struktur Gerejawi", to: "/settings/gerejawi" },
          { label: "Keuskupan", to: "/settings/gerejawi/keuskupan" },
          { label: isEdit ? "Ubah Keuskupan" : "Tambah Keuskupan" },
        ]}
        title={isEdit ? "Ubah Keuskupan" : "Tambah Keuskupan"}
        subtitle={
          isEdit
            ? "Perbarui data keuskupan atau kelola status aktifnya."
            : "Tambahkan keuskupan baru ke dalam struktur gerejawi."
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola keuskupan.
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
            <InductiveFormSection title="Identitas Keuskupan">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="keuskupan-nama"
                    className="text-xs font-semibold text-muted"
                  >
                    Nama Keuskupan
                  </label>
                  <input
                    id="keuskupan-nama"
                    required
                    value={nama}
                    onChange={(e) => {
                      const next = e.target.value;
                      setNama(next);
                      if (!isEdit && !kodeTouched) {
                        setKode(slugifyCode(next));
                      }
                    }}
                    placeholder="mis. Keuskupan Agung Jakarta"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="keuskupan-kode"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode Keuskupan
                  </label>
                  <input
                    id="keuskupan-kode"
                    required
                    value={kode}
                    onChange={(e) => {
                      setKodeTouched(true);
                      setKode(e.target.value);
                    }}
                    placeholder="mis. kaj"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 font-mono text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
              </div>

              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor="keuskupan-uskup"
                  className="text-xs font-semibold text-muted"
                >
                  Nama Uskup (opsional)
                </label>
                <input
                  id="keuskupan-uskup"
                  value={uskup}
                  onChange={(e) => setUskup(e.target.value)}
                  placeholder="mis. Mgr. Ignatius Suharyo"
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                />
              </div>

              {isEdit ? (
                <label className="inline-flex min-h-11 items-center gap-2 text-sm text-fg">
                  <input
                    type="checkbox"
                    checked={active}
                    onChange={(e) => setActive(e.target.checked)}
                    className="size-4 accent-primary"
                  />
                  Keuskupan Aktif
                </label>
              ) : null}
            </InductiveFormSection>

            <InductiveFormSection
              title="Kontak &amp; Alamat"
              description="Informasi kontak resmi dan alamat kantor keuskupan (opsional)."
            >
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="keuskupan-telepon"
                    className="text-xs font-semibold text-muted"
                  >
                    Telepon
                  </label>
                  <input
                    id="keuskupan-telepon"
                    value={telepon}
                    onChange={(e) => setTelepon(e.target.value)}
                    placeholder="+6221..."
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="keuskupan-email"
                    className="text-xs font-semibold text-muted"
                  >
                    Email
                  </label>
                  <input
                    id="keuskupan-email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="sekretariat@keuskupan.or.id"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
              </div>

              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor="keuskupan-website"
                  className="text-xs font-semibold text-muted"
                >
                  Website
                </label>
                <input
                  id="keuskupan-website"
                  value={website}
                  onChange={(e) => setWebsite(e.target.value)}
                  placeholder="https://keuskupan.or.id"
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor="keuskupan-alamat"
                  className="text-xs font-semibold text-muted"
                >
                  Alamat
                </label>
                <textarea
                  id="keuskupan-alamat"
                  rows={2}
                  value={alamat}
                  onChange={(e) => setAlamat(e.target.value)}
                  placeholder="Alamat lengkap kantor keuskupan"
                  className="rounded-lg border border-line bg-canvas p-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                />
              </div>

              <div className="flex flex-col-reverse items-stretch justify-end gap-2 pt-2 sm:flex-row sm:items-center">
                <Link
                  to="/settings/gerejawi/keuskupan"
                  className={`${buttonClass("outline")} w-full sm:w-auto`}
                >
                  Batal
                </Link>
                <Button
                  type="submit"
                  disabled={pending}
                  className="w-full sm:w-auto"
                >
                  {pending ? "Menyimpan…" : "Simpan Keuskupan"}
                </Button>
              </div>
            </InductiveFormSection>
          </form>

          {isEdit ? (
            <InductiveDangerZone
              title="Hapus Keuskupan"
              description="Menghapus keuskupan akan menonaktifkannya. Tidak dapat dilakukan jika masih memiliki kevikepan atau paroki."
              action={
                <Button
                  variant="danger"
                  disabled={deleteMut.isPending}
                  onClick={() => setConfirmDelete(true)}
                >
                  Hapus Keuskupan
                </Button>
              }
            />
          ) : null}

          <ConfirmDialog
            open={confirmDelete}
            title="Hapus keuskupan ini?"
            description="Pastikan tidak ada kevikepan atau paroki yang masih terhubung."
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
