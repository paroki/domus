import { Pencil, Plus, Trash2, Users } from "lucide-react";
import { type FormEvent, useState } from "react";
import { Link } from "react-router";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import { InductiveDangerZone } from "~/shared/ui/iui";
import { useDeleteKeluarga, useUpdateKeluarga } from "../api/actions";
import { resolveUmatErrorMessage } from "../api/errors";
import {
  useUmatKecamatanList,
  useUmatKelurahanList,
  useUmatKotaList,
  useUmatLingkunganList,
  useUmatLookups,
  useUmatWilayahList,
} from "../api/queries";
import type { KeluargaDetail } from "../api/types";
import {
  genderLabel,
  optionalStringForPatch,
  optionalUuidForPatch,
  validatePIIInput,
} from "../model";

export type KeluargaSectionKey = "identity" | "address" | "household";

function DetailRow({ label, value }: { label: string; value?: string | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-xs font-medium text-muted">{label}</dt>
      <dd className="text-sm font-medium text-fg">{value || "-"}</dd>
    </div>
  );
}

export function KeluargaSectionCards({
  parokiId,
  keluarga,
  onFeedback,
  onDeleted,
}: {
  parokiId: string;
  keluarga: KeluargaDetail;
  onFeedback: (
    fb: { tone: "success" | "error"; message: string } | null,
  ) => void;
  onDeleted: () => void;
}) {
  const [editing, setEditing] = useState<KeluargaSectionKey | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  const updateKeluarga = useUpdateKeluarga(parokiId);
  const deleteKeluarga = useDeleteKeluarga(parokiId);

  const wilayahQuery = useUmatWilayahList(parokiId);
  const allLingkunganQuery = useUmatLingkunganList(parokiId);
  const kotaQuery = useUmatKotaList();
  const statusRumahQuery = useUmatLookups(parokiId, "status_rumah");
  const statusEkonomiQuery = useUmatLookups(parokiId, "status_ekonomi");

  // Identity section state
  const [namaKeluarga, setNamaKeluarga] = useState("");
  const [nomorKk, setNomorKk] = useState("");
  const [wilayahId, setWilayahId] = useState("");
  const [lingkunganId, setLingkunganId] = useState("");
  const [aktif, setAktif] = useState(true);

  // Address section state
  const [alamat, setAlamat] = useState("");
  const [kotaId, setKotaId] = useState("");
  const [kecamatanId, setKecamatanId] = useState("");
  const [kelurahanId, setKelurahanId] = useState("");
  const [kodePos, setKodePos] = useState("");

  // Household section state
  const [statusRumah, setStatusRumah] = useState("");
  const [statusEkonomi, setStatusEkonomi] = useState("");
  const [catatan, setCatatan] = useState("");

  const activeKotaId = editing === "address" ? kotaId : (keluarga.kotaId ?? "");
  const activeKecamatanId =
    editing === "address" ? kecamatanId : (keluarga.kecamatanId ?? "");
  const kecamatanQuery = useUmatKecamatanList(
    activeKotaId,
    Boolean(activeKotaId),
  );
  const kelurahanQuery = useUmatKelurahanList(
    activeKecamatanId,
    Boolean(activeKecamatanId),
  );

  const wilayahList = wilayahQuery.data ?? [];
  const allLingkunganList = allLingkunganQuery.data ?? [];
  const filteredLingkunganList = wilayahId
    ? allLingkunganList.filter((l) => l.wilayahId === wilayahId)
    : allLingkunganList;
  const kotaList = kotaQuery.data ?? [];
  const kecamatanList = kecamatanQuery.data ?? [];
  const kelurahanList = kelurahanQuery.data ?? [];
  const statusRumahList = statusRumahQuery.data ?? [];
  const statusEkonomiList = statusEkonomiQuery.data ?? [];

  const wilayahName =
    wilayahList.find((w) => w.id === keluarga.wilayahId)?.nama ?? null;
  const lingkunganName =
    allLingkunganList.find((l) => l.id === keluarga.lingkunganId)?.nama ?? null;
  const kotaName = kotaList.find((k) => k.id === keluarga.kotaId)?.nama ?? null;
  const kecamatanName =
    kecamatanList.find((k) => k.id === keluarga.kecamatanId)?.nama ?? null;
  const kelurahanName =
    kelurahanList.find((k) => k.id === keluarga.kelurahanId)?.nama ?? null;
  const statusRumahLabel =
    statusRumahList.find((l) => l.code === keluarga.statusRumah)?.label ??
    keluarga.statusRumah ??
    null;
  const statusEkonomiLabel =
    statusEkonomiList.find((l) => l.code === keluarga.statusEkonomi)?.label ??
    keluarga.statusEkonomi ??
    null;

  function startEdit(section: KeluargaSectionKey) {
    onFeedback(null);
    if (section === "identity") {
      setNamaKeluarga(keluarga.namaKeluarga ?? "");
      setNomorKk(keluarga.nomorKk ?? "");
      setWilayahId(keluarga.wilayahId ?? "");
      setLingkunganId(keluarga.lingkunganId ?? "");
      setAktif(keluarga.aktif ?? true);
    } else if (section === "address") {
      setAlamat(keluarga.alamat ?? "");
      setKotaId(keluarga.kotaId ?? "");
      setKecamatanId(keluarga.kecamatanId ?? "");
      setKelurahanId(keluarga.kelurahanId ?? "");
      setKodePos(keluarga.kodePos ?? "");
    } else if (section === "household") {
      setStatusRumah(keluarga.statusRumah ?? "");
      setStatusEkonomi(keluarga.statusEkonomi ?? "");
      setCatatan(keluarga.catatan ?? "");
    }
    setEditing(section);
  }

  async function saveIdentity(e: FormEvent) {
    e.preventDefault();
    if (!keluarga.id) return;
    const kkErr = validatePIIInput(nomorKk, "Nomor KK");
    if (kkErr) {
      onFeedback({ tone: "error", message: kkErr });
      return;
    }
    onFeedback(null);
    try {
      await updateKeluarga.mutateAsync({
        id: keluarga.id,
        body: {
          namaKeluarga: namaKeluarga.trim(),
          nomorKk: optionalStringForPatch(nomorKk),
          wilayahId: optionalUuidForPatch(wilayahId),
          lingkunganId: optionalUuidForPatch(lingkunganId),
          aktif,
        },
      });
      setEditing(null);
      onFeedback({
        tone: "success",
        message: "Identitas Kartu Keluarga berhasil diperbarui.",
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  async function saveAddress(e: FormEvent) {
    e.preventDefault();
    if (!keluarga.id) return;
    onFeedback(null);
    try {
      await updateKeluarga.mutateAsync({
        id: keluarga.id,
        body: {
          alamat: optionalStringForPatch(alamat),
          kotaId: optionalUuidForPatch(kotaId),
          kecamatanId: optionalUuidForPatch(kecamatanId),
          kelurahanId: optionalUuidForPatch(kelurahanId),
          kodePos: optionalStringForPatch(kodePos),
        },
      });
      setEditing(null);
      onFeedback({
        tone: "success",
        message: "Alamat domisili Kartu Keluarga berhasil diperbarui.",
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  async function saveHousehold(e: FormEvent) {
    e.preventDefault();
    if (!keluarga.id) return;
    onFeedback(null);
    try {
      await updateKeluarga.mutateAsync({
        id: keluarga.id,
        body: {
          statusRumah: optionalStringForPatch(statusRumah),
          statusEkonomi: optionalStringForPatch(statusEkonomi),
          catatan: optionalStringForPatch(catatan),
        },
      });
      setEditing(null);
      onFeedback({
        tone: "success",
        message: "Kondisi rumah tangga berhasil diperbarui.",
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  async function handleConfirmDelete() {
    if (!keluarga.id) return;
    setConfirmDelete(false);
    onFeedback(null);
    try {
      await deleteKeluarga.mutateAsync(keluarga.id);
      onDeleted();
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  const anggota = keluarga.anggota ?? [];
  const addMemberParams = new URLSearchParams();
  if (keluarga.id) addMemberParams.set("keluargaId", keluarga.id);
  if (keluarga.wilayahId) addMemberParams.set("wilayahId", keluarga.wilayahId);
  if (keluarga.lingkunganId) {
    addMemberParams.set("lingkunganId", keluarga.lingkunganId);
  }

  return (
    <div className="flex flex-col gap-5">
      {/* 1. Identitas Kartu Keluarga & Teritorial Gerejawi */}
      <Card className="p-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-base font-semibold text-fg">
              Identitas Kartu Keluarga & Teritorial Gerejawi
            </h2>
            <p className="mt-0.5 text-xs text-muted">
              Nomor KK, nama kepala/keluarga, dan wilayah pelayanan paroki.
            </p>
          </div>
          {editing !== "identity" ? (
            <Button
              variant="outline"
              className="h-9 px-3 text-xs"
              onClick={() => startEdit("identity")}
            >
              <Pencil aria-hidden className="size-3.5" />
              Ubah
            </Button>
          ) : null}
        </div>

        {editing === "identity" ? (
          <form onSubmit={saveIdentity} className="mt-5 flex flex-col gap-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Nama Keluarga *
                <input
                  type="text"
                  required
                  value={namaKeluarga}
                  onChange={(e) => setNamaKeluarga(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Nomor KK
                <input
                  type="text"
                  value={nomorKk}
                  onChange={(e) => setNomorKk(e.target.value)}
                  placeholder="4–32 karakter (kosongkan untuk hapus)"
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
                      allLingkunganList.find((l) => l.id === lingkunganId)
                        ?.wilayahId !== nextWil
                    ) {
                      setLingkunganId("");
                    }
                  }}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  <option value="">-- Tidak dipilih --</option>
                  {wilayahList.map((w) => (
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
                    const found = allLingkunganList.find(
                      (l) => l.id === nextLing,
                    );
                    if (found?.wilayahId && !wilayahId) {
                      setWilayahId(found.wilayahId);
                    }
                  }}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  <option value="">-- Tidak dipilih --</option>
                  {filteredLingkunganList.map((l) => (
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

            <div className="flex items-center justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditing(null)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={updateKeluarga.isPending}>
                Simpan
              </Button>
            </div>
          </form>
        ) : (
          <dl className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div className="flex flex-col gap-0.5">
              <dt className="text-xs font-medium text-muted">
                Nomor Registrasi
              </dt>
              <dd className="font-mono text-sm font-medium text-fg">
                {keluarga.nomorRegistrasi || "-"}
              </dd>
            </div>
            <DetailRow label="Nama Keluarga" value={keluarga.namaKeluarga} />
            <div className="flex flex-col gap-0.5">
              <dt className="text-xs font-medium text-muted">Nomor KK</dt>
              <dd className="font-mono text-sm font-medium text-fg">
                {keluarga.nomorKk || "-"}
              </dd>
            </div>
            <div className="flex flex-col gap-0.5">
              <dt className="text-xs font-medium text-muted">Status</dt>
              <dd>
                <Badge tone={keluarga.aktif ? "success" : "neutral"}>
                  {keluarga.aktif ? "Aktif" : "Nonaktif"}
                </Badge>
              </dd>
            </div>
            <DetailRow label="Wilayah" value={wilayahName} />
            <DetailRow label="Lingkungan" value={lingkunganName} />
          </dl>
        )}
      </Card>

      {/* 2. Alamat Domisili & Wilayah Sipil */}
      <Card className="p-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-base font-semibold text-fg">
              Alamat Domisili & Wilayah Sipil
            </h2>
            <p className="mt-0.5 text-xs text-muted">
              Alamat tempat tinggal dan pembagian administratif sipil.
            </p>
          </div>
          {editing !== "address" ? (
            <Button
              variant="outline"
              className="h-9 px-3 text-xs"
              onClick={() => startEdit("address")}
            >
              <Pencil aria-hidden className="size-3.5" />
              Ubah
            </Button>
          ) : null}
        </div>

        {editing === "address" ? (
          <form onSubmit={saveAddress} className="mt-5 flex flex-col gap-4">
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
                  <option value="">-- Tidak dipilih --</option>
                  {kotaList.map((k) => (
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
                  <option value="">-- Tidak dipilih --</option>
                  {kecamatanList.map((k) => (
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
                    const found = kelurahanList.find((k) => k.id === nextKel);
                    if (found?.kodePos && !kodePos.trim()) {
                      setKodePos(found.kodePos);
                    }
                  }}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg disabled:opacity-60"
                >
                  <option value="">-- Tidak dipilih --</option>
                  {kelurahanList.map((k) => (
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
            <div className="flex items-center justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditing(null)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={updateKeluarga.isPending}>
                Simpan
              </Button>
            </div>
          </form>
        ) : (
          <dl className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div className="sm:col-span-2 lg:col-span-3">
              <DetailRow label="Alamat" value={keluarga.alamat} />
            </div>
            <DetailRow label="Kota / Kabupaten" value={kotaName} />
            <DetailRow label="Kecamatan" value={kecamatanName} />
            <DetailRow label="Kelurahan / Desa" value={kelurahanName} />
            <DetailRow label="Kode Pos" value={keluarga.kodePos} />
          </dl>
        )}
      </Card>

      {/* 3. Kondisi Rumah Tangga & Catatan */}
      <Card className="p-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-base font-semibold text-fg">
              Kondisi Rumah Tangga & Catatan
            </h2>
            <p className="mt-0.5 text-xs text-muted">
              Status tempat tinggal, sosial ekonomi, dan catatan pastoral.
            </p>
          </div>
          {editing !== "household" ? (
            <Button
              variant="outline"
              className="h-9 px-3 text-xs"
              onClick={() => startEdit("household")}
            >
              <Pencil aria-hidden className="size-3.5" />
              Ubah
            </Button>
          ) : null}
        </div>

        {editing === "household" ? (
          <form onSubmit={saveHousehold} className="mt-5 flex flex-col gap-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Status Rumah
                <select
                  value={statusRumah}
                  onChange={(e) => setStatusRumah(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  <option value="">-- Tidak dipilih --</option>
                  {statusRumahList.map((l) => (
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
                  <option value="">-- Tidak dipilih --</option>
                  {statusEkonomiList.map((l) => (
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
            <div className="flex items-center justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditing(null)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={updateKeluarga.isPending}>
                Simpan
              </Button>
            </div>
          </form>
        ) : (
          <dl className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
            <DetailRow label="Status Rumah" value={statusRumahLabel} />
            <DetailRow label="Status Ekonomi" value={statusEkonomiLabel} />
            <div className="sm:col-span-2">
              <DetailRow label="Catatan" value={keluarga.catatan} />
            </div>
          </dl>
        )}
      </Card>

      {/* 4. Anggota Keluarga */}
      <Card className="p-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-base font-semibold text-fg">
              Anggota Keluarga ({anggota.length})
            </h2>
            <p className="mt-0.5 text-xs text-muted">
              Daftar umat yang terdaftar dalam Kartu Keluarga ini.
            </p>
          </div>
          <Link
            to={`/umat/new?${addMemberParams.toString()}`}
            className="no-underline"
          >
            <Button variant="outline" className="h-9 px-3 text-xs">
              <Plus aria-hidden className="size-3.5" />
              Tambah Anggota Keluarga
            </Button>
          </Link>
        </div>

        {anggota.length === 0 ? (
          <div className="mt-4 flex items-center gap-2 rounded-lg border border-dashed border-line p-4 text-sm text-muted">
            <Users aria-hidden className="size-4" />
            Belum ada anggota keluarga terdaftar.
          </div>
        ) : (
          <div className="mt-4 divide-y divide-line rounded-lg border border-line">
            {anggota.map((a) => (
              <Link
                key={a.id}
                to={`/umat/${a.id}`}
                className="flex flex-wrap items-center justify-between gap-3 p-3.5 no-underline transition-colors hover:bg-surface-2/50"
              >
                <div className="flex flex-col gap-0.5">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-semibold text-fg">
                      {a.nama}
                    </span>
                    {a.hubunganKeluarga ? (
                      <Badge tone="neutral">{a.hubunganKeluarga}</Badge>
                    ) : null}
                    {a.almarhum ? <Badge tone="danger">Almarhum</Badge> : null}
                  </div>
                  <div className="flex flex-wrap items-center gap-2 text-xs text-muted">
                    <span>{genderLabel(a.jenisKelamin)}</span>
                    {a.nikMasked ? (
                      <span className="font-mono">{a.nikMasked}</span>
                    ) : null}
                  </div>
                </div>
                <span className="text-xs font-medium text-primary">
                  Detail →
                </span>
              </Link>
            ))}
          </div>
        )}
      </Card>

      {/* 5. Zona Bahaya */}
      <InductiveDangerZone
        title="Hapus Kartu Keluarga"
        description="Kartu Keluarga hanya dapat dihapus jika tidak ada anggota umat aktif yang terhubung."
        action={
          <Button
            variant="danger"
            disabled={deleteKeluarga.isPending}
            onClick={() => setConfirmDelete(true)}
          >
            <Trash2 aria-hidden className="size-4" />
            Hapus KK
          </Button>
        }
      />

      <ConfirmDialog
        open={confirmDelete}
        title="Hapus Kartu Keluarga?"
        description={`Kartu Keluarga "${keluarga.namaKeluarga}" akan dinonaktifkan dan dihapus dari daftar aktif.`}
        confirmLabel="Hapus KK"
        cancelLabel="Batal"
        confirmVariant="danger"
        pending={deleteKeluarga.isPending}
        onConfirm={handleConfirmDelete}
        onCancel={() => setConfirmDelete(false)}
      />
    </div>
  );
}
