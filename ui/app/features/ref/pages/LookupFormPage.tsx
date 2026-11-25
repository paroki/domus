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
  useCreateLookup,
  useDeleteLookup,
  useUpdateLookup,
} from "../api/actions";
import { useLookupDetail } from "../api/queries";
import {
  getLookupCategoryLabel,
  hasAnyRefAdminRole,
  isParokiAdmin,
  isSuperAdmin,
  LOOKUP_CATEGORY_META,
  slugifyCode,
} from "../model";

export function meta() {
  return [{ title: "Formulir Kamus Referensi — Domus" }];
}

export default function LookupFormPage() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = hasAnyRefAdminRole(me?.roles);
  const superAdmin = isSuperAdmin(me?.roles);
  const parokiAdmin = isParokiAdmin(me?.roles);

  const detail = useLookupDetail(id ?? "", isEdit && canManage);
  const createMut = useCreateLookup();
  const updateMut = useUpdateLookup();
  const deleteMut = useDeleteLookup();

  const [category, setCategory] = useState(
    searchParams.get("category") ?? LOOKUP_CATEGORY_META[0].slug,
  );
  const [label, setLabel] = useState("");
  const [code, setCode] = useState("");
  const [codeTouched, setCodeTouched] = useState(false);
  const [sortOrder, setSortOrder] = useState(0);
  const [scope, setScope] = useState<"global" | "paroki">(
    superAdmin && !parokiAdmin ? "global" : "paroki",
  );
  const [active, setActive] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    if (!detail.data) return;
    setCategory(detail.data.category ?? LOOKUP_CATEGORY_META[0].slug);
    setLabel(detail.data.label ?? "");
    setCode(detail.data.code ?? "");
    setSortOrder(detail.data.sortOrder ?? 0);
    setScope(detail.data.parokiId ? "paroki" : "global");
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
          code: (code || slugifyCode(label)).trim(),
          label: label.trim(),
          sortOrder,
          global: scope === "global",
        });
      }
      navigate(`/settings/lookups?category=${encodeURIComponent(category)}`);
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal menyimpan.");
    }
  }

  async function handleDelete() {
    if (!id) return;
    setErrorMsg(null);
    try {
      await deleteMut.mutateAsync(id);
      navigate(`/settings/lookups?category=${encodeURIComponent(category)}`);
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
          { label: "Kamus Referensi", to: "/settings/lookups" },
          {
            label: getLookupCategoryLabel(category),
            to: `/settings/lookups?category=${encodeURIComponent(category)}`,
          },
          { label: isEdit ? "Ubah Item" : "Tambah Item" },
        ]}
        title={isEdit ? "Ubah Item Referensi" : "Tambah Item Referensi"}
        subtitle={
          isEdit
            ? "Perbarui label, kode, urutan, atau status aktif item referensi."
            : "Tambahkan nilai referensi baru pada kategori yang dipilih."
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Admin Paroki atau Super Admin yang dapat mengelola kamus
          referensi.
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
            <InductiveFormSection title="Detail Item Referensi">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="lookup-category"
                    className="text-xs font-semibold text-muted"
                  >
                    Kategori
                  </label>
                  <select
                    id="lookup-category"
                    value={category}
                    onChange={(e) => setCategory(e.target.value)}
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  >
                    {LOOKUP_CATEGORY_META.map((c) => (
                      <option key={c.slug} value={c.slug}>
                        {c.label}
                      </option>
                    ))}
                  </select>
                </div>

                {!isEdit ? (
                  <div className="flex flex-col gap-1.5">
                    <label
                      htmlFor="lookup-scope"
                      className="text-xs font-semibold text-muted"
                    >
                      Cakupan
                    </label>
                    <select
                      id="lookup-scope"
                      value={scope}
                      onChange={(e) =>
                        setScope(e.target.value as "global" | "paroki")
                      }
                      className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                    >
                      {parokiAdmin ? (
                        <option value="paroki">Paroki Aktif</option>
                      ) : null}
                      {superAdmin ? (
                        <option value="global">Global</option>
                      ) : null}
                    </select>
                  </div>
                ) : null}
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="lookup-label"
                    className="text-xs font-semibold text-muted"
                  >
                    Label Tampilan
                  </label>
                  <input
                    id="lookup-label"
                    required
                    value={label}
                    onChange={(e) => {
                      const next = e.target.value;
                      setLabel(next);
                      if (!isEdit && !codeTouched) {
                        setCode(slugifyCode(next));
                      }
                    }}
                    placeholder="mis. Paroki Mandiri"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="lookup-code"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode
                  </label>
                  <input
                    id="lookup-code"
                    required
                    value={code}
                    onChange={(e) => {
                      setCodeTouched(true);
                      setCode(e.target.value);
                    }}
                    placeholder="mis. paroki-mandiri"
                    className="h-10 rounded-lg border border-line bg-canvas px-3 font-mono text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="lookup-sort"
                    className="text-xs font-semibold text-muted"
                  >
                    Urutan
                  </label>
                  <input
                    id="lookup-sort"
                    type="number"
                    value={sortOrder}
                    onChange={(e) => setSortOrder(Number(e.target.value) || 0)}
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                  />
                </div>

                {isEdit ? (
                  <div className="flex items-end pb-2">
                    <label className="inline-flex items-center gap-2 text-sm text-fg">
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
                  to={`/settings/lookups?category=${encodeURIComponent(category)}`}
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
              description="Menghapus item akan menonaktifkannya dari daftar pilihan."
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
