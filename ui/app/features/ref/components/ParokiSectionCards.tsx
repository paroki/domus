import { useEffect, useState } from "react";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  useKecamatanSearch,
  useKelurahanSearch,
  useKeuskupanSearch,
  useKevikepanSearch,
  useKotaSearch,
  useLookupsSearch,
  useParokiSearch,
} from "../api/queries";
import type {
  Paroki,
  UpdateParokiProfileRequest,
  UpdateParokiRequest,
} from "../api/types";
import {
  DEFAULT_PAROKI_TIMEZONE,
  formatParokiTimezone,
  NIL_UUID,
  PAROKI_TIMEZONES,
  validateCoordinatePair,
} from "../model";

export type ParokiSectionKey = "struktur" | "identitas" | "domisili" | "kontak";

function ReadField({ label, value }: { label: string; value?: string | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs font-medium text-muted">{label}</span>
      <span className="text-sm text-fg">{value || "—"}</span>
    </div>
  );
}

export function ParokiStrukturReadOnlyCard({ paroki }: { paroki: Paroki }) {
  const keuskupanQuery = useKeuskupanSearch({});
  const kevikepanQuery = useKevikepanSearch(
    { filter: { keuskupanId: paroki.keuskupanId } },
    Boolean(paroki.kevikepanId),
  );
  const parokiQuery = useParokiSearch(
    { filter: { keuskupanId: paroki.keuskupanId } },
    Boolean(paroki.indukId),
  );
  const statusQuery = useLookupsSearch({
    filter: { category: "status_paroki", scope: "all" },
  });

  const keuskupanName = keuskupanQuery.data?.pages
    .flatMap((p) => p.data ?? [])
    .find((k) => k.id === paroki.keuskupanId)?.nama;
  const kevikepanName = kevikepanQuery.data?.pages
    .flatMap((p) => p.data ?? [])
    .find((k) => k.id === paroki.kevikepanId)?.nama;
  const indukName = parokiQuery.data?.pages
    .flatMap((p) => p.data ?? [])
    .find((p) => p.id === paroki.indukId)?.nama;
  const statusLabel = statusQuery.data?.pages
    .flatMap((p) => p.data ?? [])
    .find((s) => s.code === paroki.statusParoki)?.label;

  return (
    <Card className="flex flex-col gap-4 p-6">
      <div className="flex items-center justify-between gap-2">
        <h2 className="text-base font-semibold text-fg">Struktur Gerejawi</h2>
        <Badge tone={paroki.active ? "success" : "neutral"}>
          {paroki.active ? "Aktif" : "Nonaktif"}
        </Badge>
      </div>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <ReadField
          label="Keuskupan"
          value={keuskupanName ?? paroki.keuskupanId}
        />
        <ReadField label="Kevikepan" value={kevikepanName} />
        <ReadField label="Paroki Induk" value={indukName} />
        <ReadField label="Kode" value={paroki.kode} />
        <ReadField label="Kode Resmi" value={paroki.kodeResmi} />
        <ReadField
          label="Status Paroki"
          value={statusLabel ?? paroki.statusParoki}
        />
      </div>
    </Card>
  );
}

