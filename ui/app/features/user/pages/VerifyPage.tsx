import {
  AlertCircle,
  ArrowLeft,
  CheckCircle2,
  FileCheck2,
  Info,
  ShieldCheck,
  Upload,
} from "lucide-react";
import { type FormEvent, useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router";
import { ApiError } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { compressImage } from "~/shared/util/image";
import { useVerifyConfirm, useVerifyExtract } from "../api/actions";
import {
  useMe,
  useVerifyLingkunganOptions,
  useVerifyWilayahOptions,
} from "../api/queries";
import type { ExtractedKKMember, IdentityExtraction } from "../api/types";
import type { Route } from "./+types/VerifyPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: "Verifikasi Umat · Domus" }];
}

type DocMode = "ktp" | "non_ktp";

interface SelectedDoc {
  raw: File;
  compressed?: File;
  previewUrl: string;
}

const INPUT_CLASS =
  "h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm";

function MatchStatusBanner({ status }: { status?: string }) {
  if (status === "matched") {
    return (
      <div
        role="status"
        className="flex items-start gap-3 rounded-xl border border-success-fg/20 bg-success-bg p-4 text-sm text-success-fg"
      >
        <CheckCircle2 aria-hidden className="mt-0.5 size-5 shrink-0" />
        <div>
          <p className="font-semibold">Data cocok dengan buku induk paroki</p>
          <p className="mt-0.5 opacity-90">
            Periksa kembali hasil pembacaan dokumen sebelum mengonfirmasi.
          </p>
        </div>
      </div>
    );
  }
  if (status === "conflict") {
    return (
      <div
        role="alert"
        className="flex items-start gap-3 rounded-xl border border-warning-fg/20 bg-warning-bg p-4 text-sm text-warning-fg"
      >
        <AlertCircle aria-hidden className="mt-0.5 size-5 shrink-0" />
        <div>
          <p className="font-semibold">
            Ditemukan perbedaan data dengan paroki
          </p>
          <p className="mt-0.5 opacity-90">
            Periksa kolom di bawah dengan saksama atau hubungi admin paroki jika
            terjadi konflik akun.
          </p>
        </div>
      </div>
    );
  }
  return (
    <div
      role="status"
      className="flex items-start gap-3 rounded-xl border border-line bg-surface-2/50 p-4 text-sm text-fg"
    >
      <Info aria-hidden className="mt-0.5 size-5 shrink-0 text-primary" />
      <div>
        <p className="font-semibold">Pendaftaran data umat baru</p>
        <p className="mt-0.5 text-muted">
          Periksa dan lengkapi data hasil pembacaan dokumen sebelum disimpan.
        </p>
      </div>
    </div>
  );
}

