import { type FormEvent, useEffect, useState } from "react";
import { Link, useNavigate, useParams, useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useUpdateSakramen, useUpsertSakramen } from "../api/actions";
import { resolveUmatErrorMessage } from "../api/errors";
import {
  useUmatDetail,
  useUmatKeuskupanList,
  useUmatParokiList,
} from "../api/queries";
import { UmatForbiddenCard } from "../components/UmatForbiddenCard";
import {
  isUmatAdmin,
  optionalStringForPatch,
  optionalUuidForPatch,
  SAKRAMEN_JENIS_OPTIONS,
  type SakramenJenis,
  validateYYYYMMDD,
} from "../model";

export function meta() {
  return [{ title: "Catatan Sakramen · Data Umat · Domus" }];
}

export default function UmatSakramenFormPage() {
  const { umatId = "", sakramenId } = useParams();
  const isEdit = Boolean(sakramenId);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { data: me } = useMe();

  const activeParokiId = me?.activeParokiId ?? "";
  const allowed = isUmatAdmin(me);

  const umatQuery = useUmatDetail(
    activeParokiId,
    umatId,
    Boolean(me && allowed && activeParokiId && umatId),
  );
  const keuskupanQuery = useUmatKeuskupanList(
    Boolean(me && allowed && activeParokiId),
  );

  const existingSakramen = isEdit
    ? (umatQuery.data?.sakramen ?? []).find((s) => s.id === sakramenId)
    : undefined;

  const initialJenis = (searchParams.get("jenis") as SakramenJenis) || "baptis";
  const [jenis, setJenis] = useState<SakramenJenis>(initialJenis);
  const [tanggal, setTanggal] = useState("");
  const [keuskupanId, setKeuskupanId] = useState("");
  const [parokiSakramenId, setParokiSakramenId] = useState("");
  const [keuskupanNama, setKeuskupanNama] = useState("");
  const [parokiNama, setParokiNama] = useState("");
  const [buku, setBuku] = useState("");
  const [nomor, setNomor] = useState("");
  const [catatan, setCatatan] = useState("");
  const [initialized, setInitialized] = useState(!isEdit);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (isEdit && existingSakramen && !initialized) {
      setJenis((existingSakramen.jenis as SakramenJenis) || "baptis");
      setTanggal(existingSakramen.tanggal ?? "");
      setKeuskupanId(existingSakramen.keuskupanId ?? "");
      setParokiSakramenId(existingSakramen.parokiSakramenId ?? "");
      setKeuskupanNama(existingSakramen.keuskupanNama ?? "");
      setParokiNama(existingSakramen.parokiNama ?? "");
      setBuku(existingSakramen.buku ?? "");
      setNomor(existingSakramen.nomor ?? "");
      setCatatan(existingSakramen.catatan ?? "");
      setInitialized(true);
    }
  }, [isEdit, existingSakramen, initialized]);

  const parokiQuery = useUmatParokiList(
    keuskupanId || undefined,
    Boolean(me && allowed && activeParokiId),
  );

  const upsertSakramen = useUpsertSakramen(activeParokiId, umatId);
  const updateSakramen = useUpdateSakramen(activeParokiId, umatId);

  if (me && (!allowed || !activeParokiId)) {
    return <UmatForbiddenCard />;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const dateErr = validateYYYYMMDD(tanggal, "Tanggal sakramen");
    if (dateErr) {
      setErrorMsg(dateErr);
      return;
    }
    setErrorMsg(null);
    try {
      if (isEdit && sakramenId) {
        await updateSakramen.mutateAsync({
          sakramenId,
          body: {
            jenis,
            tanggal: optionalStringForPatch(tanggal),
            keuskupanId: optionalUuidForPatch(keuskupanId),
            parokiSakramenId: optionalUuidForPatch(parokiSakramenId),
            keuskupanNama: optionalStringForPatch(keuskupanNama),
            parokiNama: optionalStringForPatch(parokiNama),
            buku: optionalStringForPatch(buku),
            nomor: optionalStringForPatch(nomor),
            catatan: optionalStringForPatch(catatan),
          },
        });
      } else {
        await upsertSakramen.mutateAsync({
          jenis,
          ...(tanggal.trim() ? { tanggal: tanggal.trim() } : {}),
          ...(keuskupanId ? { keuskupanId } : {}),
          ...(parokiSakramenId ? { parokiSakramenId } : {}),
          ...(keuskupanNama.trim()
            ? { keuskupanNama: keuskupanNama.trim() }
            : {}),
          ...(parokiNama.trim() ? { parokiNama: parokiNama.trim() } : {}),
          ...(buku.trim() ? { buku: buku.trim() } : {}),
          ...(nomor.trim() ? { nomor: nomor.trim() } : {}),
          ...(catatan.trim() ? { catatan: catatan.trim() } : {}),
        });
      }
      navigate(`/umat/${umatId}`);
    } catch (err) {
      setErrorMsg(resolveUmatErrorMessage(err));
    }
  }

  const umatName = umatQuery.data?.nama ?? "Profil Umat";
  const isPending = upsertSakramen.isPending || updateSakramen.isPending;

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Data Umat", to: "/umat" },
          { label: umatName, to: `/umat/${umatId}` },
          { label: isEdit ? "Ubah Sakramen" : "Catat Sakramen" },
        ]}
        title={isEdit ? "Ubah Catatan Sakramen" : "Catat Sakramen"}
        subtitle={`Riwayat penerimaan sakramen untuk ${umatName}.`}
      />

      {errorMsg ? (
        <div
          role="alert"
          className="rounded-lg border border-danger/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
        >
          {errorMsg}
        </div>
      ) : null}

      {isEdit && umatQuery.isLoading ? (
        <Card className="p-4 text-sm text-muted sm:p-6">
          Memuat catatan sakramen…
        </Card>
      ) : isEdit && !existingSakramen ? (
        <Card className="flex flex-col items-start gap-3 p-4 sm:p-6">
          <p className="text-sm text-danger-fg">
            Catatan sakramen tidak ditemukan.
          </p>
          <Link to={`/umat/${umatId}`} className="no-underline">
            <Button variant="outline">Kembali ke profil umat</Button>
          </Link>
        </Card>
      ) : (
        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          <InductiveFormSection
            title="Detail Sakramen & Buku Register"
            description="Pilih jenis sakramen, tanggal penerimaan, dan nomor register."
          >
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Jenis Sakramen *
                <select
                  value={jenis}
                  onChange={(e) => setJenis(e.target.value as SakramenJenis)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  {SAKRAMEN_JENIS_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Tanggal Sakramen
                <input
                  type="date"
                  value={tanggal}
                  onChange={(e) => setTanggal(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Buku Register
                <input
                  type="text"
                  value={buku}
                  onChange={(e) => setBuku(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Nomor Register
                <input
                  type="text"
                  value={nomor}
                  onChange={(e) => setNomor(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>
            </div>
          </InductiveFormSection>

          <InductiveFormSection
            title="Tempat Penerimaan Sakramen"
            description="Pilih dari daftar Keuskupan/Paroki atau isi teks manual jika di luar sistem."
          >
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Keuskupan (Master)
                <select
                  value={keuskupanId}
                  onChange={(e) => {
                    setKeuskupanId(e.target.value);
                    setParokiSakramenId("");
                  }}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  <option value="">-- Tidak dipilih --</option>
                  {(keuskupanQuery.data ?? []).map((k) => (
                    <option key={k.id} value={k.id}>
                      {k.nama}
                    </option>
                  ))}
                </select>
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Paroki Sakramen (Master)
                <select
                  value={parokiSakramenId}
                  onChange={(e) => {
                    const nextParokiId = e.target.value;
                    setParokiSakramenId(nextParokiId);
                    const found = (parokiQuery.data ?? []).find(
                      (p) => p.id === nextParokiId,
                    );
                    if (found?.keuskupanId && !keuskupanId) {
                      setKeuskupanId(found.keuskupanId);
                    }
                  }}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  <option value="">-- Tidak dipilih --</option>
                  {(parokiQuery.data ?? []).map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.nama}
                    </option>
                  ))}
                </select>
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Nama Keuskupan (Teks Bebas)
                <input
                  type="text"
                  value={keuskupanNama}
                  onChange={(e) => setKeuskupanNama(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Nama Paroki (Teks Bebas)
                <input
                  type="text"
                  value={parokiNama}
                  onChange={(e) => setParokiNama(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>
            </div>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Catatan
              <textarea
                rows={3}
                value={catatan}
                onChange={(e) => setCatatan(e.target.value)}
                className="rounded-lg border border-line bg-canvas px-3.5 py-2.5 text-sm text-fg"
              />
            </label>
          </InductiveFormSection>

          <div className="flex flex-col-reverse gap-3 sm:flex-row sm:items-center sm:justify-end">
            <Link
              to={`/umat/${umatId}`}
              className="w-full no-underline sm:w-auto"
            >
              <Button
                type="button"
                variant="outline"
                className="w-full sm:w-auto"
              >
                Batal
              </Button>
            </Link>
            <Button
              type="submit"
              className="w-full sm:w-auto"
              disabled={isPending}
            >
              {isPending ? "Menyimpan…" : "Simpan Sakramen"}
            </Button>
          </div>
        </form>
      )}
    </div>
  );
}