export function ParokiStrukturCard({
  paroki,
  editing,
  pending,
  onEdit,
  onCancel,
  onSave,
}: {
  paroki: Paroki;
  editing: boolean;
  pending: boolean;
  onEdit: () => void;
  onCancel: () => void;
  onSave: (body: UpdateParokiRequest) => Promise<void>;
}) {
  const [keuskupanId, setKeuskupanId] = useState(paroki.keuskupanId ?? "");
  const [kevikepanId, setKevikepanId] = useState(paroki.kevikepanId ?? "");
  const [indukId, setIndukId] = useState(paroki.indukId ?? "");
  const [kode, setKode] = useState(paroki.kode ?? "");
  const [kodeResmi, setKodeResmi] = useState(paroki.kodeResmi ?? "");
  const [statusParoki, setStatusParoki] = useState(paroki.statusParoki ?? "");
  const [active, setActive] = useState(paroki.active ?? true);

  useEffect(() => {
    setKeuskupanId(paroki.keuskupanId ?? "");
    setKevikepanId(paroki.kevikepanId ?? "");
    setIndukId(paroki.indukId ?? "");
    setKode(paroki.kode ?? "");
    setKodeResmi(paroki.kodeResmi ?? "");
    setStatusParoki(paroki.statusParoki ?? "");
    setActive(paroki.active ?? true);
  }, [paroki]);

  const keuskupanQuery = useKeuskupanSearch({});
  const keuskupanList =
    keuskupanQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const kevikepanQuery = useKevikepanSearch(
    { filter: keuskupanId ? { keuskupanId } : undefined },
    Boolean(keuskupanId),
  );
  const kevikepanList =
    kevikepanQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const rootParokiQuery = useParokiSearch(
    { filter: keuskupanId ? { keuskupanId } : undefined },
    Boolean(keuskupanId),
  );
  const rootParokiList = (
    rootParokiQuery.data?.pages.flatMap((p) => p.data ?? []) ?? []
  ).filter((p) => p.id !== paroki.id && !p.indukId);
  const statusQuery = useLookupsSearch({
    filter: { category: "status_paroki", scope: "all" },
  });
  const statusOptions =
    statusQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  if (!editing) {
    const keuskupanName = keuskupanList.find(
      (k) => k.id === paroki.keuskupanId,
    )?.nama;
    const kevikepanName = kevikepanList.find(
      (k) => k.id === paroki.kevikepanId,
    )?.nama;
    const indukName = rootParokiList.find((p) => p.id === paroki.indukId)?.nama;
    const statusLabel = statusOptions.find(
      (s) => s.code === paroki.statusParoki,
    )?.label;

    return (
      <Card className="flex flex-col gap-4 p-6">
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <h2 className="text-base font-semibold text-fg">
              Struktur &amp; Status Gerejawi
            </h2>
            <Badge tone={paroki.active ? "success" : "neutral"}>
              {paroki.active ? "Aktif" : "Nonaktif"}
            </Badge>
          </div>
          <Button variant="outline" onClick={onEdit}>
            Ubah
          </Button>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <ReadField label="Keuskupan" value={keuskupanName} />
          <ReadField label="Kevikepan" value={kevikepanName} />
          <ReadField label="Paroki Induk" value={indukName} />
          <ReadField label="Kode" value={paroki.kode} />
          <ReadField label="Kode Resmi" value={paroki.kodeResmi} />
          <ReadField
            label="Status Paroki"
            value={statusLabel ?? paroki.statusParoki}
          />
        </div>
      </Card>
    );
  }

  return (
    <Card className="flex flex-col gap-4 p-6">
      <h2 className="text-base font-semibold text-fg">
        Ubah Struktur &amp; Status Gerejawi
      </h2>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          await onSave({
            keuskupanId,
            kevikepanId: kevikepanId || NIL_UUID,
            indukId: indukId || NIL_UUID,
            kode: kode.trim(),
            kodeResmi: kodeResmi.trim(),
            statusParoki,
            active,
          });
        }}
        className="flex flex-col gap-4"
      >
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-keuskupan"
              className="text-xs font-semibold text-muted"
            >
              Keuskupan
            </label>
            <select
              id="sec-keuskupan"
              required
              value={keuskupanId}
              onChange={(e) => {
                setKeuskupanId(e.target.value);
                setKevikepanId("");
                setIndukId("");
              }}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            >
              {keuskupanList.map((k) => (
                <option key={k.id} value={k.id}>
                  {k.nama}
                </option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-kevikepan"
              className="text-xs font-semibold text-muted"
            >
              Kevikepan
            </label>
            <select
              id="sec-kevikepan"
              value={kevikepanId}
              onChange={(e) => setKevikepanId(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            >
              <option value="">Tanpa kevikepan</option>
              {kevikepanList.map((k) => (
                <option key={k.id} value={k.id}>
                  {k.nama}
                </option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-induk"
              className="text-xs font-semibold text-muted"
            >
              Paroki Induk
            </label>
            <select
              id="sec-induk"
              value={indukId}
              onChange={(e) => setIndukId(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
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
              htmlFor="sec-kode"
              className="text-xs font-semibold text-muted"
            >
              Kode Paroki
            </label>
            <input
              id="sec-kode"
              required
              value={kode}
              onChange={(e) => setKode(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 font-mono text-sm text-fg"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-koderesmi"
              className="text-xs font-semibold text-muted"
            >
              Kode Resmi
            </label>
            <input
              id="sec-koderesmi"
              value={kodeResmi}
              onChange={(e) => setKodeResmi(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-status"
              className="text-xs font-semibold text-muted"
            >
              Status Paroki
            </label>
            <select
              id="sec-status"
              value={statusParoki}
              onChange={(e) => setStatusParoki(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            >
              <option value="">Kosongkan</option>
              {statusOptions.map((s) => (
                <option key={s.id} value={s.code}>
                  {s.label}
                </option>
              ))}
            </select>
          </div>
        </div>
        <label className="inline-flex items-center gap-2 text-sm text-fg">
          <input
            type="checkbox"
            checked={active}
            onChange={(e) => setActive(e.target.checked)}
            className="size-4 accent-primary"
          />
          Paroki Aktif
        </label>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onCancel}>
            Batal
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Menyimpan…" : "Simpan Perubahan"}
          </Button>
        </div>
      </form>
    </Card>
  );
}

export function ParokiIdentitasCard({
  paroki,
  editing,
  pending,
  onEdit,
  onCancel,
  onSave,
}: {
  paroki: Paroki;
  editing: boolean;
  pending: boolean;
  onEdit: () => void;
  onCancel: () => void;
  onSave: (body: UpdateParokiProfileRequest) => Promise<void>;
}) {
  const [nama, setNama] = useState(paroki.nama ?? "");
  const [pelindung, setPelindung] = useState(paroki.pelindung ?? "");
  const [pastor, setPastor] = useState(paroki.pastor ?? "");
  const [logoUrl, setLogoUrl] = useState(paroki.logoUrl ?? "");
  const [kopSurat, setKopSurat] = useState(paroki.kopSurat ?? "");

  useEffect(() => {
    setNama(paroki.nama ?? "");
    setPelindung(paroki.pelindung ?? "");
    setPastor(paroki.pastor ?? "");
    setLogoUrl(paroki.logoUrl ?? "");
    setKopSurat(paroki.kopSurat ?? "");
  }, [paroki]);

  if (!editing) {
    return (
      <Card className="flex flex-col gap-4 p-6">
        <div className="flex items-center justify-between gap-2">
          <h2 className="text-base font-semibold text-fg">
            Identitas Gereja &amp; Kop Surat
          </h2>
          <Button variant="outline" onClick={onEdit}>
            Ubah
          </Button>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <ReadField label="Nama Paroki" value={paroki.nama} />
          <ReadField label="Pelindung" value={paroki.pelindung} />
          <ReadField label="Pastor Paroki" value={paroki.pastor} />
          <ReadField label="URL Logo" value={paroki.logoUrl} />
          <ReadField label="Kop Surat" value={paroki.kopSurat} />
        </div>
      </Card>
    );
  }

  return (
    <Card className="flex flex-col gap-4 p-6">
      <h2 className="text-base font-semibold text-fg">
        Ubah Identitas Gereja &amp; Kop Surat
      </h2>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          await onSave({
            nama: nama.trim(),
            pelindung: pelindung.trim(),
            pastor: pastor.trim(),
            logoUrl: logoUrl.trim(),
            kopSurat: kopSurat.trim(),
          });
        }}
        className="flex flex-col gap-4"
      >
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-nama"
              className="text-xs font-semibold text-muted"
            >
              Nama Paroki
            </label>
            <input
              id="sec-nama"
              required
              value={nama}
              onChange={(e) => setNama(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-pelindung"
              className="text-xs font-semibold text-muted"
            >
              Pelindung
            </label>
            <input
              id="sec-pelindung"
              value={pelindung}
              onChange={(e) => setPelindung(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-pastor"
              className="text-xs font-semibold text-muted"
            >
              Pastor Paroki
            </label>
            <input
              id="sec-pastor"
              value={pastor}
              onChange={(e) => setPastor(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-logo"
              className="text-xs font-semibold text-muted"
            >
              URL Logo
            </label>
            <input
              id="sec-logo"
              value={logoUrl}
              onChange={(e) => setLogoUrl(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
        </div>
        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="sec-kopsurat"
            className="text-xs font-semibold text-muted"
          >
            Kop Surat
          </label>
          <input
            id="sec-kopsurat"
            value={kopSurat}
            onChange={(e) => setKopSurat(e.target.value)}
            className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
          />
        </div>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onCancel}>
            Batal
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Menyimpan…" : "Simpan Perubahan"}
          </Button>
        </div>
      </form>
    </Card>
  );
}

export function ParokiDomisiliCard({
  paroki,
  editing,
  pending,
  onEdit,
  onCancel,
  onSave,
}: {
  paroki: Paroki;
  editing: boolean;
  pending: boolean;
  onEdit: () => void;
  onCancel: () => void;
  onSave: (body: UpdateParokiProfileRequest) => Promise<void>;
}) {
  const [timezone, setTimezone] = useState(
    paroki.timezone ?? DEFAULT_PAROKI_TIMEZONE,
  );
  const [kotaId, setKotaId] = useState(paroki.kotaId ?? "");
  const [kecamatanId, setKecamatanId] = useState(paroki.kecamatanId ?? "");
  const [kelurahanId, setKelurahanId] = useState(paroki.kelurahanId ?? "");
  const [alamat, setAlamat] = useState(paroki.alamat ?? "");
  const [latStr, setLatStr] = useState(
    paroki.latitude != null ? String(paroki.latitude) : "",
  );
  const [lngStr, setLngStr] = useState(
    paroki.longitude != null ? String(paroki.longitude) : "",
  );
  const [coordError, setCoordError] = useState<string | null>(null);

  useEffect(() => {
    setTimezone(paroki.timezone ?? DEFAULT_PAROKI_TIMEZONE);
    setKotaId(paroki.kotaId ?? "");
    setKecamatanId(paroki.kecamatanId ?? "");
    setKelurahanId(paroki.kelurahanId ?? "");
    setAlamat(paroki.alamat ?? "");
    setLatStr(paroki.latitude != null ? String(paroki.latitude) : "");
    setLngStr(paroki.longitude != null ? String(paroki.longitude) : "");
    setCoordError(null);
  }, [paroki]);

  const kotaQuery = useKotaSearch({});
  const kotaList = kotaQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const kecQuery = useKecamatanSearch(
    { filter: kotaId ? { kotaId } : undefined },
    Boolean(kotaId),
  );
  const kecList = kecQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const kelQuery = useKelurahanSearch(
    { filter: kecamatanId ? { kecamatanId } : undefined },
    Boolean(kecamatanId),
  );
  const kelList = kelQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  if (!editing) {
    const kotaName = kotaList.find((k) => k.id === paroki.kotaId)?.nama;
    const kecName = kecList.find((k) => k.id === paroki.kecamatanId)?.nama;
    const kelName = kelList.find((k) => k.id === paroki.kelurahanId)?.nama;
    const coordStr =
      paroki.latitude != null && paroki.longitude != null
        ? `${paroki.latitude}, ${paroki.longitude}`
        : undefined;

    return (
      <Card className="flex flex-col gap-4 p-6">
        <div className="flex items-center justify-between gap-2">
          <h2 className="text-base font-semibold text-fg">
            Domisili Wilayah Sipil, Zona Waktu &amp; Koordinat
          </h2>
          <Button variant="outline" onClick={onEdit}>
            Ubah
          </Button>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <ReadField
            label="Zona Waktu"
            value={`${formatParokiTimezone(paroki.timezone)} (${paroki.timezone ?? DEFAULT_PAROKI_TIMEZONE})`}
          />
          <ReadField label="Kota / Kabupaten" value={kotaName} />
          <ReadField label="Kecamatan" value={kecName} />
          <ReadField label="Kelurahan / Desa" value={kelName} />
          <ReadField label="Alamat" value={paroki.alamat} />
          <ReadField label="Koordinat" value={coordStr} />
        </div>
      </Card>
    );
  }

  return (
    <Card className="flex flex-col gap-4 p-6">
      <h2 className="text-base font-semibold text-fg">
        Ubah Domisili Wilayah Sipil, Zona Waktu &amp; Koordinat
      </h2>
      {coordError ? (
        <div
          role="alert"
          className="rounded-lg border border-danger-fg/30 bg-danger-bg px-3 py-2 text-xs text-danger-fg"
        >
          {coordError}
        </div>
      ) : null}
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          const coords = validateCoordinatePair(latStr, lngStr);
          if (!coords.valid) {
            setCoordError(coords.error ?? "Koordinat tidak valid.");
            return;
          }
          setCoordError(null);
          await onSave({
            timezone,
            kotaId: kotaId || NIL_UUID,
            kecamatanId: kecamatanId || NIL_UUID,
            kelurahanId: kelurahanId || NIL_UUID,
            alamat: alamat.trim(),
            ...(coords.latitude != null && coords.longitude != null
              ? { latitude: coords.latitude, longitude: coords.longitude }
              : paroki.latitude != null || paroki.longitude != null
                ? { clearCoordinates: true }
                : {}),
          });
        }}
        className="flex flex-col gap-4"
      >
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-timezone"
              className="text-xs font-semibold text-muted"
            >
              Zona Waktu
            </label>
            <select
              id="sec-timezone"
              value={timezone}
              onChange={(e) => setTimezone(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
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
              htmlFor="sec-kota"
              className="text-xs font-semibold text-muted"
            >
              Kota / Kabupaten
            </label>
            <select
              id="sec-kota"
              value={kotaId}
              onChange={(e) => {
                setKotaId(e.target.value);
                setKecamatanId("");
                setKelurahanId("");
              }}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            >
              <option value="">Kosongkan</option>
              {kotaList.map((k) => (
                <option key={k.id} value={k.id}>
                  {k.nama}
                </option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-kecamatan"
              className="text-xs font-semibold text-muted"
            >
              Kecamatan
            </label>
            <select
              id="sec-kecamatan"
              disabled={!kotaId}
              value={kecamatanId}
              onChange={(e) => {
                setKecamatanId(e.target.value);
                setKelurahanId("");
              }}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg disabled:opacity-50"
            >
              <option value="">Kosongkan</option>
              {kecList.map((k) => (
                <option key={k.id} value={k.id}>
                  {k.nama}
                </option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-kelurahan"
              className="text-xs font-semibold text-muted"
            >
              Kelurahan / Desa
            </label>
            <select
              id="sec-kelurahan"
              disabled={!kecamatanId}
              value={kelurahanId}
              onChange={(e) => setKelurahanId(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg disabled:opacity-50"
            >
              <option value="">Kosongkan</option>
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
            htmlFor="sec-alamat"
            className="text-xs font-semibold text-muted"
          >
            Alamat
          </label>
          <textarea
            id="sec-alamat"
            rows={2}
            value={alamat}
            onChange={(e) => setAlamat(e.target.value)}
            className="rounded-lg border border-line bg-canvas p-3 text-sm text-fg"
          />
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-lat"
              className="text-xs font-semibold text-muted"
            >
              Latitude
            </label>
            <input
              id="sec-lat"
              type="number"
              step="any"
              value={latStr}
              onChange={(e) => setLatStr(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-lng"
              className="text-xs font-semibold text-muted"
            >
              Longitude
            </label>
            <input
              id="sec-lng"
              type="number"
              step="any"
              value={lngStr}
              onChange={(e) => setLngStr(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
        </div>
        <div className="flex flex-wrap items-center justify-between gap-2">
          {paroki.latitude != null || paroki.longitude != null ? (
            <Button
              variant="ghost"
              disabled={pending}
              onClick={async () => {
                setLatStr("");
                setLngStr("");
                await onSave({ clearCoordinates: true });
              }}
            >
              Kosongkan Koordinat
            </Button>
          ) : (
            <span />
          )}
          <div className="flex gap-2">
            <Button variant="outline" onClick={onCancel}>
              Batal
            </Button>
            <Button type="submit" disabled={pending}>
              {pending ? "Menyimpan…" : "Simpan Perubahan"}
            </Button>
          </div>
        </div>
      </form>
    </Card>
  );
}

export function ParokiKontakCard({
  paroki,
  editing,
  pending,
  onEdit,
  onCancel,
  onSave,
}: {
  paroki: Paroki;
  editing: boolean;
  pending: boolean;
  onEdit: () => void;
  onCancel: () => void;
  onSave: (body: UpdateParokiProfileRequest) => Promise<void>;
}) {
  const [telepon, setTelepon] = useState(paroki.telepon ?? "");
  const [fax, setFax] = useState(paroki.fax ?? "");
  const [email, setEmail] = useState(paroki.email ?? "");
  const [website, setWebsite] = useState(paroki.website ?? "");

  useEffect(() => {
    setTelepon(paroki.telepon ?? "");
    setFax(paroki.fax ?? "");
    setEmail(paroki.email ?? "");
    setWebsite(paroki.website ?? "");
  }, [paroki]);

  if (!editing) {
    return (
      <Card className="flex flex-col gap-4 p-6">
        <div className="flex items-center justify-between gap-2">
          <h2 className="text-base font-semibold text-fg">Kontak Resmi</h2>
          <Button variant="outline" onClick={onEdit}>
            Ubah
          </Button>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <ReadField label="Telepon" value={paroki.telepon} />
          <ReadField label="Fax" value={paroki.fax} />
          <ReadField label="Email" value={paroki.email} />
          <ReadField label="Website" value={paroki.website} />
        </div>
      </Card>
    );
  }

  return (
    <Card className="flex flex-col gap-4 p-6">
      <h2 className="text-base font-semibold text-fg">Ubah Kontak Resmi</h2>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          await onSave({
            telepon: telepon.trim(),
            fax: fax.trim(),
            email: email.trim(),
            website: website.trim(),
          });
        }}
        className="flex flex-col gap-4"
      >
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-telepon"
              className="text-xs font-semibold text-muted"
            >
              Telepon
            </label>
            <input
              id="sec-telepon"
              value={telepon}
              onChange={(e) => setTelepon(e.target.value)}
              placeholder="+62812..."
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-fax"
              className="text-xs font-semibold text-muted"
            >
              Fax
            </label>
            <input
              id="sec-fax"
              value={fax}
              onChange={(e) => setFax(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-email"
              className="text-xs font-semibold text-muted"
            >
              Email
            </label>
            <input
              id="sec-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="sec-website"
              className="text-xs font-semibold text-muted"
            >
              Website
            </label>
            <input
              id="sec-website"
              value={website}
              onChange={(e) => setWebsite(e.target.value)}
              className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            />
          </div>
        </div>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onCancel}>
            Batal
          </Button>
          <Button type="submit" disabled={pending}>
            {pending ? "Menyimpan…" : "Simpan Perubahan"}
          </Button>
        </div>
      </form>
    </Card>
  );
}
