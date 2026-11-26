import { Pencil, Plus, Trash2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { Link } from "react-router";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import { InductiveDangerZone } from "~/shared/ui/iui";
import {
  useDeleteSakramen,
  useDeleteUmat,
  useUpdateUmat,
} from "../api/actions";
import { resolveUmatErrorMessage } from "../api/errors";
import {
  useKeluargaSearch,
  useUmatKecamatanList,
  useUmatKelurahanList,
  useUmatKeuskupanList,
  useUmatKotaList,
  useUmatLingkunganList,
  useUmatLookups,
  useUmatParokiList,
  useUmatParokiMembersList,
  useUmatWilayahList,
} from "../api/queries";
import type { UmatDetail } from "../api/types";
import {
  genderLabel,
  optionalStringForPatch,
  optionalUuidForPatch,
  sakramenJenisLabel,
  validatePIIInput,
  validateYYYYMMDD,
} from "../model";

export type UmatSectionKey =
  | "identity"
  | "household"
  | "contact"
  | "education"
  | "userLink";

function DetailRow({ label, value }: { label: string; value?: string | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-xs font-medium text-muted">{label}</dt>
      <dd className="text-sm font-medium text-fg">{value || "-"}</dd>
    </div>
  );
}

export function UmatSectionCards({
  parokiId,
  umat,
  onFeedback,
  onDeleted,
}: {
  parokiId: string;
  umat: UmatDetail;
  onFeedback: (
    fb: { tone: "success" | "error"; message: string } | null,
  ) => void;
  onDeleted: () => void;
}) {
  const [editing, setEditing] = useState<UmatSectionKey | null>(null);
  const [confirmDeleteUmat, setConfirmDeleteUmat] = useState(false);
  const [pendingDeleteSakramen, setPendingDeleteSakramen] = useState<{
    id: string;
    label: string;
  } | null>(null);

  const updateUmat = useUpdateUmat(parokiId);
  const deleteUmat = useDeleteUmat(parokiId);
  const deleteSakramen = useDeleteSakramen(parokiId, umat.id ?? "");

  // Reference data
  const keluargaListQuery = useKeluargaSearch(parokiId, { limit: 100 });
  const wilayahQuery = useUmatWilayahList(parokiId);
  const allLingkunganQuery = useUmatLingkunganList(parokiId);
  const kotaQuery = useUmatKotaList();
  const keuskupanQuery = useUmatKeuskupanList();
  const parokiQuery = useUmatParokiList();
  const membersQuery = useUmatParokiMembersList(parokiId);

  const agamaQuery = useUmatLookups(parokiId, "agama");
  const golDarahQuery = useUmatLookups(parokiId, "golongan_darah");
  const sukuQuery = useUmatLookups(parokiId, "suku");
  const hubunganQuery = useUmatLookups(parokiId, "hubungan_keluarga");
  const statusPernikahanQuery = useUmatLookups(parokiId, "status_pernikahan");
  const pendidikanQuery = useUmatLookups(parokiId, "pendidikan");
  const pekerjaanQuery = useUmatLookups(parokiId, "pekerjaan");
  const profesiQuery = useUmatLookups(parokiId, "profesi");
  const disabilitasQuery = useUmatLookups(parokiId, "disabilitas");
  const aktivitasQuery = useUmatLookups(parokiId, "aktivitas_sosial");

  // 1. Identity state
  const [nama, setNama] = useState("");
  const [namaBaptis, setNamaBaptis] = useState("");
  const [nik, setNik] = useState("");
  const [jenisKelamin, setJenisKelamin] = useState<"pria" | "wanita">("pria");
  const [tempatLahir, setTempatLahir] = useState("");
  const [tanggalLahir, setTanggalLahir] = useState("");
  const [golonganDarah, setGolonganDarah] = useState("");
  const [suku, setSuku] = useState("");
  const [agama, setAgama] = useState("");
  const [kewarganegaraan, setKewarganegaraan] = useState("");
  const [negara, setNegara] = useState("");

  // 2. Household & status state
  const [keluargaId, setKeluargaId] = useState("");
  const [hubunganKeluarga, setHubunganKeluarga] = useState("");
  const [wilayahId, setWilayahId] = useState("");
  const [lingkunganId, setLingkunganId] = useState("");
  const [statusPernikahan, setStatusPernikahan] = useState("");
  const [almarhum, setAlmarhum] = useState(false);
  const [tanggalMeninggal, setTanggalMeninggal] = useState("");

  // 3. Contact & domicile override state
  const [telepon, setTelepon] = useState("");
  const [email, setEmail] = useState("");
  const [fotoUrl, setFotoUrl] = useState("");
  const [alamat, setAlamat] = useState("");
  const [kotaId, setKotaId] = useState("");
  const [kecamatanId, setKecamatanId] = useState("");
  const [kelurahanId, setKelurahanId] = useState("");

  // 4. Education & occupation state
  const [pendidikan, setPendidikan] = useState("");
  const [jurusan, setJurusan] = useState("");
  const [pekerjaan, setPekerjaan] = useState("");
  const [profesi, setProfesi] = useState("");
  const [keterampilan, setKeterampilan] = useState("");
  const [disabilitas, setDisabilitas] = useState("");
  const [aktivitasSosial, setAktivitasSosial] = useState("");

  // 5. User link state
  const [userId, setUserId] = useState("");

  const activeKotaId = editing === "contact" ? kotaId : (umat.kotaId ?? "");
  const activeKecamatanId =
    editing === "contact" ? kecamatanId : (umat.kecamatanId ?? "");
  const kecamatanQuery = useUmatKecamatanList(
    activeKotaId,
    Boolean(activeKotaId),
  );
  const kelurahanQuery = useUmatKelurahanList(
    activeKecamatanId,
    Boolean(activeKecamatanId),
  );

  const keluargaOptions =
    keluargaListQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const wilayahList = wilayahQuery.data ?? [];
  const allLingkunganList = allLingkunganQuery.data ?? [];
  const filteredLingkunganList = wilayahId
    ? allLingkunganList.filter((l) => l.wilayahId === wilayahId)
    : allLingkunganList;
  const kotaList = kotaQuery.data ?? [];
  const kecamatanList = kecamatanQuery.data ?? [];
  const kelurahanList = kelurahanQuery.data ?? [];
  const keuskupanList = keuskupanQuery.data ?? [];
  const parokiList = parokiQuery.data ?? [];
  const memberList = membersQuery.data ?? [];

  const lookupLabel = (
    list: Array<{ code?: string; label?: string }> | undefined,
    code?: string | null,
  ) => {
    if (!code) return null;
    return list?.find((l) => l.code === code)?.label ?? code;
  };

  const linkedMember = umat.userId
    ? memberList.find((m) => m.userId === umat.userId)
    : undefined;

  function startEdit(section: UmatSectionKey) {
    onFeedback(null);
    if (section === "identity") {
      setNama(umat.nama ?? "");
      setNamaBaptis(umat.namaBaptis ?? "");
      setNik(umat.nik ?? "");
      setJenisKelamin((umat.jenisKelamin as "pria" | "wanita") || "pria");
      setTempatLahir(umat.tempatLahir ?? "");
      setTanggalLahir(umat.tanggalLahir ?? "");
      setGolonganDarah(umat.golonganDarah ?? "");
      setSuku(umat.suku ?? "");
      setAgama(umat.agama ?? "");
      setKewarganegaraan(umat.kewarganegaraan ?? "");
      setNegara(umat.negara ?? "");
    } else if (section === "household") {
      setKeluargaId(umat.keluargaId ?? "");
      setHubunganKeluarga(umat.hubunganKeluarga ?? "");
      setWilayahId(umat.wilayahId ?? "");
      setLingkunganId(umat.lingkunganId ?? "");
      setStatusPernikahan(umat.statusPernikahan ?? "");
      setAlmarhum(Boolean(umat.almarhum));
      setTanggalMeninggal(umat.tanggalMeninggal ?? "");
    } else if (section === "contact") {
      setTelepon(umat.telepon ?? "");
      setEmail(umat.email ?? "");
      setFotoUrl(umat.fotoUrl ?? "");
      setAlamat(umat.alamat ?? "");
      setKotaId(umat.kotaId ?? "");
      setKecamatanId(umat.kecamatanId ?? "");
      setKelurahanId(umat.kelurahanId ?? "");
    } else if (section === "education") {
      setPendidikan(umat.pendidikan ?? "");
      setJurusan(umat.jurusan ?? "");
      setPekerjaan(umat.pekerjaan ?? "");
      setProfesi(umat.profesi ?? "");
      setKeterampilan(umat.keterampilan ?? "");
      setDisabilitas(umat.disabilitas ?? "");
      setAktivitasSosial(umat.aktivitasSosial ?? "");
    } else if (section === "userLink") {
      setUserId(umat.userId ?? "");
    }
    setEditing(section);
  }

  async function saveIdentity(e: FormEvent) {
    e.preventDefault();
    if (!umat.id) return;
    const nikErr = validatePIIInput(nik, "NIK");
    if (nikErr) {
      onFeedback({ tone: "error", message: nikErr });
      return;
    }
    const birthErr = validateYYYYMMDD(tanggalLahir, "Tanggal lahir");
    if (birthErr) {
      onFeedback({ tone: "error", message: birthErr });
      return;
    }
    onFeedback(null);
    try {
      await updateUmat.mutateAsync({
        id: umat.id,
        body: {
          nama: nama.trim(),
          namaBaptis: optionalStringForPatch(namaBaptis),
          nik: optionalStringForPatch(nik),
          jenisKelamin,
          tempatLahir: optionalStringForPatch(tempatLahir),
          tanggalLahir: optionalStringForPatch(tanggalLahir),
          golonganDarah: optionalStringForPatch(golonganDarah),
          suku: optionalStringForPatch(suku),
          agama: optionalStringForPatch(agama),
          kewarganegaraan: optionalStringForPatch(kewarganegaraan),
          negara: optionalStringForPatch(negara),
        },
      });
      setEditing(null);
      onFeedback({
        tone: "success",
        message: "Identitas pribadi berhasil diperbarui.",
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  async function saveHousehold(e: FormEvent) {
    e.preventDefault();
    if (!umat.id) return;
    const deathErr = almarhum
      ? validateYYYYMMDD(tanggalMeninggal, "Tanggal meninggal")
      : null;
    if (deathErr) {
      onFeedback({ tone: "error", message: deathErr });
      return;
    }
    onFeedback(null);
    try {
      await updateUmat.mutateAsync({
        id: umat.id,
        body: {
          keluargaId: optionalUuidForPatch(keluargaId),
          hubunganKeluarga: optionalStringForPatch(hubunganKeluarga),
          wilayahId: optionalUuidForPatch(wilayahId),
          lingkunganId: optionalUuidForPatch(lingkunganId),
          statusPernikahan: optionalStringForPatch(statusPernikahan),
          almarhum,
          ...(almarhum
            ? { tanggalMeninggal: optionalStringForPatch(tanggalMeninggal) }
            : {}),
        },
      });
      setEditing(null);
      onFeedback({
        tone: "success",
        message: "Kartu Keluarga & status hidup berhasil diperbarui.",
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  async function saveContact(e: FormEvent) {
    e.preventDefault();
    if (!umat.id) return;
    onFeedback(null);
    try {
      await updateUmat.mutateAsync({
        id: umat.id,
        body: {
          telepon: optionalStringForPatch(telepon),
          email: optionalStringForPatch(email),
          fotoUrl: optionalStringForPatch(fotoUrl),
          alamat: optionalStringForPatch(alamat),
          kotaId: optionalUuidForPatch(kotaId),
          kecamatanId: optionalUuidForPatch(kecamatanId),
          kelurahanId: optionalUuidForPatch(kelurahanId),
        },
      });
      setEditing(null);
      onFeedback({
        tone: "success",
        message: "Kontak & domisili berhasil diperbarui.",
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  async function saveEducation(e: FormEvent) {
    e.preventDefault();
    if (!umat.id) return;
    onFeedback(null);
    try {
      await updateUmat.mutateAsync({
        id: umat.id,
        body: {
          pendidikan: optionalStringForPatch(pendidikan),
          jurusan: optionalStringForPatch(jurusan),
          pekerjaan: optionalStringForPatch(pekerjaan),
          profesi: optionalStringForPatch(profesi),
          keterampilan: optionalStringForPatch(keterampilan),
          disabilitas: optionalStringForPatch(disabilitas),
          aktivitasSosial: optionalStringForPatch(aktivitasSosial),
        },
      });
      setEditing(null);
      onFeedback({
        tone: "success",
        message: "Pendidikan & pekerjaan berhasil diperbarui.",
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  async function saveUserLink(e: FormEvent) {
    e.preventDefault();
    if (!umat.id) return;
    onFeedback(null);
    try {
      await updateUmat.mutateAsync({
        id: umat.id,
        body: { userId: optionalUuidForPatch(userId) },
      });
      setEditing(null);
      onFeedback({
        tone: "success",
        message: "Tautan akun pengguna Domus berhasil diperbarui.",
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  async function handleConfirmDeleteUmat() {
    if (!umat.id) return;
    setConfirmDeleteUmat(false);
    onFeedback(null);
    try {
      await deleteUmat.mutateAsync(umat.id);
      onDeleted();
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  async function handleConfirmDeleteSakramen() {
    if (!pendingDeleteSakramen) return;
    const target = pendingDeleteSakramen;
    setPendingDeleteSakramen(null);
    onFeedback(null);
    try {
      await deleteSakramen.mutateAsync(target.id);
      onFeedback({
        tone: "success",
        message: `Catatan sakramen ${target.label} berhasil dihapus.`,
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveUmatErrorMessage(err) });
    }
  }

  const sakramen = umat.sakramen ?? [];

  return (
    <div className="flex flex-col gap-5">
      {/* 1. Identitas Pribadi & Kependudukan */}
      <Card className="p-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-base font-semibold text-fg">
              Identitas Pribadi & Kependudukan
            </h2>
            <p className="mt-0.5 text-xs text-muted">
              Nama, nama baptis, NIK terenkripsi, kelahiran, dan asal-usul.
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
                  placeholder="4–32 karakter (kosongkan untuk hapus)"
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
                Golongan Darah
                <select
                  value={golonganDarah}
                  onChange={(e) => setGolonganDarah(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  <option value="">-- Tidak dipilih --</option>
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
                  <option value="">-- Tidak dipilih --</option>
                  {(sukuQuery.data ?? []).map((l) => (
                    <option key={l.id} value={l.code}>
                      {l.label}
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
                  <option value="">-- Tidak dipilih --</option>
                  {(agamaQuery.data ?? []).map((l) => (
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
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Negara
                <input
                  type="text"
                  value={negara}
                  onChange={(e) => setNegara(e.target.value)}
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
              <Button type="submit" disabled={updateUmat.isPending}>
                Simpan Perubahan
              </Button>
            </div>
          </form>
        ) : (
          <dl className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <DetailRow label="Nama Lengkap" value={umat.nama} />
            <DetailRow label="Nama Baptis" value={umat.namaBaptis} />
            <div className="flex flex-col gap-0.5">
              <dt className="text-xs font-medium text-muted">NIK</dt>
              <dd className="font-mono text-sm font-medium text-fg">
                {umat.nik || "-"}
              </dd>
            </div>
            <DetailRow
              label="Jenis Kelamin"
              value={genderLabel(umat.jenisKelamin)}
            />
            <DetailRow label="Tempat Lahir" value={umat.tempatLahir} />
            <DetailRow label="Tanggal Lahir" value={umat.tanggalLahir} />
            <DetailRow
              label="Golongan Darah"
              value={lookupLabel(golDarahQuery.data, umat.golonganDarah)}
            />
            <DetailRow
              label="Suku"
              value={lookupLabel(sukuQuery.data, umat.suku)}
            />
            <DetailRow
              label="Agama"
              value={lookupLabel(agamaQuery.data, umat.agama)}
            />
            <DetailRow label="Kewarganegaraan" value={umat.kewarganegaraan} />
            <DetailRow label="Negara" value={umat.negara} />
          </dl>
        )}
      </Card>

      {/* 2. Kartu Keluarga, Lingkungan & Status Hidup */}
      <Card className="p-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-base font-semibold text-fg">
              Kartu Keluarga, Lingkungan & Status Hidup
            </h2>
            <p className="mt-0.5 text-xs text-muted">
              Keanggotaan KK, wilayah pelayanan, status perkawinan, dan status
              hidup.
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
                Kartu Keluarga
                <select
                  value={keluargaId}
                  onChange={(e) => setKeluargaId(e.target.value)}
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
                  <option value="">-- Tidak dipilih --</option>
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
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Status Pernikahan
                <select
                  value={statusPernikahan}
                  onChange={(e) => setStatusPernikahan(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  <option value="">-- Tidak dipilih --</option>
                  {(statusPernikahanQuery.data ?? []).map((l) => (
                    <option key={l.id} value={l.code}>
                      {l.label}
                    </option>
                  ))}
                </select>
              </label>
            </div>
            <div className="flex flex-wrap items-center gap-4">
              <label className="flex items-center gap-2 text-sm font-medium text-fg">
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
                    className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                  />
                </label>
              ) : null}
            </div>
            <div className="flex items-center justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditing(null)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={updateUmat.isPending}>
                Simpan Perubahan
              </Button>
            </div>
          </form>
        ) : (
          <dl className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div className="flex flex-col gap-0.5">
              <dt className="text-xs font-medium text-muted">Kartu Keluarga</dt>
              <dd className="flex flex-wrap items-center gap-2 text-sm font-medium text-fg">
                {umat.keluarga ? (
                  <>
                    <Link
                      to={`/umat/keluarga/${umat.keluarga.id}`}
                      className="text-primary no-underline hover:underline"
                    >
                      {umat.keluarga.namaKeluarga}
                    </Link>
                    {umat.keluarga.nomorKkMasked ? (
                      <Badge tone="neutral">
                        <span className="font-mono">
                          {umat.keluarga.nomorKkMasked}
                        </span>
                      </Badge>
                    ) : null}
                  </>
                ) : (
                  "-"
                )}
              </dd>
            </div>
            <DetailRow
              label="Hubungan Keluarga"
              value={lookupLabel(hubunganQuery.data, umat.hubunganKeluarga)}
            />
            <DetailRow
              label="Wilayah"
              value={wilayahList.find((w) => w.id === umat.wilayahId)?.nama}
            />
            <DetailRow
              label="Lingkungan"
              value={
                allLingkunganList.find((l) => l.id === umat.lingkunganId)?.nama
              }
            />
            <DetailRow
              label="Status Pernikahan"
              value={lookupLabel(
                statusPernikahanQuery.data,
                umat.statusPernikahan,
              )}
            />
            <div className="flex flex-col gap-0.5">
              <dt className="text-xs font-medium text-muted">Status Hidup</dt>
              <dd className="flex items-center gap-2">
                <Badge tone={umat.almarhum ? "danger" : "success"}>
                  {umat.almarhum ? "Almarhum" : "Hidup"}
                </Badge>
                {umat.almarhum && umat.tanggalMeninggal ? (
                  <span className="text-xs text-muted">
                    ({umat.tanggalMeninggal})
                  </span>
                ) : null}
              </dd>
            </div>
          </dl>
        )}
      </Card>

      {/* 3. Kontak & Domisili Override */}
      <Card className="p-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-base font-semibold text-fg">
              Kontak & Domisili Override
            </h2>
            <p className="mt-0.5 text-xs text-muted">
              Nomor telepon, email, dan alamat pribadi jika berbeda dengan KK.
            </p>
          </div>
          {editing !== "contact" ? (
            <Button
              variant="outline"
              className="h-9 px-3 text-xs"
              onClick={() => startEdit("contact")}
            >
              <Pencil aria-hidden className="size-3.5" />
              Ubah
            </Button>
          ) : null}
        </div>

        {editing === "contact" ? (
          <form onSubmit={saveContact} className="mt-5 flex flex-col gap-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Telepon
                <input
                  type="tel"
                  value={telepon}
                  onChange={(e) => setTelepon(e.target.value)}
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
                Alamat Domisili
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
                  onChange={(e) => setKelurahanId(e.target.value)}
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
            </div>
            <div className="flex items-center justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditing(null)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={updateUmat.isPending}>
                Simpan Perubahan
              </Button>
            </div>
          </form>
        ) : (
          <dl className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <DetailRow label="Telepon" value={umat.telepon} />
            <DetailRow label="Email" value={umat.email} />
            <DetailRow label="URL Foto" value={umat.fotoUrl} />
            <div className="sm:col-span-2 lg:col-span-3">
              <DetailRow label="Alamat" value={umat.alamat} />
            </div>
            <DetailRow
              label="Kota / Kabupaten"
              value={kotaList.find((k) => k.id === umat.kotaId)?.nama}
            />
            <DetailRow
              label="Kecamatan"
              value={kecamatanList.find((k) => k.id === umat.kecamatanId)?.nama}
            />
            <DetailRow
              label="Kelurahan / Desa"
              value={kelurahanList.find((k) => k.id === umat.kelurahanId)?.nama}
            />
          </dl>
        )}
      </Card>

      {/* 4. Pendidikan, Pekerjaan & Aktivitas Sosial */}
      <Card className="p-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-base font-semibold text-fg">
              Pendidikan, Pekerjaan & Aktivitas Sosial
            </h2>
            <p className="mt-0.5 text-xs text-muted">
              Riwayat pendidikan, profesi, keterampilan, dan pelayanan sosial.
            </p>
          </div>
          {editing !== "education" ? (
            <Button
              variant="outline"
              className="h-9 px-3 text-xs"
              onClick={() => startEdit("education")}
            >
              <Pencil aria-hidden className="size-3.5" />
              Ubah
            </Button>
          ) : null}
        </div>

        {editing === "education" ? (
          <form onSubmit={saveEducation} className="mt-5 flex flex-col gap-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Pendidikan
                <select
                  value={pendidikan}
                  onChange={(e) => setPendidikan(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                >
                  <option value="">-- Tidak dipilih --</option>
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
                  <option value="">-- Tidak dipilih --</option>
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
                  <option value="">-- Tidak dipilih --</option>
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
                  <option value="">-- Tidak dipilih --</option>
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
                  <option value="">-- Tidak dipilih --</option>
                  {(aktivitasQuery.data ?? []).map((l) => (
                    <option key={l.id} value={l.code}>
                      {l.label}
                    </option>
                  ))}
                </select>
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
              <Button type="submit" disabled={updateUmat.isPending}>
                Simpan Perubahan
              </Button>
            </div>
          </form>
        ) : (
          <dl className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <DetailRow
              label="Pendidikan"
              value={lookupLabel(pendidikanQuery.data, umat.pendidikan)}
            />
            <DetailRow label="Jurusan" value={umat.jurusan} />
            <DetailRow
              label="Pekerjaan"
              value={lookupLabel(pekerjaanQuery.data, umat.pekerjaan)}
            />
            <DetailRow
              label="Profesi"
              value={lookupLabel(profesiQuery.data, umat.profesi)}
            />
            <DetailRow label="Keterampilan" value={umat.keterampilan} />
            <DetailRow
              label="Disabilitas"
              value={lookupLabel(disabilitasQuery.data, umat.disabilitas)}
            />
            <DetailRow
              label="Aktivitas Sosial"
              value={lookupLabel(aktivitasQuery.data, umat.aktivitasSosial)}
            />
          </dl>
        )}
      </Card>

      {/* 5. Tautan Akun Pengguna Domus */}
      <Card className="p-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-base font-semibold text-fg">
              Tautan Akun Pengguna Domus
            </h2>
            <p className="mt-0.5 text-xs text-muted">
              Hubungkan data umat ini ke akun login anggota paroki secara
              manual.
            </p>
          </div>
          {editing !== "userLink" ? (
            <Button
              variant="outline"
              className="h-9 px-3 text-xs"
              onClick={() => startEdit("userLink")}
            >
              <Pencil aria-hidden className="size-3.5" />
              Ubah
            </Button>
          ) : null}
        </div>

        {editing === "userLink" ? (
          <form onSubmit={saveUserLink} className="mt-5 flex flex-col gap-4">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Anggota Paroki Terdaftar
              <select
                value={userId}
                onChange={(e) => setUserId(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              >
                <option value="">-- Tidak terhubung akun --</option>
                {memberList.map((m) => (
                  <option key={m.userId} value={m.userId}>
                    {m.name ? `${m.name} (${m.email})` : m.email}
                  </option>
                ))}
              </select>
            </label>
            <div className="flex items-center justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditing(null)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={updateUmat.isPending}>
                Simpan Perubahan
              </Button>
            </div>
          </form>
        ) : (
          <dl className="mt-4">
            <DetailRow
              label="Akun Pengguna Terhubung"
              value={
                linkedMember
                  ? `${linkedMember.name || linkedMember.email} (${linkedMember.email})`
                  : umat.userId
                    ? umat.userId
                    : "-"
              }
            />
          </dl>
        )}
      </Card>

      {/* 6. Riwayat Sakramen */}
      <Card className="p-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-base font-semibold text-fg">
              Riwayat Sakramen ({sakramen.length})
            </h2>
            <p className="mt-0.5 text-xs text-muted">
              Catatan sakramen yang diterima umat secara kronologis.
            </p>
          </div>
          <Link to={`/umat/${umat.id}/sakramen/new`} className="no-underline">
            <Button variant="outline" className="h-9 px-3 text-xs">
              <Plus aria-hidden className="size-3.5" />
              Catat Sakramen
            </Button>
          </Link>
        </div>

        {sakramen.length === 0 ? (
          <p className="mt-4 rounded-lg border border-dashed border-line p-4 text-sm text-muted">
            Belum ada catatan sakramen untuk umat ini.
          </p>
        ) : (
          <div className="mt-4 divide-y divide-line rounded-lg border border-line">
            {sakramen.map((sk) => {
              const keuskupanName =
                keuskupanList.find((k) => k.id === sk.keuskupanId)?.nama ??
                sk.keuskupanNama;
              const parokiName =
                parokiList.find((p) => p.id === sk.parokiSakramenId)?.nama ??
                sk.parokiNama;
              const label = sakramenJenisLabel(sk.jenis);

              return (
                <div
                  key={sk.id}
                  className="flex flex-wrap items-center justify-between gap-3 p-4"
                >
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="text-sm font-bold text-fg">{label}</span>
                      {sk.tanggal ? (
                        <Badge tone="neutral">{sk.tanggal}</Badge>
                      ) : null}
                    </div>
                    <p className="mt-1 text-xs text-muted">
                      {[
                        parokiName,
                        keuskupanName,
                        sk.buku ? `Buku ${sk.buku}` : null,
                        sk.nomor ? `No. ${sk.nomor}` : null,
                      ]
                        .filter(Boolean)
                        .join(" · ") || "-"}
                    </p>
                    {sk.catatan ? (
                      <p className="mt-1 text-xs text-fg">{sk.catatan}</p>
                    ) : null}
                  </div>
                  <div className="flex items-center gap-2">
                    <Link
                      to={`/umat/${umat.id}/sakramen/${sk.id}`}
                      className="no-underline"
                    >
                      <Button variant="outline" className="h-8 px-2.5 text-xs">
                        Ubah
                      </Button>
                    </Link>
                    <Button
                      variant="danger"
                      className="h-8 px-2.5 text-xs"
                      disabled={deleteSakramen.isPending}
                      onClick={() =>
                        sk.id && setPendingDeleteSakramen({ id: sk.id, label })
                      }
                    >
                      Hapus
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </Card>

      {/* 7. Zona Bahaya */}
      <InductiveDangerZone
        title="Hapus Data Umat"
        description="Data umat akan dihapus secara lunak (soft delete) dari direktori aktif."
        action={
          <Button
            variant="danger"
            disabled={deleteUmat.isPending}
            onClick={() => setConfirmDeleteUmat(true)}
          >
            <Trash2 aria-hidden className="size-4" />
            Hapus Umat
          </Button>
        }
      />

      <ConfirmDialog
        open={confirmDeleteUmat}
        title="Hapus Data Umat?"
        description={`Data umat "${umat.nama}" akan dihapus dari direktori aktif paroki.`}
        confirmLabel="Hapus Umat"
        cancelLabel="Batal"
        confirmVariant="danger"
        pending={deleteUmat.isPending}
        onConfirm={handleConfirmDeleteUmat}
        onCancel={() => setConfirmDeleteUmat(false)}
      />

      <ConfirmDialog
        open={Boolean(pendingDeleteSakramen)}
        title="Hapus Catatan Sakramen?"
        description={`Catatan sakramen ${pendingDeleteSakramen?.label ?? ""} akan dihapus permanen.`}
        confirmLabel="Hapus"
        cancelLabel="Batal"
        confirmVariant="danger"
        pending={deleteSakramen.isPending}
        onConfirm={handleConfirmDeleteSakramen}
        onCancel={() => setPendingDeleteSakramen(null)}
      />
    </div>
  );
}
