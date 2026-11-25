import { Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useNavigate, useParams, useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import {
  InductiveDangerZone,
  InductiveFormSection,
  InductivePageHeader,
} from "~/shared/ui/iui";
import {
  useAddParokiMember,
  useCreateParoki,
  useDeleteParoki,
  useRemoveParokiMember,
  useUpdateParoki,
  useUpdateParokiMember,
} from "../api/actions";
import {
  useKeuskupanSearch,
  useParokiDetail,
  useParokiMembersSearch,
} from "../api/queries";
import { isSuperAdmin } from "../model";

export function meta() {
  return [{ title: "Formulir Paroki — Domus" }];
}

export default function ParokiFormPage() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  const keuskupanQuery = useKeuskupanSearch({}, canManage);
  const keuskupanList =
    keuskupanQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const detail = useParokiDetail(id ?? "", isEdit && canManage);
  const membersQuery = useParokiMembersSearch(
    id ?? "",
    {},
    isEdit && canManage,
  );
  const members = membersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const createMut = useCreateParoki();
  const updateMut = useUpdateParoki();
  const deleteMut = useDeleteParoki();
  const addMemberMut = useAddParokiMember();
  const updateMemberMut = useUpdateParokiMember();
  const removeMemberMut = useRemoveParokiMember();

  const [keuskupanId, setKeuskupanId] = useState(
    searchParams.get("keuskupanId") ?? "",
  );
  const [kode, setKode] = useState("");
  const [nama, setNama] = useState("");
  const [pelindung, setPelindung] = useState("");
  const [pastor, setPastor] = useState("");
  const [active, setActive] = useState(true);

  const [adminEmail, setAdminEmail] = useState("");
  const [adminName, setAdminName] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [statusMsg, setStatusMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    if (!detail.data) return;
    setKeuskupanId(detail.data.keuskupanId ?? "");
    setKode(detail.data.kode ?? "");
    setNama(detail.data.nama ?? "");
    setPelindung(detail.data.pelindung ?? "");
    setPastor(detail.data.pastor ?? "");
    setActive(detail.data.active ?? true);
  }, [detail.data]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    setStatusMsg(null);
    try {
      if (isEdit && id) {
        await updateMut.mutateAsync({
          id,
          body: {
            keuskupanId,
            kode: kode.trim(),
            nama: nama.trim(),
            pelindung: pelindung.trim(),
            pastor: pastor.trim(),
            active,
          },
        });
        setStatusMsg("Data paroki berhasil diperbarui.");
      } else {
        const created = await createMut.mutateAsync({
          keuskupanId,
          kode: kode.trim(),
          nama: nama.trim(),
          pelindung: pelindung.trim() || undefined,
          pastor: pastor.trim() || undefined,
        });
        navigate(`/settings/gerejawi/paroki/${created.id}`);
      }
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal menyimpan.");
    }
  }

  async function handleAddAdmin(e: React.FormEvent) {
    e.preventDefault();
    if (!id) return;
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
    if (!id) return;
    setErrorMsg(null);
    try {
      await deleteMut.mutateAsync(id);
      navigate("/settings/gerejawi/paroki");
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
          { label: "Paroki", to: "/settings/gerejawi/paroki" },
          { label: isEdit ? "Ubah Paroki" : "Tambah Paroki" },
        ]}
        title={isEdit ? "Ubah Paroki" : "Tambah Paroki"}
        subtitle={
          isEdit
            ? "Perbarui identitas paroki dan kelola daftar admin paroki."
            : "Tambahkan tenant paroki baru di bawah keuskupan."
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola paroki.
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

      {canManage ? (
        <>
          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            <InductiveFormSection title="Data Paroki">
              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor="paroki-keuskupan"
                  className="text-xs font-semibold text-muted"
                >
                  Keuskupan
                </label>
                <select
                  id="paroki-keuskupan"
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
                    htmlFor="paroki-kode"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode Paroki
                  </label>
                  <input
                    id="paroki-kode"
                    required
                    value={kode}
                    onChange={(e) => setKode(e.target.value)}
                    placeholder="mis. katedral-jkt"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 font-mono text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="paroki-nama"
                    className="text-xs font-semibold text-muted"
                  >
                    Nama Paroki
                  </label>
                  <input
                    id="paroki-nama"
                    required
                    value={nama}
                    onChange={(e) => setNama(e.target.value)}
                    placeholder="mis. Paroki Katedral Jakarta"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="paroki-pelindung"
                    className="text-xs font-semibold text-muted"
                  >
                    Pelindung (opsional)
                  </label>
                  <input
                    id="paroki-pelindung"
                    value={pelindung}
                    onChange={(e) => setPelindung(e.target.value)}
                    placeholder="mis. Santa Perawan Maria"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="paroki-pastor"
                    className="text-xs font-semibold text-muted"
                  >
                    Pastor Paroki (opsional)
                  </label>
                  <input
                    id="paroki-pastor"
                    value={pastor}
                    onChange={(e) => setPastor(e.target.value)}
                    placeholder="mis. RD. Yohanes"
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
                  Paroki Aktif
                </label>
              ) : null}

              <div className="flex items-center justify-end gap-2 pt-2">
                <Link
                  to="/settings/gerejawi/paroki"
                  className={buttonClass("outline")}
                >
                  Kembali
                </Link>
                <Button type="submit" disabled={pending}>
                  {pending ? "Menyimpan…" : "Simpan Paroki"}
                </Button>
              </div>
            </InductiveFormSection>
          </form>

          {isEdit && id ? (
            <>
              <InductiveFormSection
                title="Admin & Anggota Paroki"
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
                  <p className="text-sm text-muted">
                    Belum ada anggota terdaftar.
                  </p>
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
                          <Badge
                            tone={m.role === "admin" ? "success" : "neutral"}
                          >
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
                            <Trash2
                              aria-hidden
                              className="size-4 text-danger-fg"
                            />
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </InductiveFormSection>

              <InductiveDangerZone
                title="Hapus Paroki"
                description="Menghapus paroki akan menonaktifkannya. Tidak dapat dilakukan jika masih memiliki data wilayah, organisasi, atau kegiatan."
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
        </>
      ) : null}
    </div>
  );
}
