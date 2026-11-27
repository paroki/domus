import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button, buttonClass } from "~/shared/ui/Button";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useActivateSessionParoki, useCreateParoki } from "../api/actions";
import {
  useKecamatanSearch,
  useKelurahanSearch,
  useKeuskupanSearch,
  useKevikepanSearch,
  useKotaSearch,
  useLookupsSearch,
  useParokiSearch,
} from "../api/queries";
import {
  DEFAULT_PAROKI_TIMEZONE,
  isSuperAdmin,
  PAROKI_TIMEZONES,
  slugifyCode,
  validateCoordinatePair,
} from "../model";

export function meta() {
  return [{ title: "Tambah Paroki — Domus" }];
}

export default function ParokiCreateWizardPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  const createMut = useCreateParoki();
  const activateMut = useActivateSessionParoki();

  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Step 1: Struktur Gerejawi & Kode
  const [keuskupanId, setKeuskupanId] = useState(
    searchParams.get("keuskupanId") ?? "",
  );
  const [kevikepanId, setKevikepanId] = useState(
    searchParams.get("kevikepanId") ?? "",
  );
  const [indukId, setIndukId] = useState("");
  const [nama, setNama] = useState("");
  const [kode, setKode] = useState("");
  const [kodeTouched, setKodeTouched] = useState(false);
  const [kodeResmi, setKodeResmi] = useState("");
  const [statusParoki, setStatusParoki] = useState("");

  // Step 2: Identitas Gereja & Kontak
  const [pelindung, setPelindung] = useState("");
  const [pastor, setPastor] = useState("");
  const [telepon, setTelepon] = useState("");
  const [fax, setFax] = useState("");
  const [email, setEmail] = useState("");
  const [website, setWebsite] = useState("");
  const [logoUrl, setLogoUrl] = useState("");
  const [kopSurat, setKopSurat] = useState("");

  // Step 3: Domisili Wilayah Sipil, Zona Waktu & Koordinat
  const [timezone, setTimezone] = useState(DEFAULT_PAROKI_TIMEZONE);
  const [kotaId, setKotaId] = useState("");
  const [kecamatanId, setKecamatanId] = useState("");
  const [kelurahanId, setKelurahanId] = useState("");
  const [alamat, setAlamat] = useState("");
  const [latitudeStr, setLatitudeStr] = useState("");
  const [longitudeStr, setLongitudeStr] = useState("");

  const keuskupanQuery = useKeuskupanSearch({}, canManage);
  const keuskupanList =
    keuskupanQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const kevikepanQuery = useKevikepanSearch(
    { filter: keuskupanId ? { keuskupanId } : undefined },
    canManage && Boolean(keuskupanId),
  );
  const kevikepanList =
    kevikepanQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const rootParokiQuery = useParokiSearch(
    { filter: keuskupanId ? { keuskupanId } : undefined },
    canManage && Boolean(keuskupanId),
  );
  const rootParokiList = (
    rootParokiQuery.data?.pages.flatMap((p) => p.data ?? []) ?? []
  ).filter((p) => !p.indukId);

  const statusLookups = useLookupsSearch(
    { filter: { category: "status_paroki", scope: "all" } },
    canManage,
  );
  const statusOptions =
    statusLookups.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const kotaQuery = useKotaSearch({}, canManage);
  const kotaList = kotaQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const kecQuery = useKecamatanSearch(
    { filter: kotaId ? { kotaId } : undefined },
    canManage && Boolean(kotaId),
  );
  const kecList = kecQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const kelQuery = useKelurahanSearch(
    { filter: kecamatanId ? { kecamatanId } : undefined },
    canManage && Boolean(kecamatanId),
  );
  const kelList = kelQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  async function submitCreate() {
    setErrorMsg(null);
    if (!keuskupanId || !nama.trim() || !(kode || slugifyCode(nama)).trim()) {
      setErrorMsg("Keuskupan, nama, dan kode paroki wajib diisi.");
      setStep(1);
      return;
    }
    const coords = validateCoordinatePair(latitudeStr, longitudeStr);
    if (!coords.valid) {
      setErrorMsg(coords.error ?? "Koordinat tidak valid.");
      return;
    }

    try {
      const created = await createMut.mutateAsync({
        keuskupanId,
        kevikepanId: kevikepanId || undefined,
        indukId: indukId || undefined,
        kode: (kode || slugifyCode(nama)).trim(),
        kodeResmi: kodeResmi.trim() || undefined,
        nama: nama.trim(),
        statusParoki: statusParoki || undefined,
        pelindung: pelindung.trim() || undefined,
        pastor: pastor.trim() || undefined,
        telepon: telepon.trim() || undefined,
        fax: fax.trim() || undefined,
        email: email.trim() || undefined,
        website: website.trim() || undefined,
        logoUrl: logoUrl.trim() || undefined,
        kopSurat: kopSurat.trim() || undefined,
        timezone: timezone || DEFAULT_PAROKI_TIMEZONE,
        kotaId: kotaId || undefined,
        kecamatanId: kecamatanId || undefined,
        kelurahanId: kelurahanId || undefined,
        alamat: alamat.trim() || undefined,
        latitude: coords.latitude,
        longitude: coords.longitude,
      });
      if (!me?.activeParokiId && created.id) {
        await activateMut.mutateAsync(created.id).catch(() => null);
      }
      navigate(`/settings/gerejawi/paroki/${created.id}`);
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal membuat paroki.");
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Struktur Gerejawi", to: "/settings/gerejawi" },
          { label: "Paroki", to: "/settings/gerejawi/paroki" },
          { label: "Tambah Paroki" },
        ]}
        title="Tambah Paroki Baru"
        subtitle={`Langkah ${step} dari 3 — lengkapi struktur gerejawi, identitas, dan domisili paroki.`}
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat menambahkan paroki.
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
          {step === 1 ? (
            <InductiveFormSection
              title="Langkah 1: Struktur Gerejawi &amp; Kode"
              description="Tentukan keuskupan induk, kevikepan, serta kode unik paroki."
            >
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-keuskupan"
                    className="text-xs font-semibold text-muted"
                  >
                    Keuskupan
                  </label>
                  <select
                    id="wiz-keuskupan"
                    required
                    value={keuskupanId}
                    onChange={(e) => {
                      setKeuskupanId(e.target.value);
                      setKevikepanId("");
                      setIndukId("");
                    }}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  >
                    <option value="">Pilih keuskupan…</option>
                    {keuskupanList.map((k) => (
                      <option key={k.id} value={k.id}>
                        {k.nama}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-kevikepan"
                    className="text-xs font-semibold text-muted"
                  >
                    Kevikepan (opsional)
                  </label>
                  <select
                    id="wiz-kevikepan"
                    value={kevikepanId}
                    onChange={(e) => setKevikepanId(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  >
                    <option value="">Tanpa kevikepan</option>
                    {kevikepanList.map((k) => (
                      <option key={k.id} value={k.id}>
                        {k.nama}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-nama"
                    className="text-xs font-semibold text-muted"
                  >
                    Nama Paroki
                  </label>
                  <input
                    id="wiz-nama"
                    required
                    value={nama}
                    onChange={(e) => {
                      const next = e.target.value;
                      setNama(next);
                      if (!kodeTouched) {
                        setKode(slugifyCode(next));
                      }
                    }}
                    placeholder="mis. Paroki Katedral Jakarta"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>

                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-kode"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode Paroki
                  </label>
                  <input
                    id="wiz-kode"
                    required
                    value={kode}
                    onChange={(e) => {
                      setKodeTouched(true);
                      setKode(e.target.value);
                    }}
                    placeholder="mis. katedral-jkt"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 font-mono text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-induk"
                    className="text-xs font-semibold text-muted"
                  >
                    Paroki Induk (opsional)
                  </label>
                  <select
                    id="wiz-induk"
                    value={indukId}
                    onChange={(e) => setIndukId(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  >
                    <option value="">Paroki Mandiri</option>
                    {rootParokiList.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.nama}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-koderesmi"
                    className="text-xs font-semibold text-muted"
                  >
                    Kode Resmi (opsional)
                  </label>
                  <input
                    id="wiz-koderesmi"
                    value={kodeResmi}
                    onChange={(e) => setKodeResmi(e.target.value)}
                    placeholder="mis. KAJ-001"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>

                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-status"
                    className="text-xs font-semibold text-muted"
                  >
                    Status Paroki (opsional)
                  </label>
                  <select
                    id="wiz-status"
                    value={statusParoki}
                    onChange={(e) => setStatusParoki(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  >
                    <option value="">Pilih status…</option>
                    {statusOptions.map((s) => (
                      <option key={s.id} value={s.code}>
                        {s.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="flex flex-col-reverse items-stretch justify-end gap-2 pt-2 sm:flex-row sm:items-center">
                <Link
                  to="/settings/gerejawi/paroki"
                  className={`${buttonClass("outline")} w-full sm:w-auto`}
                >
                  Batal
                </Link>
                <Button
                  className="w-full sm:w-auto"
                  onClick={() => {
                    if (
                      !keuskupanId ||
                      !nama.trim() ||
                      !(kode || slugifyCode(nama)).trim()
                    ) {
                      setErrorMsg(
                        "Keuskupan, nama, dan kode paroki wajib diisi.",
                      );
                      return;
                    }
                    setErrorMsg(null);
                    setStep(2);
                  }}
                >
                  Lanjut ke Identitas →
                </Button>
              </div>
            </InductiveFormSection>
          ) : null}

          {step === 2 ? (
            <InductiveFormSection
              title="Langkah 2: Identitas Gereja &amp; Kontak"
              description="Isi pelindung, pastor paroki, dan informasi kontak resmi (opsional)."
            >
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-pelindung"
                    className="text-xs font-semibold text-muted"
                  >
                    Santo/Santa Pelindung
                  </label>
                  <input
                    id="wiz-pelindung"
                    value={pelindung}
                    onChange={(e) => setPelindung(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-pastor"
                    className="text-xs font-semibold text-muted"
                  >
                    Pastor Paroki
                  </label>
                  <input
                    id="wiz-pastor"
                    value={pastor}
                    onChange={(e) => setPastor(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-telepon"
                    className="text-xs font-semibold text-muted"
                  >
                    Telepon
                  </label>
                  <input
                    id="wiz-telepon"
                    value={telepon}
                    onChange={(e) => setTelepon(e.target.value)}
                    placeholder="+62812..."
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-fax"
                    className="text-xs font-semibold text-muted"
                  >
                    Fax
                  </label>
                  <input
                    id="wiz-fax"
                    value={fax}
                    onChange={(e) => setFax(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-email"
                    className="text-xs font-semibold text-muted"
                  >
                    Email
                  </label>
                  <input
                    id="wiz-email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-website"
                    className="text-xs font-semibold text-muted"
                  >
                    Website
                  </label>
                  <input
                    id="wiz-website"
                    value={website}
                    onChange={(e) => setWebsite(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-logo"
                    className="text-xs font-semibold text-muted"
                  >
                    URL Logo
                  </label>
                  <input
                    id="wiz-logo"
                    value={logoUrl}
                    onChange={(e) => setLogoUrl(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-kopsurat"
                    className="text-xs font-semibold text-muted"
                  >
                    Kop Surat
                  </label>
                  <input
                    id="wiz-kopsurat"
                    value={kopSurat}
                    onChange={(e) => setKopSurat(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
              </div>

              <div className="flex flex-col-reverse items-stretch justify-between gap-2 pt-2 sm:flex-row sm:items-center">
                <Button
                  variant="outline"
                  className="w-full sm:w-auto"
                  onClick={() => setStep(1)}
                >
                  ← Kembali
                </Button>
                <div className="flex flex-col-reverse items-stretch gap-2 sm:flex-row sm:items-center">
                  <Button
                    variant="outline"
                    className="w-full sm:w-auto"
                    disabled={createMut.isPending}
                    onClick={submitCreate}
                  >
                    Simpan &amp; Buat Paroki
                  </Button>
                  <Button
                    className="w-full sm:w-auto"
                    onClick={() => setStep(3)}
                  >
                    Lanjut ke Domisili →
                  </Button>
                </div>
              </div>
            </InductiveFormSection>
          ) : null}

          {step === 3 ? (
            <InductiveFormSection
              title="Langkah 3: Domisili Wilayah Sipil, Zona Waktu &amp; Koordinat"
              description="Atur zona waktu, wilayah administratif sipil, alamat, dan koordinat paroki."
            >
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-timezone"
                    className="text-xs font-semibold text-muted"
                  >
                    Zona Waktu
                  </label>
                  <select
                    id="wiz-timezone"
                    value={timezone}
                    onChange={(e) => setTimezone(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  >
                    {PAROKI_TIMEZONES.map((tz) => (
                      <option key={tz.value} value={tz.value}>
                        {tz.label}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-kota"
                    className="text-xs font-semibold text-muted"
                  >
                    Kota / Kabupaten (opsional)
                  </label>
                  <select
                    id="wiz-kota"
                    value={kotaId}
                    onChange={(e) => {
                      setKotaId(e.target.value);
                      setKecamatanId("");
                      setKelurahanId("");
                    }}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  >
                    <option value="">Pilih kota/kabupaten…</option>
                    {kotaList.map((k) => (
                      <option key={k.id} value={k.id}>
                        {k.nama}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-kecamatan"
                    className="text-xs font-semibold text-muted"
                  >
                    Kecamatan (opsional)
                  </label>
                  <select
                    id="wiz-kecamatan"
                    disabled={!kotaId}
                    value={kecamatanId}
                    onChange={(e) => {
                      setKecamatanId(e.target.value);
                      setKelurahanId("");
                    }}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg disabled:opacity-50 focus:border-primary focus:outline-none sm:text-sm"
                  >
                    <option value="">Pilih kecamatan…</option>
                    {kecList.map((k) => (
                      <option key={k.id} value={k.id}>
                        {k.nama}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-kelurahan"
                    className="text-xs font-semibold text-muted"
                  >
                    Kelurahan / Desa (opsional)
                  </label>
                  <select
                    id="wiz-kelurahan"
                    disabled={!kecamatanId}
                    value={kelurahanId}
                    onChange={(e) => setKelurahanId(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg disabled:opacity-50 focus:border-primary focus:outline-none sm:text-sm"
                  >
                    <option value="">Pilih kelurahan/desa…</option>
                    {kelList.map((k) => (
                      <option key={k.id} value={k.id}>
                        {k.nama}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor="wiz-alamat"
                  className="text-xs font-semibold text-muted"
                >
                  Alamat Lengkap
                </label>
                <textarea
                  id="wiz-alamat"
                  rows={2}
                  value={alamat}
                  onChange={(e) => setAlamat(e.target.value)}
                  className="rounded-lg border border-line bg-canvas p-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                />
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-lat"
                    className="text-xs font-semibold text-muted"
                  >
                    Latitude (opsional)
                  </label>
                  <input
                    id="wiz-lat"
                    type="number"
                    step="any"
                    value={latitudeStr}
                    onChange={(e) => setLatitudeStr(e.target.value)}
                    placeholder="-6.1695"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor="wiz-lng"
                    className="text-xs font-semibold text-muted"
                  >
                    Longitude (opsional)
                  </label>
                  <input
                    id="wiz-lng"
                    type="number"
                    step="any"
                    value={longitudeStr}
                    onChange={(e) => setLongitudeStr(e.target.value)}
                    placeholder="106.8331"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                  />
                </div>
              </div>

              <div className="flex flex-col-reverse items-stretch justify-between gap-2 pt-2 sm:flex-row sm:items-center">
                <Button
                  variant="outline"
                  className="w-full sm:w-auto"
                  onClick={() => setStep(2)}
                >
                  ← Kembali
                </Button>
                <Button
                  className="w-full sm:w-auto"
                  disabled={createMut.isPending}
                  onClick={submitCreate}
                >
                  {createMut.isPending ? "Menyimpan…" : "Simpan & Buat Paroki"}
                </Button>
              </div>
            </InductiveFormSection>
          ) : null}
        </>
      ) : null}
    </div>
  );
}
