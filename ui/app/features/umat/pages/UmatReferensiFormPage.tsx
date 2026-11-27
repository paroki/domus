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
  useCreateUmatReferensi,
  useDeleteUmatReferensi,
  useUpdateUmatReferensi,
} from "../api/actions";
import { resolveUmatReferensiErrorMessage } from "../api/errors";
import { useUmatReferensiDetail } from "../api/queries";
import {
  canEditUmatReferensi,
  getUmatReferensiCategoryLabel,
  isSuperAdmin,
  isUmatAdmin,
  slugifyReferensiCode,
  UMAT_REFERENSI_CATEGORIES,
} from "../model";

export function meta() {
  return [{ title: "Formulir Kamus Referensi Umat — Domus" }];
}

export default function UmatReferensiFormPage() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const parokiId = me?.activeParokiId ?? "";
  const superAdmin = isSuperAdmin(me);
  const umatAdmin = isUmatAdmin(me);
  const canCreate = umatAdmin || superAdmin;

  const detail = useUmatReferensiDetail(
    parokiId,
    id ?? "",
    isEdit && Boolean(me),
  );
  const createMut = useCreateUmatReferensi();
  const updateMut = useUpdateUmatReferensi(parokiId);
  const deleteMut = useDeleteUmatReferensi(parokiId);

  const [category, setCategory] = useState(
    searchParams.get("category") ?? UMAT_REFERENSI_CATEGORIES[0].slug,
  );
  const [label, setLabel] = useState("");
  const [code, setCode] = useState("");
  const [codeTouched, setCodeTouched] = useState(false);
  const [sortOrder, setSortOrder] = useState(0);
  const [isGlobal, setIsGlobal] = useState(superAdmin && !umatAdmin);
  const [active, setActive] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    if (!detail.data) return;
    setCategory(detail.data.category ?? UMAT_REFERENSI_CATEGORIES[0].slug);
    setLabel(detail.data.label ?? "");
    setCode(detail.data.code ?? "");
    setSortOrder(detail.data.sortOrder ?? 0);
    setIsGlobal(!detail.data.parokiId);
    setActive(detail.data.active ?? true);
  }, [detail.data]);

  const canModify = isEdit
    ? detail.data
      ? canEditUmatReferensi(me, detail.data.parokiId)
      : canCreate
    : canCreate;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      if (isEdit && id) {
        await updateMut.mutateAsync({
          id,
          body: {
            category,
            code: code.trim(),
            label: label.trim(),
            sortOrder,
            active,
          },
        });
      } else {
        await createMut.mutateAsync({
          category,
          code: (code || slugifyReferensiCode(label)).trim(),
          label: label.trim(),
          sortOrder,
          global: superAdmin ? isGlobal : false,
        });
      }
      navigate(`/umat/referensi?category=${encodeURIComponent(category)}`);
    } catch (err) {
      setErrorMsg(resolveUmatReferensiErrorMessage(err));
    }
  }

  async function handleDelete() {
    if (!id) return;
    setErrorMsg(null);
    try {
      await deleteMut.mutateAsync(id);
      navigate(`/umat/referensi?category=${encodeURIComponent(category)}`);
    } catch (err) {
      setConfirmDelete(false);
      setErrorMsg(resolveUmatReferensiErrorMessage(err));
    }
  }

  const pending = createMut.isPending || updateMut.isPending;

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Data Umat", to: "/umat" },
          { label: "Kamus Referensi Umat", to: "/umat/referensi" },
          {
            label: getUmatReferensiCategoryLabel(category),
            to: `/umat/referensi?category=${encodeURIComponent(category)}`,
          },
          { label: isEdit ? "Ubah Item" : "Tambah Item" },
        ]}
        title={isEdit ? "Ubah Item Referensi" : "Tambah Item Referensi"}
        subtitle={
          isEdit
            ? "Perbarui label, kode, urutan, atau status aktif item referensi."
            : "Tambahkan nilai referensi baru pada kategori Data Umat."
        }
      />

      {!meLoading && !canModify ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Anda tidak memiliki izin untuk mengubah item referensi ini.
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

      {canModify && isEdit && detail.isLoading ? (
        <p className="text-sm text-muted">Memuat data…</p>
      ) : canModify ? (
        <>
          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            <InductiveFormSection title="Detail Item Referensi">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="referensi-category"
                    className="text-xs font-semibold text-muted"
                  >
                    Kategori
                  </label>
                  <select
                    id="referensi-category"
                    value={category}
                    onChange={(e) => setCategory(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  >
                    {UMAT_REFERENSI_CATEGORIES.map((c) => (
                      <option key={c.slug} value={c.slug}>
                        {c.label}
                      </option>
                    ))}
                  </select>
                </div>

                {!isEdit && superAdmin ? (
                  <div className="flex items-end pb-2">
                    <label
                      htmlFor="referensi-global"
                      className="inline-flex h-11 items-center gap-2 text-base text-fg sm:text-sm"
                    >
                      <input
                        id="referensi-global"
                        type="checkbox"
                        checked={isGlobal}
                        onChange={(e) => setIsGlobal(e.target.checked)}
                        className="size-4 accent-primary"
                      />
                      Item Global (semua paroki)
                    </label>
                  </div>
                ) : null}
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="referensi-label"
                    className="text-xs font-semibold text-muted"
                  >
                    Label Tampilan
                  </label>
                  <input
                    id="referensi-label"
                    required
                    value={label}
                    onChange={(e) => {
                      const next = e.target.value;
                      setLabel(next);
                      if (!isEdit && !codeTouched) {
                        setCode(slugifyReferensiCode(next));
                      }
                    }}
                    placeholder="mis. Katolik"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="referensi-code"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode
                  </label>
                  <input
                    id="referensi-code"
                    required
                    value={code}
                    onChange={(e) => {
                      setCodeTouched(true);
                      setCode(e.target.value);
                    }}
                    placeholder="mis. katolik"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 font-mono text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="referensi-sort"
                    className="text-xs font-semibold text-muted"
                  >
                    Urutan
                  </label>
                  <input
                    id="referensi-sort"
                    type="number"
                    value={sortOrder}
                    onChange={(e) => setSortOrder(Number(e.target.value) || 0)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>

                {isEdit ? (
                  <div className="flex items-end pb-2">
                    <label className="inline-flex h-11 items-center gap-2 text-base text-fg sm:text-sm">
                      <input
                        type="checkbox"
                        checked={active}
                        onChange={(e) => setActive(e.target.checked)}
                        className="size-4 accent-primary"
                      />
                      Item Aktif
                    </label>
                  </div>
                ) : null}
              </div>

              <div className="flex items-center justify-end gap-2 pt-2">
                <Link
                  to={`/umat/referensi?category=${encodeURIComponent(category)}`}
                  className={buttonClass("outline")}
                >
                  Batal
                </Link>
                <Button type="submit" disabled={pending}>
                  {pending ? "Menyimpan…" : "Simpan Item"}
                </Button>
              </div>
            </InductiveFormSection>
          </form>

          {isEdit ? (
            <InductiveDangerZone
              title="Hapus Item Referensi"
              description="Menghapus item akan menonaktifkannya dari daftar pilihan jika tidak sedang digunakan."
              action={
                <Button
                  variant="danger"
                  disabled={deleteMut.isPending}
                  onClick={() => setConfirmDelete(true)}
                >
                  Hapus Item
                </Button>
              }
            />
          ) : null}

          <ConfirmDialog
            open={confirmDelete}
            title="Hapus item referensi ini?"
            description="Item yang dihapus tidak lagi muncul dalam pilihan aktif."
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