export default function VerifyPage() {
  const navigate = useNavigate();
  const { data: me } = useMe();
  const activeParokiId = me?.activeParokiId ?? "";

  const extractMutation = useVerifyExtract();
  const confirmMutation = useVerifyConfirm();

  const [step, setStep] = useState<1 | 2>(1);
  const [docMode, setDocMode] = useState<DocMode>("ktp");

  const [ktpDoc, setKtpDoc] = useState<SelectedDoc | null>(null);
  const [kkDoc, setKkDoc] = useState<SelectedDoc | null>(null);
  const [kartuPelajarDoc, setKartuPelajarDoc] = useState<SelectedDoc | null>(
    null,
  );

  const [matchStatus, setMatchStatus] = useState<string | undefined>(undefined);
  const [extraction, setExtraction] = useState<IdentityExtraction>({});
  const [memberIndex, setMemberIndex] = useState<number | null>(null);
  const [wilayahId, setWilayahId] = useState("");
  const [lingkunganId, setLingkunganId] = useState("");

  const [isCompressing, setIsCompressing] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const previewUrlsRef = useRef<Set<string>>(new Set());

  useEffect(() => {
    const urls = previewUrlsRef.current;
    return () => {
      for (const url of urls) {
        URL.revokeObjectURL(url);
      }
      urls.clear();
    };
  }, []);

  const wilayahQuery = useVerifyWilayahOptions(
    activeParokiId,
    Boolean(activeParokiId) && step === 2,
  );
  const lingkunganQuery = useVerifyLingkunganOptions(
    activeParokiId,
    wilayahId || undefined,
    Boolean(activeParokiId) && step === 2,
  );

  function selectDoc(
    file: File | null | undefined,
    current: SelectedDoc | null,
    setter: (doc: SelectedDoc | null) => void,
  ) {
    if (current?.previewUrl) {
      URL.revokeObjectURL(current.previewUrl);
      previewUrlsRef.current.delete(current.previewUrl);
    }
    if (!file) {
      setter(null);
      return;
    }
    const previewUrl = URL.createObjectURL(file);
    previewUrlsRef.current.add(previewUrl);
    setter({ raw: file, previewUrl });
  }

  async function ensureCompressed(doc: SelectedDoc): Promise<File> {
    if (doc.compressed) return doc.compressed;
    const compressed = await compressImage(doc.raw);
    doc.compressed = compressed;
    return compressed;
  }

  async function handleExtractSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);

    if (docMode === "ktp" && !ktpDoc) {
      setErrorMsg("Foto KTP wajib diunggah.");
      return;
    }
    if (docMode === "non_ktp" && !kkDoc) {
      setErrorMsg("Foto Kartu Keluarga (KK) wajib diunggah.");
      return;
    }

    try {
      setIsCompressing(true);
      const ktp =
        docMode === "ktp" && ktpDoc
          ? await ensureCompressed(ktpDoc)
          : undefined;
      const kk = kkDoc ? await ensureCompressed(kkDoc) : undefined;
      const kartuPelajar =
        docMode === "non_ktp" && kartuPelajarDoc
          ? await ensureCompressed(kartuPelajarDoc)
          : undefined;
      setIsCompressing(false);

      const res = await extractMutation.mutateAsync({
        ktp,
        kk,
        kartuPelajar,
      });

      const nextExtraction = res.extraction ?? {};
      setExtraction(nextExtraction);
      setMatchStatus(res.matchStatus);
      const anggota = nextExtraction.kk?.anggota ?? [];
      setMemberIndex(anggota.length === 1 ? 0 : null);
      setStep(2);
    } catch (err) {
      setIsCompressing(false);
      if (err instanceof ApiError) {
        setErrorMsg(err.details[0]?.message ?? err.message);
      } else if (err instanceof Error) {
        setErrorMsg(err.message);
      } else {
        setErrorMsg("Gagal memproses dokumen.");
      }
    }
  }

  function updateKtpField(
    field: keyof NonNullable<IdentityExtraction["ktp"]>,
    value: string,
  ) {
    setExtraction((prev) => ({
      ...prev,
      ktp: {
        ...prev.ktp,
        [field]: value,
      },
    }));
  }

  function updateKkField(
    field: keyof Omit<NonNullable<IdentityExtraction["kk"]>, "anggota">,
    value: string,
  ) {
    setExtraction((prev) => ({
      ...prev,
      kk: {
        ...prev.kk,
        [field]: value,
      },
    }));
  }

  const kkUploaded = Boolean(kkDoc);
  const kkMembers: ExtractedKKMember[] = extraction.kk?.anggota ?? [];

  async function handleConfirmSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);

    if (kkUploaded && kkMembers.length > 0 && memberIndex === null) {
      setErrorMsg("Pilih baris nama Anda pada daftar anggota Kartu Keluarga.");
      return;
    }

    try {
      await confirmMutation.mutateAsync({
        payload: extraction,
        ktp: docMode === "ktp" ? ktpDoc?.compressed : undefined,
        kk: kkDoc?.compressed,
        kartuPelajar:
          docMode === "non_ktp" ? kartuPelajarDoc?.compressed : undefined,
        memberIndex:
          kkUploaded && memberIndex !== null ? memberIndex : undefined,
        wilayahId: wilayahId || undefined,
        lingkunganId: lingkunganId || undefined,
      });
      navigate("/me", { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        setErrorMsg(err.details[0]?.message ?? err.message);
      } else if (err instanceof Error) {
        setErrorMsg(err.message);
      } else {
        setErrorMsg("Gagal mengonfirmasi verifikasi.");
      }
    }
  }

  if (me?.verified) {
    return (
      <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
        <InductivePageHeader
          breadcrumbs={[
            { label: "Beranda", to: "/" },
            { label: "Profil", to: "/me" },
            { label: "Verifikasi Umat" },
          ]}
          title="Verifikasi Umat"
          subtitle="Status keanggotaan dan data sensus keluarga paroki Anda."
        />
        <Card className="flex flex-col gap-4 p-4 sm:p-6">
          <div className="flex items-center gap-2">
            <Badge tone="success">
              <ShieldCheck aria-hidden className="mr-1 size-4" />
              Terverifikasi
            </Badge>
          </div>
          <p className="text-sm text-fg">
            Akun Anda sudah terverifikasi dan terhubung dengan data umat paroki.
          </p>
          <div>
            <Link to="/me" className="no-underline">
              <Button variant="outline" className="w-full sm:w-auto">
                Kembali ke Profil Saya
              </Button>
            </Link>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Profil", to: "/me" },
          { label: "Verifikasi Umat" },
        ]}
        title="Verifikasi Data Umat"
        subtitle={
          step === 1
            ? "Langkah 1 dari 2 · Unggah foto dokumen identitas Anda."
            : "Langkah 2 dari 2 · Periksa hasil ekstraksi dan konfirmasi data."
        }
      />

      {errorMsg && (
        <div
          role="alert"
          className="rounded-xl border border-danger-fg/20 bg-danger-bg px-4 py-3 text-sm font-medium text-danger-fg"
        >
          {errorMsg}
        </div>
      )}

      {step === 1 ? (
        <form onSubmit={handleExtractSubmit} className="flex flex-col gap-6">
          <InductiveFormSection
            title="Pilih Jenis Dokumen"
            description="Pilih apakah Anda sudah memiliki KTP atau masih menggunakan Kartu Keluarga / Kartu Pelajar."
          >
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <label className="flex cursor-pointer items-center gap-3 rounded-xl border border-line bg-canvas p-3.5 text-sm font-medium text-fg has-[:checked]:border-primary has-[:checked]:bg-primary/5">
                <input
                  type="radio"
                  name="docMode"
                  value="ktp"
                  checked={docMode === "ktp"}
                  onChange={() => setDocMode("ktp")}
                  className="size-4 accent-primary"
                />
                <span>Punya KTP</span>
              </label>
              <label className="flex cursor-pointer items-center gap-3 rounded-xl border border-line bg-canvas p-3.5 text-sm font-medium text-fg has-[:checked]:border-primary has-[:checked]:bg-primary/5">
                <input
                  type="radio"
                  name="docMode"
                  value="non_ktp"
                  checked={docMode === "non_ktp"}
                  onChange={() => setDocMode("non_ktp")}
                  className="size-4 accent-primary"
                />
                <span>Belum punya KTP (pelajar/anak)</span>
              </label>
            </div>
          </InductiveFormSection>

          <InductiveFormSection
            title="Unggah Foto Dokumen"
            description="Foto akan dikompresi otomatis di perangkat Anda sebelum dikirim (maks. 2 MiB per berkas)."
          >
            <div className="flex flex-col gap-5">
              {docMode === "ktp" ? (
                <div className="flex flex-col gap-2">
                  <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                    <span>
                      Foto KTP <span className="text-danger-fg">*</span>
                    </span>
                    <input
                      type="file"
                      accept="image/*"
                      onChange={(e) =>
                        selectDoc(e.target.files?.[0], ktpDoc, setKtpDoc)
                      }
                      className="block w-full rounded-lg border border-line bg-canvas p-2 text-base text-fg file:mr-3 file:rounded-md file:border-0 file:bg-primary/10 file:px-3 file:py-1.5 file:text-sm file:font-semibold file:text-primary sm:text-sm"
                    />
                  </label>
                  {ktpDoc?.previewUrl && (
                    <img
                      src={ktpDoc.previewUrl}
                      alt="Pratinjau KTP"
                      className="max-h-48 w-auto rounded-lg border border-line object-contain"
                    />
                  )}
                </div>
              ) : (
                <div className="flex flex-col gap-2">
                  <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                    <span>Foto Kartu Pelajar / KIA (opsional)</span>
                    <input
                      type="file"
                      accept="image/*"
                      onChange={(e) =>
                        selectDoc(
                          e.target.files?.[0],
                          kartuPelajarDoc,
                          setKartuPelajarDoc,
                        )
                      }
                      className="block w-full rounded-lg border border-line bg-canvas p-2 text-base text-fg file:mr-3 file:rounded-md file:border-0 file:bg-primary/10 file:px-3 file:py-1.5 file:text-sm file:font-semibold file:text-primary sm:text-sm"
                    />
                  </label>
                  {kartuPelajarDoc?.previewUrl && (
                    <img
                      src={kartuPelajarDoc.previewUrl}
                      alt="Pratinjau Kartu Pelajar"
                      className="max-h-48 w-auto rounded-lg border border-line object-contain"
                    />
                  )}
                </div>
              )}

              <div className="flex flex-col gap-2">
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  <span>
                    Foto Kartu Keluarga (KK){" "}
                    {docMode === "non_ktp" ? (
                      <span className="text-danger-fg">*</span>
                    ) : (
                      <span className="text-muted">(opsional)</span>
                    )}
                  </span>
                  <input
                    type="file"
                    accept="image/*"
                    onChange={(e) =>
                      selectDoc(e.target.files?.[0], kkDoc, setKkDoc)
                    }
                    className="block w-full rounded-lg border border-line bg-canvas p-2 text-base text-fg file:mr-3 file:rounded-md file:border-0 file:bg-primary/10 file:px-3 file:py-1.5 file:text-sm file:font-semibold file:text-primary sm:text-sm"
                  />
                </label>
                {kkDoc?.previewUrl && (
                  <img
                    src={kkDoc.previewUrl}
                    alt="Pratinjau Kartu Keluarga"
                    className="max-h-48 w-auto rounded-lg border border-line object-contain"
                  />
                )}
              </div>
            </div>
          </InductiveFormSection>

          <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-between">
            <Link to="/me" className="no-underline">
              <Button variant="outline" className="w-full sm:w-auto">
                <ArrowLeft aria-hidden className="size-4" />
                Batal
              </Button>
            </Link>
            <Button
              type="submit"
              disabled={isCompressing || extractMutation.isPending}
              className="w-full sm:w-auto"
            >
              <Upload aria-hidden className="size-4" />
              {isCompressing
                ? "Mengompresi foto..."
                : extractMutation.isPending
                  ? "Membaca dokumen..."
                  : "Lanjut ke Tinjauan"}
            </Button>
          </div>
        </form>
      ) : (
        <form onSubmit={handleConfirmSubmit} className="flex flex-col gap-6">
          <MatchStatusBanner status={matchStatus} />

          {(docMode === "ktp" || extraction.ktp) && (
            <InductiveFormSection
              title="Data Pribadi (KTP)"
              description="Perbaiki jika ada huruf atau angka yang kurang tepat."
            >
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
                  Nama Lengkap
                  <input
                    type="text"
                    value={extraction.ktp?.nama ?? ""}
                    onChange={(e) => updateKtpField("nama", e.target.value)}
                    className={INPUT_CLASS}
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  NIK
                  <input
                    type="text"
                    value={extraction.ktp?.nik ?? ""}
                    onChange={(e) => updateKtpField("nik", e.target.value)}
                    className={`${INPUT_CLASS} font-mono`}
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Jenis Kelamin
                  <select
                    value={extraction.ktp?.jenisKelamin ?? ""}
                    onChange={(e) =>
                      updateKtpField("jenisKelamin", e.target.value)
                    }
                    className={INPUT_CLASS}
                  >
                    <option value="">Pilih jenis kelamin</option>
                    <option value="pria">Laki-laki (Pria)</option>
                    <option value="wanita">Perempuan (Wanita)</option>
                  </select>
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Tempat Lahir
                  <input
                    type="text"
                    value={extraction.ktp?.tempatLahir ?? ""}
                    onChange={(e) =>
                      updateKtpField("tempatLahir", e.target.value)
                    }
                    className={INPUT_CLASS}
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Tanggal Lahir
                  <input
                    type="date"
                    value={extraction.ktp?.tanggalLahir ?? ""}
                    onChange={(e) =>
                      updateKtpField("tanggalLahir", e.target.value)
                    }
                    className={INPUT_CLASS}
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
                  Alamat
                  <input
                    type="text"
                    value={extraction.ktp?.alamat ?? ""}
                    onChange={(e) => updateKtpField("alamat", e.target.value)}
                    className={INPUT_CLASS}
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Agama
                  <input
                    type="text"
                    value={extraction.ktp?.agama ?? ""}
                    onChange={(e) => updateKtpField("agama", e.target.value)}
                    className={INPUT_CLASS}
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Status Pernikahan
                  <input
                    type="text"
                    value={extraction.ktp?.statusPernikahan ?? ""}
                    onChange={(e) =>
                      updateKtpField("statusPernikahan", e.target.value)
                    }
                    className={INPUT_CLASS}
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Pekerjaan
                  <input
                    type="text"
                    value={extraction.ktp?.pekerjaan ?? ""}
                    onChange={(e) =>
                      updateKtpField("pekerjaan", e.target.value)
                    }
                    className={INPUT_CLASS}
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Golongan Darah
                  <input
                    type="text"
                    value={extraction.ktp?.golonganDarah ?? ""}
                    onChange={(e) =>
                      updateKtpField("golonganDarah", e.target.value)
                    }
                    className={INPUT_CLASS}
                  />
                </label>
              </div>
            </InductiveFormSection>
          )}

          {(kkUploaded || extraction.kk) && (
            <InductiveFormSection
              title="Data Kartu Keluarga (KK)"
              description="Periksa nomor KK, nama keluarga, dan pilih baris Anda pada daftar anggota."
            >
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Nomor KK
                  <input
                    type="text"
                    value={extraction.kk?.nomorKk ?? ""}
                    onChange={(e) => updateKkField("nomorKk", e.target.value)}
                    className={`${INPUT_CLASS} font-mono`}
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Nama Kepala Keluarga
                  <input
                    type="text"
                    value={extraction.kk?.namaKeluarga ?? ""}
                    onChange={(e) =>
                      updateKkField("namaKeluarga", e.target.value)
                    }
                    className={INPUT_CLASS}
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
                  Alamat KK
                  <input
                    type="text"
                    value={extraction.kk?.alamat ?? ""}
                    onChange={(e) => updateKkField("alamat", e.target.value)}
                    className={INPUT_CLASS}
                  />
                </label>
              </div>

              {kkMembers.length > 0 && (
                <div className="mt-2 flex flex-col gap-2">
                  <p className="text-sm font-medium text-fg">
                    Pilih baris Anda dalam Kartu Keluarga{" "}
                    <span className="text-danger-fg">*</span>
                  </p>
                  <div className="overflow-x-auto rounded-xl border border-line">
                    <table className="w-full text-left text-sm">
                      <thead className="bg-surface-2/60 text-xs font-semibold text-muted">
                        <tr>
                          <th className="px-3 py-2.5">Pilih</th>
                          <th className="px-3 py-2.5">Nama</th>
                          <th className="px-3 py-2.5">NIK</th>
                          <th className="px-3 py-2.5">Hubungan</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-line">
                        {kkMembers.map((member, idx) => (
                          <tr
                            key={`${member.nik ?? member.nama ?? "row"}-${idx}`}
                            onClick={() => setMemberIndex(idx)}
                            className={`cursor-pointer transition-colors hover:bg-surface-2/40 ${
                              memberIndex === idx ? "bg-primary/5" : ""
                            }`}
                          >
                            <td className="px-3 py-2.5">
                              <input
                                type="radio"
                                name="kkMemberIndex"
                                value={idx}
                                checked={memberIndex === idx}
                                onChange={() => setMemberIndex(idx)}
                                aria-label={`Pilih ${member.nama ?? `anggota ${idx + 1}`}`}
                                className="size-4 accent-primary"
                              />
                            </td>
                            <td className="px-3 py-2.5 font-medium text-fg">
                              {member.nama || "-"}
                            </td>
                            <td className="px-3 py-2.5 font-mono text-xs text-muted">
                              {member.nik || "-"}
                            </td>
                            <td className="px-3 py-2.5 text-muted">
                              {member.hubunganKeluarga || "-"}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </InductiveFormSection>
          )}

          <InductiveFormSection
            title="Wilayah & Lingkungan (Opsional)"
            description="Pilih wilayah dan lingkungan domisili Anda di paroki jika sudah mengetahui."
          >
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Wilayah
                <select
                  value={wilayahId}
                  onChange={(e) => {
                    setWilayahId(e.target.value);
                    setLingkunganId("");
                  }}
                  className={INPUT_CLASS}
                >
                  <option value="">Belum dipilih</option>
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
                  onChange={(e) => setLingkunganId(e.target.value)}
                  className={INPUT_CLASS}
                >
                  <option value="">Belum dipilih</option>
                  {(lingkunganQuery.data ?? []).map((l) => (
                    <option key={l.id} value={l.id}>
                      {l.nama}
                    </option>
                  ))}
                </select>
              </label>
            </div>
          </InductiveFormSection>

          <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-between">
            <Button
              variant="outline"
              onClick={() => {
                setErrorMsg(null);
                setStep(1);
              }}
              className="w-full sm:w-auto"
            >
              <ArrowLeft aria-hidden className="size-4" />
              Kembali
            </Button>
            <Button
              type="submit"
              disabled={confirmMutation.isPending}
              className="w-full sm:w-auto"
            >
              <FileCheck2 aria-hidden className="size-4" />
              {confirmMutation.isPending
                ? "Menyimpan..."
                : "Konfirmasi & Simpan"}
            </Button>
          </div>
        </form>
      )}
    </div>
  );
}
