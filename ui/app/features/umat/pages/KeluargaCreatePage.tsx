import { type FormEvent, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button } from "~/shared/ui/Button";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useCreateKeluarga } from "../api/actions";
import { resolveUmatErrorMessage } from "../api/errors";
import {
  useUmatKecamatanList,
  useUmatKelurahanList,
  useUmatKotaList,
  useUmatLingkunganList,
  useUmatLookups,
  useUmatWilayahList,
} from "../api/queries";
import { UmatForbiddenCard } from "../components/UmatForbiddenCard";
import { isUmatAdmin, validatePIIInput } from "../model";

export function meta() {
  return [{ title: "Tambah Kartu Keluarga · Data Umat · Domus" }];
}

export default function KeluargaCreatePage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { data: me } = useMe();

  const activeParokiId = me?.activeParokiId ?? "";
  const allowed = isUmatAdmin(me);

  const [namaKeluarga, setNamaKeluarga] = useState("");
  const [nomorKk, setNomorKk] = useState("");
  const [wilayahId, setWilayahId] = useState(
    searchParams.get("wilayahId") ?? "",
  );
  const [lingkunganId, setLingkunganId] = useState(
    searchParams.get("lingkunganId") ?? "",
  );
  const [aktif, setAktif] = useState(true);

  const [alamat, setAlamat] = useState("");
  const [kotaId, setKotaId] = useState("");
  const [kecamatanId, setKecamatanId] = useState("");
  const [kelurahanId, setKelurahanId] = useState("");
  const [kodePos, setKodePos] = useState("");

  const [statusRumah, setStatusRumah] = useState("");
  const [statusEkonomi, setStatusEkonomi] = useState("");
  const [catatan, setCatatan] = useState("");

  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const createKeluarga = useCreateKeluarga(activeParokiId);

  const wilayahQuery = useUmatWilayahList(
    activeParokiId,
    Boolean(me && allowed && activeParokiId),
  );
  const lingkunganQuery = useUmatLingkunganList(
    activeParokiId,
    undefined,
    Boolean(me && allowed && activeParokiId),
  );
  const kotaQuery = useUmatKotaList(Boolean(me && allowed && activeParokiId));
  const kecamatanQuery = useUmatKecamatanList(
    kotaId || undefined,
    Boolean(me && allowed && kotaId),
  );
  const kelurahanQuery = useUmatKelurahanList(
    kecamatanId || undefined,
    Boolean(me && allowed && kecamatanId),
  );
  const statusRumahQuery = useUmatLookups(
    activeParokiId,
    "status_rumah",
    Boolean(me && allowed && activeParokiId),
  );
  const statusEkonomiQuery = useUmatLookups(
    activeParokiId,
    "status_ekonomi",
    Boolean(me && allowed && activeParokiId),
  );

  if (me && (!allowed || !activeParokiId)) {
    return <UmatForbiddenCard />;
  }

  const allLingkungan = lingkunganQuery.data ?? [];
  const filteredLingkungan = wilayahId
    ? allLingkungan.filter((l) => l.wilayahId === wilayahId)
    : allLingkungan;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const kkErr = validatePIIInput(nomorKk, "Nomor KK");
    if (kkErr) {
      setErrorMsg(kkErr);
      return;
    }
    setErrorMsg(null);
    try {
      const created = await createKeluarga.mutateAsync({
        namaKeluarga: namaKeluarga.trim(),
        ...(nomorKk.trim() ? { nomorKk: nomorKk.trim() } : {}),
        ...(wilayahId ? { wilayahId } : {}),
        ...(lingkunganId ? { lingkunganId } : {}),
        aktif,
        ...(alamat.trim() ? { alamat: alamat.trim() } : {}),
        ...(kotaId ? { kotaId } : {}),
        ...(kecamatanId ? { kecamatanId } : {}),
        ...(kelurahanId ? { kelurahanId } : {}),
        ...(kodePos.trim() ? { kodePos: kodePos.trim() } : {}),
        ...(statusRumah ? { statusRumah } : {}),
        ...(statusEkonomi ? { statusEkonomi } : {}),
        ...(catatan.trim() ? { catatan: catatan.trim() } : {}),
      });
      if (created?.id) {
        navigate(`/umat/keluarga/${created.id}`);
      } else {
        navigate("/umat/keluarga");
      }
    } catch (err) {
      setErrorMsg(resolveUmatErrorMessage(err));
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Data Umat", to: "/umat" },
          { label: "Kartu Keluarga", to: "/umat/keluarga" },
          { label: "Tambah Kartu Keluarga" },
        ]}
        title="Tambah Kartu Keluarga"
        subtitle="Buat data Kartu Keluarga baru untuk paroki aktif."
      />

      {errorMsg ? (
        <div
          role="alert"
          className="rounded-lg border border-danger/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
        >
          {errorMsg}
        </div>
      ) : null}

      <form onSubmit={handleSubmit} className="flex flex-col gap-5">
        <InductiveFormSection
          title="1. Identitas Kartu Keluarga & Teritorial Gerejawi"
          description="Nama keluarga wajib diisi; Nomor KK bersifat opsional dan dienkripsi."
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Nama Keluarga *
              <input
                type="text"
                required
                value={namaKeluarga}
                onChange={(e) => setNamaKeluarga(e.target.value)}
                placeholder="Contoh: Keluarga Santoso"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Nomor KK
              <input
                type="text"
                value={nomorKk}
                onChange={(e) => setNomorKk(e.target.value)}
                placeholder="4–32 karakter (opsional)"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 font-mono text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Wilayah
              <select
                value={wilayahId}
                onChange={(e) => {
                  const nextWil = e.target.value;
                  setWilayahId(nextWil);
                  if (
                    lingkunganId &&
                    allLingkungan.find((l) => l.id === lingkunganId)
                      ?.wilayahId !== nextWil
                  ) {
                    setLingkunganId("");
                  }
                }}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Wilayah --</option>
                {(wilayahQuery.data ?? []).map((w) => (
                  <option key={w.id} value={w.id}>
                    {w.nama}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Lingkungan
              <select
                value={lingkunganId}
                onChange={(e) => {
                  const nextLing = e.target.value;
                  setLingkunganId(nextLing);
                  const found = allLingkungan.find((l) => l.id === nextLing);
                  if (found?.wilayahId && !wilayahId) {
                    setWilayahId(found.wilayahId);
                  }
                }}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Lingkungan --</option>
                {filteredLingkungan.map((l) => (
                  <option key={l.id} value={l.id}>
                    {l.nama}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <label className="flex items-center gap-2 text-sm font-medium text-fg">
            <input
              type="checkbox"
              checked={aktif}
              onChange={(e) => setAktif(e.target.checked)}
              className="size-4 rounded border-line"
            />
            Kartu Keluarga Aktif
          </label>
        </InductiveFormSection>

        <InductiveFormSection
          title="2. Alamat Domisili & Wilayah Sipil"
          description="Pilih wilayah administratif secara berjenjang jika tersedia."
        >
          <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
            Alamat Lengkap
            <textarea
              rows={3}
              value={alamat}
              onChange={(e) => setAlamat(e.target.value)}
              className="rounded-lg border border-line bg-canvas px-3.5 py-2.5 text-sm text-fg"
            />
          </label>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Kota / Kabupaten
              <select
                value={kotaId}
                onChange={(e) => {
                  setKotaId(e.target.value);
                  setKecamatanId("");
                  setKelurahanId("");
                }}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Kota --</option>
                {(kotaQuery.data ?? []).map((k) => (
                  <option key={k.id} value={k.id}>
                    {k.nama}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Kecamatan
              <select
                value={kecamatanId}
                disabled={!kotaId}
                onChange={(e) => {
                  setKecamatanId(e.target.value);
                  setKelurahanId("");
                }}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg disabled:opacity-60"
              >
                <option value="">-- Pilih Kecamatan --</option>
                {(kecamatanQuery.data ?? []).map((k) => (
                  <option key={k.id} value={k.id}>
                    {k.nama}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Kelurahan / Desa
              <select
                value={kelurahanId}
                disabled={!kecamatanId}
                onChange={(e) => {
                  const nextKel = e.target.value;
                  setKelurahanId(nextKel);
                  const found = (kelurahanQuery.data ?? []).find(
                    (k) => k.id === nextKel,
                  );
                  if (found?.kodePos && !kodePos.trim()) {
                    setKodePos(found.kodePos);
                  }
                }}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg disabled:opacity-60"
              >
                <option value="">-- Pilih Kelurahan --</option>
                {(kelurahanQuery.data ?? []).map((k) => (
                  <option key={k.id} value={k.id}>
                    {k.nama}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Kode Pos
              <input
                type="text"
                value={kodePos}
                onChange={(e) => setKodePos(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
          </div>
        </InductiveFormSection>

        <InductiveFormSection
          title="3. Kondisi Rumah Tangga & Catatan"
          description="Informasi pendukung pelayanan pastoral keluarga."
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Status Rumah
              <select
                value={statusRumah}
                onChange={(e) => setStatusRumah(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Status Rumah --</option>
                {(statusRumahQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Status Ekonomi
              <select
                value={statusEkonomi}
                onChange={(e) => setStatusEkonomi(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Status Ekonomi --</option>
                {(statusEkonomiQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
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

        <div className="flex items-center justify-end gap-3">
          <Link to="/umat/keluarga" className="no-underline">
            <Button type="button" variant="outline">
              Batal
            </Button>
          </Link>
          <Button type="submit" disabled={createKeluarga.isPending}>
            {createKeluarga.isPending ? "Menyimpan…" : "Simpan Kartu Keluarga"}
          </Button>
        </div>
      </form>
    </div>
  );
}
