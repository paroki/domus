import { type FormEvent, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductivePageHeader } from "~/shared/ui/iui";
import { useCreateOrgPeriode, useUpdateOrgPeriode } from "../api/actions";
import { resolveOrgError } from "../api/errors";
import { useOrg, useOrgPeriodeDetail } from "../api/queries";
import { isParokiAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Periode Kepengurusan — Domus" }];
}

export default function OrgPeriodeFormPage() {
  const { orgId = "", periodeId = "" } = useParams();
  const isEdit = Boolean(periodeId);
  const navigate = useNavigate();
  const { data: me } = useMe();
  const parokiAdmin = isParokiAdmin(me?.roles);

  const orgQuery = useOrg(orgId);
  const org = orgQuery.data;
  const canManage = org?.myRole === "admin" || parokiAdmin;

  const detailQuery = useOrgPeriodeDetail(orgId, periodeId, isEdit);
  const createPeriode = useCreateOrgPeriode(orgId);
  const updatePeriode = useUpdateOrgPeriode(orgId);

  const [nama, setNama] = useState("");
  const [kode, setKode] = useState("");
  const [kodeEdited, setKodeEdited] = useState(false);
  const [tanggalMulai, setTanggalMulai] = useState("");
  const [tanggalSelesai, setTanggalSelesai] = useState("");
  const [status, setStatus] = useState<"open" | "closed">("open");
  const [isCurrent, setIsCurrent] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    const d = detailQuery.data;
    if (!d) return;
    setNama(d.nama ?? "");
    setKode(d.kode ?? "");
    setKodeEdited(true);
    setTanggalMulai(d.tanggalMulai ?? "");
    setTanggalSelesai(d.tanggalSelesai ?? "");
    setStatus(d.status === "closed" ? "closed" : "open");
    setIsCurrent(Boolean(d.isCurrent));
  }, [detailQuery.data]);

  function handleNamaChange(next: string) {
    setNama(next);
    if (!isEdit && !kodeEdited) {
      setKode(slugifyCode(next));
    }
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    if (!nama.trim() || !kode.trim() || !tanggalMulai || !tanggalSelesai) {
      setErrorMsg(
        "Nama, kode, tanggal mulai, dan tanggal selesai wajib diisi.",
      );
      return;
    }
    try {
      if (isEdit) {
        await updatePeriode.mutateAsync({
          periodeId,
          body: {
            kode: kode.trim(),
            nama: nama.trim(),
            tanggalMulai,
            tanggalSelesai,
            status,
            isCurrent,
          },
        });
      } else {
        await createPeriode.mutateAsync({
          kode: kode.trim(),
          nama: nama.trim(),
          tanggalMulai,
          tanggalSelesai,
          status,
          isCurrent,
        });
      }
      navigate(`/org/${orgId}`, { replace: true });
    } catch (err) {
      setErrorMsg(resolveOrgError(err));
    }
  }

  const pending = createPeriode.isPending || updatePeriode.isPending;

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Organisasi", to: "/org" },
          { label: org?.name ?? "Detail Organisasi", to: `/org/${orgId}` },
          { label: isEdit ? "Ubah Periode" : "Tambah Periode" },
        ]}
        title={
          isEdit ? "Ubah Periode Kepengurusan" : "Tambah Periode Kepengurusan"
        }
        subtitle="Atur rentang masa bakti kepengurusan dan status periode aktif organisasi."
      />

      {!orgQuery.isLoading && !canManage ? (
        <Card className="p-6 text-sm text-danger-fg">
          Hanya Admin Organisasi atau Admin Paroki yang dapat mengelola periode
          kepengurusan.
        </Card>
      ) : isEdit && detailQuery.isLoading ? (
        <Card className="p-6 text-sm text-muted">Memuat data periode…</Card>
      ) : (
        <Card className="p-6">
          {errorMsg && (
            <div
              role="alert"
              className="mb-4 rounded-lg border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
            >
              {errorMsg}
            </div>
          )}

          <form onSubmit={handleSubmit} className="grid gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
              Nama Periode
              <input
                type="text"
                required
                value={nama}
                onChange={(e) => handleNamaChange(e.target.value)}
                placeholder="Misal: Periode 2026–2029"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Kode Periode
              <input
                type="text"
                required
                value={kode}
                onChange={(e) => {
                  setKodeEdited(true);
                  setKode(e.target.value);
                }}
                placeholder="misal: 2026-2029"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 font-mono text-sm text-fg"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Status Periode
              <select
                value={status}
                onChange={(e) =>
                  setStatus(e.target.value === "closed" ? "closed" : "open")
                }
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="open">Terbuka</option>
                <option value="closed">Ditutup</option>
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Tanggal Mulai
              <input
                type="date"
                required
                value={tanggalMulai}
                onChange={(e) => setTanggalMulai(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Tanggal Selesai
              <input
                type="date"
                required
                value={tanggalSelesai}
                onChange={(e) => setTanggalSelesai(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>

            <label className="flex items-center gap-2.5 text-sm font-medium text-fg sm:col-span-2">
              <input
                type="checkbox"
                checked={isCurrent}
                onChange={(e) => setIsCurrent(e.target.checked)}
                className="size-4 accent-primary"
              />
              Jadikan sebagai Periode Aktif saat ini
            </label>

            <div className="flex justify-end gap-3 pt-2 sm:col-span-2">
              <Button
                variant="outline"
                onClick={() => navigate(`/org/${orgId}`)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={pending}>
                {isEdit ? "Simpan Perubahan" : "Simpan Periode"}
              </Button>
            </div>
          </form>
        </Card>
      )}
    </div>
  );
}
