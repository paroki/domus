import { useEffect, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
import { ApiError } from "~/shared/api/client";
import { useMe } from "~/shared/api/session";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useCreateUmat } from "../api/actions";
import { resolveUmatErrorMessage } from "../api/errors";
import {
  useKeluargaDetail,
  useKeluargaSearch,
  useUmatKecamatanList,
  useUmatKelurahanList,
  useUmatKotaList,
  useUmatLingkunganList,
  useUmatLookups,
  useUmatWilayahList,
} from "../api/queries";
import { UmatForbiddenCard } from "../components/UmatForbiddenCard";
import { isUmatAdmin, validatePIIInput, validateYYYYMMDD } from "../model";

export function meta() {
  return [{ title: "Tambah Data Umat · Data Umat · Domus" }];
}

const STEP_LABELS = [
  "1. Identitas Pribadi, Keluarga & Lingkungan",
  "2. Kontak, Domisili Override & Status Hidup",
  "3. Pendidikan, Pekerjaan & Aktivitas Sosial",
] as const;

export default function UmatCreateWizardPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { data: me } = useMe();

  const activeParokiId = me?.activeParokiId ?? "";
  const allowed = isUmatAdmin(me);

  const initialKeluargaId = searchParams.get("keluargaId") ?? "";
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Step 1
  const [nama, setNama] = useState("");
  const [namaBaptis, setNamaBaptis] = useState("");
  const [nik, setNik] = useState("");
  const [jenisKelamin, setJenisKelamin] = useState<"pria" | "wanita">("pria");
  const [tempatLahir, setTempatLahir] = useState("");
  const [tanggalLahir, setTanggalLahir] = useState("");
  const [keluargaId, setKeluargaId] = useState(initialKeluargaId);
  const [hubunganKeluarga, setHubunganKeluarga] = useState("");
  const [wilayahId, setWilayahId] = useState(
    searchParams.get("wilayahId") ?? "",
  );
  const [lingkunganId, setLingkunganId] = useState(
    searchParams.get("lingkunganId") ?? "",
  );
  const [agama, setAgama] = useState("");
  const [statusPernikahan, setStatusPernikahan] = useState("");

  // Step 2
  const [telepon, setTelepon] = useState("");
  const [email, setEmail] = useState("");
  const [fotoUrl, setFotoUrl] = useState("");
  const [alamat, setAlamat] = useState("");
  const [kotaId, setKotaId] = useState("");
  const [kecamatanId, setKecamatanId] = useState("");
  const [kelurahanId, setKelurahanId] = useState("");
  const [golonganDarah, setGolonganDarah] = useState("");
  const [suku, setSuku] = useState("");
  const [kewarganegaraan, setKewarganegaraan] = useState("");
  const [negara, setNegara] = useState("");
  const [almarhum, setAlmarhum] = useState(false);
  const [tanggalMeninggal, setTanggalMeninggal] = useState("");

  // Step 3
  const [pendidikan, setPendidikan] = useState("");
  const [jurusan, setJurusan] = useState("");
  const [pekerjaan, setPekerjaan] = useState("");
  const [profesi, setProfesi] = useState("");
  const [keterampilan, setKeterampilan] = useState("");
  const [disabilitas, setDisabilitas] = useState("");
  const [aktivitasSosial, setAktivitasSosial] = useState("");

  const createUmat = useCreateUmat(activeParokiId);

  const keluargaListQuery = useKeluargaSearch(
    activeParokiId,
    { limit: 100 },
    Boolean(me && allowed && activeParokiId),
  );
  const selectedKeluargaQuery = useKeluargaDetail(
    activeParokiId,
    keluargaId || undefined,
    Boolean(me && allowed && activeParokiId && keluargaId),
  );
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

  const hubunganQuery = useUmatLookups(
    activeParokiId,
    "hubungan_keluarga",
    Boolean(me && allowed && activeParokiId),
  );
  const agamaQuery = useUmatLookups(
    activeParokiId,
    "agama",
    Boolean(me && allowed && activeParokiId),
  );
  const statusPernikahanQuery = useUmatLookups(
    activeParokiId,
    "status_pernikahan",
    Boolean(me && allowed && activeParokiId),
  );
  const golDarahQuery = useUmatLookups(
    activeParokiId,
    "golongan_darah",
    Boolean(me && allowed && activeParokiId),
  );
  const sukuQuery = useUmatLookups(
    activeParokiId,
    "suku",
    Boolean(me && allowed && activeParokiId),
  );
  const pendidikanQuery = useUmatLookups(
    activeParokiId,
    "pendidikan",
    Boolean(me && allowed && activeParokiId),
  );
  const pekerjaanQuery = useUmatLookups(
    activeParokiId,
    "pekerjaan",
    Boolean(me && allowed && activeParokiId),
  );
  const profesiQuery = useUmatLookups(
    activeParokiId,
    "profesi",
    Boolean(me && allowed && activeParokiId),
  );
  const disabilitasQuery = useUmatLookups(
    activeParokiId,
    "disabilitas",
    Boolean(me && allowed && activeParokiId),
  );
  const aktivitasQuery = useUmatLookups(
    activeParokiId,
    "aktivitas_sosial",
    Boolean(me && allowed && activeParokiId),
  );

  // Auto-fill wilayahId/lingkunganId from selected Keluarga if still empty.
  useEffect(() => {
    const kk = selectedKeluargaQuery.data;
    if (!kk) return;
    if (!wilayahId && kk.wilayahId) {
      setWilayahId(kk.wilayahId);
    }
    if (!lingkunganId && kk.lingkunganId) {
      setLingkunganId(kk.lingkunganId);
    }
  }, [selectedKeluargaQuery.data, wilayahId, lingkunganId]);

  if (me && (!allowed || !activeParokiId)) {
    return <UmatForbiddenCard />;
  }

  const keluargaOptions =
    keluargaListQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const allLingkungan = lingkunganQuery.data ?? [];
  const filteredLingkungan = wilayahId
    ? allLingkungan.filter((l) => l.wilayahId === wilayahId)
    : allLingkungan;

  function validateStep1(): string | null {
    if (!nama.trim()) return "Nama lengkap wajib diisi.";
    const nikErr = validatePIIInput(nik, "NIK");
    if (nikErr) return nikErr;
    const birthErr = validateYYYYMMDD(tanggalLahir, "Tanggal lahir");
    if (birthErr) return birthErr;
    return null;
  }

  function handleNextFromStep1() {
    const err = validateStep1();
    if (err) {
      setErrorMsg(err);
      return;
    }
    setErrorMsg(null);
    setStep(2);
  }

  async function handleSubmit() {
    const step1Err = validateStep1();
    if (step1Err) {
      setErrorMsg(step1Err);
      setStep(1);
      return;
    }
    const deathErr = almarhum
      ? validateYYYYMMDD(tanggalMeninggal, "Tanggal meninggal")
      : null;
    if (deathErr) {
      setErrorMsg(deathErr);
      setStep(2);
      return;
    }
    setErrorMsg(null);
    try {
      const created = await createUmat.mutateAsync({
        nama: nama.trim(),
        jenisKelamin,
        ...(namaBaptis.trim() ? { namaBaptis: namaBaptis.trim() } : {}),
        ...(nik.trim() ? { nik: nik.trim() } : {}),
        ...(tempatLahir.trim() ? { tempatLahir: tempatLahir.trim() } : {}),
        ...(tanggalLahir.trim() ? { tanggalLahir: tanggalLahir.trim() } : {}),
        ...(keluargaId ? { keluargaId } : {}),
        ...(hubunganKeluarga ? { hubunganKeluarga } : {}),
        ...(wilayahId ? { wilayahId } : {}),
        ...(lingkunganId ? { lingkunganId } : {}),
        ...(agama ? { agama } : {}),
        ...(statusPernikahan ? { statusPernikahan } : {}),
        ...(telepon.trim() ? { telepon: telepon.trim() } : {}),
        ...(email.trim() ? { email: email.trim() } : {}),
        ...(fotoUrl.trim() ? { fotoUrl: fotoUrl.trim() } : {}),
        ...(alamat.trim() ? { alamat: alamat.trim() } : {}),
        ...(kotaId ? { kotaId } : {}),
        ...(kecamatanId ? { kecamatanId } : {}),
        ...(kelurahanId ? { kelurahanId } : {}),
        ...(golonganDarah ? { golonganDarah } : {}),
        ...(suku ? { suku } : {}),
        ...(kewarganegaraan.trim()
          ? { kewarganegaraan: kewarganegaraan.trim() }
          : {}),
        ...(negara.trim() ? { negara: negara.trim() } : {}),
        almarhum,
        ...(almarhum && tanggalMeninggal.trim()
          ? { tanggalMeninggal: tanggalMeninggal.trim() }
          : {}),
        ...(pendidikan ? { pendidikan } : {}),
        ...(jurusan.trim() ? { jurusan: jurusan.trim() } : {}),
        ...(pekerjaan ? { pekerjaan } : {}),
        ...(profesi ? { profesi } : {}),
        ...(keterampilan.trim() ? { keterampilan: keterampilan.trim() } : {}),
        ...(disabilitas ? { disabilitas } : {}),
        ...(aktivitasSosial ? { aktivitasSosial } : {}),
      });
      if (created?.id) {
        navigate(`/umat/${created.id}`);
      } else {
        navigate("/umat");
      }
    } catch (err) {
      if (err instanceof ApiError && err.code === "UMAT_NIK_CONFLICT") {
        setStep(1);
      }
      setErrorMsg(resolveUmatErrorMessage(err));
    }
  }

  const breadcrumbs = [
    { label: "Beranda", to: "/" },
    { label: "Data Umat", to: "/umat" },
    ...(initialKeluargaId && selectedKeluargaQuery.data?.namaKeluarga
      ? [
          {
            label: selectedKeluargaQuery.data.namaKeluarga,
            to: `/umat/keluarga/${initialKeluargaId}`,
          },
        ]
      : []),
    { label: "Tambah Data Umat" },
  ];

  const cancelHref = initialKeluargaId
    ? `/umat/keluarga/${initialKeluargaId}`
    : "/umat";

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={breadcrumbs}
        title="Tambah Data Umat"
        subtitle="Lengkapi profil umat paroki secara bertahap atau simpan langsung setelah identitas dasar."
      />

      <Card className="flex flex-wrap items-center gap-2 p-4">
        {STEP_LABELS.map((label, idx) => {
          const num = (idx + 1) as 1 | 2 | 3;
          const active = step === num;
          return (
            <span
              key={label}
              className={`inline-flex min-h-10 items-center rounded-full px-3.5 py-1.5 text-xs font-semibold ${
                active
                  ? "bg-primary text-on-primary"
                  : "bg-surface-2 text-muted"
              }`}
            >
              {label}
            </span>
          );
        })}
      </Card>

      {errorMsg ? (
        <div
          role="alert"
          className="rounded-lg border border-danger/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
        >
          {errorMsg}
        </div>
      ) : null}

      {step === 1 ? (
        <InductiveFormSection
          title="Langkah 1 — Identitas Pribadi, Keluarga & Lingkungan"
          description="Nama lengkap dan jenis kelamin wajib diisi."
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Nama Lengkap *
              <input
                type="text"
                required
                value={nama}
                onChange={(e) => setNama(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Nama Baptis
              <input
                type="text"
                value={namaBaptis}
                onChange={(e) => setNamaBaptis(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              NIK
              <input
                type="text"
                value={nik}
                onChange={(e) => setNik(e.target.value)}
                placeholder="4–32 karakter (opsional)"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 font-mono text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Jenis Kelamin *
              <select
                value={jenisKelamin}
                onChange={(e) =>
                  setJenisKelamin(e.target.value as "pria" | "wanita")
                }
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="pria">Pria</option>
                <option value="wanita">Wanita</option>
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Tempat Lahir
              <input
                type="text"
                value={tempatLahir}
                onChange={(e) => setTempatLahir(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Tanggal Lahir
              <input
                type="date"
                value={tanggalLahir}
                onChange={(e) => setTanggalLahir(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Kartu Keluarga
              <select
                value={keluargaId}
                onChange={(e) => {
                  const nextId = e.target.value;
                  setKeluargaId(nextId);
                  const found = keluargaOptions.find((k) => k.id === nextId);
                  if (found) {
                    if (!wilayahId && found.wilayahId) {
                      setWilayahId(found.wilayahId);
                    }
                    if (!lingkunganId && found.lingkunganId) {
                      setLingkunganId(found.lingkunganId);
                    }
                  }
                }}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Tidak terhubung KK --</option>
                {keluargaOptions.map((k) => (
                  <option key={k.id} value={k.id}>
                    {k.namaKeluarga}
                    {k.nomorKkMasked ? ` (${k.nomorKkMasked})` : ""}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Hubungan Keluarga
              <select
                value={hubunganKeluarga}
                onChange={(e) => setHubunganKeluarga(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Hubungan --</option>
                {(hubunganQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
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
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Agama
              <select
                value={agama}
                onChange={(e) => setAgama(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Agama --</option>
                {(agamaQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Status Pernikahan
              <select
                value={statusPernikahan}
                onChange={(e) => setStatusPernikahan(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Status --</option>
                {(statusPernikahanQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="flex flex-col-reverse gap-3 pt-2 sm:flex-row sm:items-center sm:justify-between">
            <Link to={cancelHref} className="w-full no-underline sm:w-auto">
              <Button
                type="button"
                variant="outline"
                className="w-full sm:w-auto"
              >
                Batal
              </Button>
            </Link>
            <div className="flex flex-col-reverse gap-2 sm:flex-row sm:items-center">
              <Button
                type="button"
                variant="outline"
                className="w-full sm:w-auto"
                disabled={!nama.trim() || createUmat.isPending}
                onClick={handleSubmit}
              >
                Simpan & Buat Umat
              </Button>
              <Button
                type="button"
                className="w-full sm:w-auto"
                onClick={handleNextFromStep1}
              >
                Lanjut ke Kontak & Domisili →
              </Button>
            </div>
          </div>
        </InductiveFormSection>
      ) : null}

      {step === 2 ? (
        <InductiveFormSection
          title="Langkah 2 — Kontak, Domisili Override & Status Hidup"
          description="Isi kontak pribadi dan domisili jika berbeda dengan Kartu Keluarga."
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Telepon
              <input
                type="tel"
                value={telepon}
                onChange={(e) => setTelepon(e.target.value)}
                placeholder="0812..."
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Email
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
              URL Foto
              <input
                type="url"
                value={fotoUrl}
                onChange={(e) => setFotoUrl(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
              Alamat Domisili (Override)
              <textarea
                rows={2}
                value={alamat}
                onChange={(e) => setAlamat(e.target.value)}
                className="rounded-lg border border-line bg-canvas px-3.5 py-2.5 text-sm text-fg"
              />
            </label>
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
                onChange={(e) => setKelurahanId(e.target.value)}
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
              Golongan Darah
              <select
                value={golonganDarah}
                onChange={(e) => setGolonganDarah(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Golongan Darah --</option>
                {(golDarahQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Suku
              <select
                value={suku}
                onChange={(e) => setSuku(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Suku --</option>
                {(sukuQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Kewarganegaraan
              <input
                type="text"
                value={kewarganegaraan}
                onChange={(e) => setKewarganegaraan(e.target.value)}
                placeholder="WNI / WNA"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Negara
              <input
                type="text"
                value={negara}
                onChange={(e) => setNegara(e.target.value)}
                placeholder="Indonesia"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
          </div>

          <div className="flex flex-wrap items-center gap-4 pt-1">
            <label className="flex min-h-11 items-center gap-2.5 text-sm font-medium text-fg">
              <input
                type="checkbox"
                checked={almarhum}
                onChange={(e) => setAlmarhum(e.target.checked)}
                className="size-4 rounded border-line"
              />
              Almarhum
            </label>
            {almarhum ? (
              <label className="flex flex-col gap-1 text-sm font-medium text-fg">
                Tanggal Meninggal
                <input
                  type="date"
                  value={tanggalMeninggal}
                  onChange={(e) => setTanggalMeninggal(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>
            ) : null}
          </div>

          <div className="flex flex-col-reverse gap-3 pt-2 sm:flex-row sm:items-center sm:justify-between">
            <Button
              type="button"
              variant="outline"
              className="w-full sm:w-auto"
              onClick={() => setStep(1)}
            >
              ← Kembali
            </Button>
            <div className="flex flex-col-reverse gap-2 sm:flex-row sm:items-center">
              <Button
                type="button"
                variant="outline"
                className="w-full sm:w-auto"
                disabled={createUmat.isPending}
                onClick={handleSubmit}
              >
                Simpan & Buat Umat
              </Button>
              <Button
                type="button"
                className="w-full sm:w-auto"
                onClick={() => setStep(3)}
              >
                Lanjut ke Pendidikan & Sosial →
              </Button>
            </div>
          </div>
        </InductiveFormSection>
      ) : null}

      {step === 3 ? (
        <InductiveFormSection
          title="Langkah 3 — Pendidikan, Pekerjaan & Aktivitas Sosial"
          description="Informasi latar belakang pendidikan, profesi, dan pelayanan sosial."
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Pendidikan Terakhir
              <select
                value={pendidikan}
                onChange={(e) => setPendidikan(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Pendidikan --</option>
                {(pendidikanQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Jurusan
              <input
                type="text"
                value={jurusan}
                onChange={(e) => setJurusan(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Pekerjaan
              <select
                value={pekerjaan}
                onChange={(e) => setPekerjaan(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Pekerjaan --</option>
                {(pekerjaanQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Profesi
              <select
                value={profesi}
                onChange={(e) => setProfesi(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Profesi --</option>
                {(profesiQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Keterampilan
              <input
                type="text"
                value={keterampilan}
                onChange={(e) => setKeterampilan(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Disabilitas
              <select
                value={disabilitas}
                onChange={(e) => setDisabilitas(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Tidak ada / Pilih --</option>
                {(disabilitasQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
              Aktivitas Sosial
              <select
                value={aktivitasSosial}
                onChange={(e) => setAktivitasSosial(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Pilih Aktivitas Sosial --</option>
                {(aktivitasQuery.data ?? []).map((l) => (
                  <option key={l.id} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="flex flex-col-reverse gap-3 pt-2 sm:flex-row sm:items-center sm:justify-between">
            <Button
              type="button"
              variant="outline"
              className="w-full sm:w-auto"
              onClick={() => setStep(2)}
            >
              ← Kembali
            </Button>
            <Button
              type="button"
              className="w-full sm:w-auto"
              disabled={createUmat.isPending}
              onClick={handleSubmit}
            >
              {createUmat.isPending ? "Menyimpan…" : "Simpan & Buat Umat"}
            </Button>
          </div>
        </InductiveFormSection>
      ) : null}
    </div>
  );
}
