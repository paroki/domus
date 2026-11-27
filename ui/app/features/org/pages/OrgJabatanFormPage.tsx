import { type FormEvent, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductivePageHeader } from "~/shared/ui/iui";
import { useCreateOrgJabatan, useUpdateOrgJabatan } from "../api/actions";
import { resolveOrgError } from "../api/errors";
import {
  useOrg,
  useOrgJabatanDetail,
  useOrgJabatanList,
  useOrgMembers,
  useOrgPeriodeDetail,
  useOrgPositionLookups,
  useOrgUmatLookup,
} from "../api/queries";
import {
  buildJabatanTree,
  DEFAULT_POSITION_SUGGESTIONS,
  flattenJabatanTreeForSelect,
  isParokiAdmin,
} from "../model";

type OfficerSource = "member" | "umat" | "manual";

export function meta() {
  return [{ title: "Jabatan Organisasi — Domus" }];
}

export default function OrgJabatanFormPage() {
  const { orgId = "", periodeId = "", jabatanId = "" } = useParams();
  const [searchParams] = useSearchParams();
  const isEdit = Boolean(jabatanId);
  const navigate = useNavigate();
  const { data: me } = useMe();
  const parokiAdmin = isParokiAdmin(me?.roles);

  const orgQuery = useOrg(orgId);
  const org = orgQuery.data;
  const canManage = org?.myRole === "admin" || parokiAdmin;

  const periodeQuery = useOrgPeriodeDetail(
    orgId,
    periodeId,
    Boolean(periodeId),
  );
  const detailQuery = useOrgJabatanDetail(orgId, jabatanId, isEdit);
  const jabatanListQuery = useOrgJabatanList(
    orgId,
    { filter: { periodeId } },
    Boolean(orgId) && Boolean(periodeId),
  );
  const membersQuery = useOrgMembers(
    orgId,
    { filter: { status: "active" } },
    Boolean(org) && canManage,
  );
  const lookupsQuery = useOrgPositionLookups(canManage);

  const [umatSearch, setUmatSearch] = useState("");
  const umatLookupQuery = useOrgUmatLookup(
    { q: umatSearch.trim() || undefined, limit: 50 },
    canManage,
  );

  const createJabatan = useCreateOrgJabatan(orgId);
  const updateJabatan = useUpdateOrgJabatan(orgId);

  const [jabatan, setJabatan] = useState("");
  const [parentId, setParentId] = useState(
    () => searchParams.get("parentId") ?? "",
  );
  const [officerSource, setOfficerSource] = useState<OfficerSource>("manual");
  const [selectedUserId, setSelectedUserId] = useState("");
  const [selectedUmatId, setSelectedUmatId] = useState("");
  const [manualNama, setManualNama] = useState("");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [urutan, setUrutan] = useState("0");
  const [active, setActive] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    const d = detailQuery.data;
    if (!d) return;
    setJabatan(d.jabatan ?? "");
    setParentId(d.parentId ?? "");
    if (d.umatId) {
      setOfficerSource("umat");
      setSelectedUmatId(d.umatId);
      setSelectedUserId(d.userId ?? "");
      setManualNama(d.nama ?? "");
    } else if (d.userId) {
      setOfficerSource("member");
      setSelectedUserId(d.userId);
      setSelectedUmatId("");
      setManualNama(d.nama ?? "");
    } else {
      setOfficerSource("manual");
      setSelectedUserId("");
      setSelectedUmatId("");
      setManualNama(d.nama ?? "");
    }
    setStartDate(d.startDate ?? "");
    setEndDate(d.endDate ?? "");
    setUrutan(String(d.urutan ?? 0));
    setActive(Boolean(d.active));
  }, [detailQuery.data]);

  const positionSuggestions = useMemo(
    () =>
      Array.from(
        new Set([
          ...(lookupsQuery.data ?? []).map((l) => l.label),
          ...DEFAULT_POSITION_SUGGESTIONS,
        ]),
      ),
    [lookupsQuery.data],
  );

  const parentOptions = useMemo(() => {
    const items = jabatanListQuery.data ?? [];
    const tree = buildJabatanTree(items);
    return flattenJabatanTreeForSelect(tree, isEdit ? jabatanId : undefined);
  }, [jabatanListQuery.data, isEdit, jabatanId]);

  const activeMembers = useMemo(
    () => membersQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [],
    [membersQuery.data],
  );
  const umatItems = umatLookupQuery.data ?? [];

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    if (!jabatan.trim()) {
      setErrorMsg("Nama jabatan wajib diisi.");
      return;
    }

    const parsedUrutan = Number.parseInt(urutan.trim() || "0", 10);
    const safeUrutan = Number.isNaN(parsedUrutan) ? 0 : parsedUrutan;

    let userIdVal: string | undefined;
    let umatIdVal: string | undefined;
    let namaVal: string | undefined;

    if (officerSource === "member" && selectedUserId) {
      userIdVal = selectedUserId;
    } else if (officerSource === "umat" && selectedUmatId) {
      umatIdVal = selectedUmatId;
    } else if (officerSource === "manual" && manualNama.trim()) {
      namaVal = manualNama.trim();
    }

    try {
      if (isEdit) {
        await updateJabatan.mutateAsync({
          jabatanId,
          body: {
            jabatan: jabatan.trim(),
            parentId: parentId || undefined,
            clearParentId: !parentId,
            userId: userIdVal,
            clearUserId: !userIdVal,
            umatId: umatIdVal,
            clearUmatId: !umatIdVal,
            nama: namaVal,
            clearNama: !userIdVal && !umatIdVal && !namaVal,
            startDate: startDate || undefined,
            clearStartDate: !startDate,
            endDate: endDate || undefined,
            clearEndDate: !endDate,
            urutan: safeUrutan,
            active,
          },
        });
      } else {
        await createJabatan.mutateAsync({
          periodeId,
          parentId: parentId || undefined,
          jabatan: jabatan.trim(),
          userId: userIdVal,
          umatId: umatIdVal,
          nama: namaVal,
          startDate: startDate || undefined,
          endDate: endDate || undefined,
          urutan: safeUrutan,
          active,
        });
      }
      navigate(`/org/${orgId}`, { replace: true });
    } catch (err) {
      setErrorMsg(resolveOrgError(err));
    }
  }

  const pending = createJabatan.isPending || updateJabatan.isPending;

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Organisasi", to: "/org" },
          { label: org?.name ?? "Detail Organisasi", to: `/org/${orgId}` },
          { label: isEdit ? "Ubah Jabatan" : "Tambah Jabatan" },
        ]}
        title={isEdit ? "Ubah Jabatan Organisasi" : "Tambah Jabatan Organisasi"}
        subtitle={
          periodeQuery.data
            ? `Periode: ${periodeQuery.data.nama} (${periodeQuery.data.tanggalMulai} – ${periodeQuery.data.tanggalSelesai})`
            : "Kelola posisi struktural dan pengurus pada periode kepengurusan."
        }
      />

      {!orgQuery.isLoading && !canManage ? (
        <Card className="p-4 text-sm text-danger-fg sm:p-6">
          Hanya Admin Organisasi atau Admin Paroki yang dapat mengelola struktur
          jabatan.
        </Card>
      ) : isEdit && detailQuery.isLoading ? (
        <Card className="p-4 text-sm text-muted sm:p-6">
          Memuat data jabatan…
        </Card>
      ) : (
        <Card className="p-4 sm:p-6">
          {errorMsg && (
            <div
              role="alert"
              className="mb-4 rounded-lg border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
            >
              {errorMsg}
            </div>
          )}

          <datalist id="org-jabatan-suggestions">
            {positionSuggestions.map((pos) => (
              <option key={pos} value={pos} />
            ))}
          </datalist>

          <form onSubmit={handleSubmit} className="grid gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Nama Jabatan
              <input
                type="text"
                required
                list="org-jabatan-suggestions"
                value={jabatan}
                onChange={(e) => setJabatan(e.target.value)}
                placeholder="Misal: Ketua, Sekretaris, Koordinator"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Atasan / Induk Jabatan
              <select
                value={parentId}
                onChange={(e) => setParentId(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">Posisi Puncak (Tanpa Induk)</option>
                {parentOptions.map((opt) => (
                  <option key={opt.id} value={opt.id}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </label>

            <div className="flex flex-col gap-3 rounded-lg border border-line bg-canvas p-4 sm:col-span-2">
              <span className="text-sm font-semibold text-fg">
                Sumber Data Pengurus
              </span>
              <div className="flex flex-col gap-1 sm:flex-row sm:flex-wrap sm:gap-3">
                {(
                  [
                    {
                      value: "member",
                      label: "Pilih dari Anggota Organisasi",
                    },
                    { value: "umat", label: "Pilih dari Data Umat Paroki" },
                    {
                      value: "manual",
                      label: "Ketik Nama Manual / Kosongkan",
                    },
                  ] as const
                ).map((opt) => (
                  <label
                    key={opt.value}
                    className="flex min-h-11 cursor-pointer items-center gap-2 text-sm text-fg"
                  >
                    <input
                      type="radio"
                      name="officerSource"
                      checked={officerSource === opt.value}
                      onChange={() => setOfficerSource(opt.value)}
                      className="size-4 accent-primary"
                    />
                    {opt.label}
                  </label>
                ))}
              </div>

              {officerSource === "member" && (
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Anggota Organisasi
                  <select
                    value={selectedUserId}
                    onChange={(e) => setSelectedUserId(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-surface px-3.5 text-sm text-fg"
                  >
                    <option value="">— Pilih anggota —</option>
                    {activeMembers.map((m) => (
                      <option key={m.userId} value={m.userId}>
                        {m.name} ({m.email})
                      </option>
                    ))}
                  </select>
                </label>
              )}

              {officerSource === "umat" && (
                <div className="grid gap-3 sm:grid-cols-2">
                  <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                    Cari Nama Umat
                    <input
                      type="search"
                      value={umatSearch}
                      onChange={(e) => setUmatSearch(e.target.value)}
                      placeholder="Ketik nama atau nama baptis…"
                      className="h-11 rounded-lg border border-line bg-surface px-3.5 text-sm text-fg"
                    />
                  </label>
                  <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                    Pilih Data Umat
                    <select
                      value={selectedUmatId}
                      onChange={(e) => setSelectedUmatId(e.target.value)}
                      className="h-11 rounded-lg border border-line bg-surface px-3.5 text-sm text-fg"
                    >
                      <option value="">— Pilih umat —</option>
                      {umatItems.map((u) => (
                        <option key={u.id} value={u.id}>
                          {u.nama}
                          {u.namaBaptis ? ` (${u.namaBaptis})` : ""}
                        </option>
                      ))}
                    </select>
                  </label>
                </div>
              )}

              {officerSource === "manual" && (
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Nama Pengurus (opsional — kosongkan jika jabatan lowong)
                  <input
                    type="text"
                    value={manualNama}
                    onChange={(e) => setManualNama(e.target.value)}
                    placeholder="Misal: Bpk. Antonius (atau kosongkan)"
                    className="h-11 rounded-lg border border-line bg-surface px-3.5 text-sm text-fg"
                  />
                </label>
              )}
            </div>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Tanggal Mulai Menjabat (opsional)
              <input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Tanggal Selesai Menjabat (opsional)
              <input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Urutan Tampil
              <input
                type="number"
                value={urutan}
                onChange={(e) => setUrutan(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>

            <label className="flex min-h-11 cursor-pointer items-center gap-2.5 self-end text-sm font-medium text-fg">
              <input
                type="checkbox"
                checked={active}
                onChange={(e) => setActive(e.target.checked)}
                className="size-4 accent-primary"
              />
              Jabatan Aktif
            </label>

            <div className="flex flex-col-reverse gap-2 pt-2 sm:col-span-2 sm:flex-row sm:justify-end sm:gap-3">
              <Button
                variant="outline"
                onClick={() => navigate(`/org/${orgId}`)}
                className="w-full sm:w-auto"
              >
                Batal
              </Button>
              <Button
                type="submit"
                disabled={pending}
                className="w-full sm:w-auto"
              >
                {isEdit ? "Simpan Perubahan" : "Simpan Jabatan"}
              </Button>
            </div>
          </form>
        </Card>
      )}
    </div>
  );
}
