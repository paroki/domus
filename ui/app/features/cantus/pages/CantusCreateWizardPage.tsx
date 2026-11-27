import { ArrowLeft, ArrowRight, Plus, Trash2, X } from "lucide-react";
import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { ApiError } from "~/shared/api/client";
import { useMe } from "~/shared/api/session";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useCreateCantus } from "../api/actions";
import { resolveCantusErrorMessage } from "../api/errors";
import type { CantusMedia } from "../api/types";
import { CantusForbiddenCard } from "../components/CantusForbiddenCard";
import {
  buildLaguMeta,
  buildMazmurMeta,
  CANTUS_JENIS_OPTIONS,
  CANTUS_KATEGORI_SUGGESTIONS,
  CANTUS_MEDIA_TYPES,
  CANTUS_SIKLUS_OPTIONS,
  CANTUS_SUMBER_SUGGESTIONS,
  isCantusAdmin,
  normalizeTagList,
  slugifyCantus,
  validateCantusMediaURL,
  validateCantusSlug,
} from "../model";

export function meta() {
  return [{ title: "Tambah Nyanyian · Nyanyian · Domus" }];
}

const STEP_LABELS = [
  "1. Identitas & Klasifikasi",
  "2. Lirik / Bait Mazmur & BPI",
  "3. Kredit & Media Partitur / Audio",
] as const;

export default function CantusCreateWizardPage() {
  const navigate = useNavigate();
  const { data: me } = useMe();
  const allowed = isCantusAdmin(me);

  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Step 1
  const [jenis, setJenis] = useState<"lagu" | "mazmur" | "profan">("lagu");
  const [judul, setJudul] = useState("");
  const [sumber, setSumber] = useState("");
  const [nomor, setNomor] = useState("");
  const [slug, setSlug] = useState("");
  const [slugManual, setSlugManual] = useState(false);
  const [kategori, setKategori] = useState("");
  const [nadaDasar, setNadaDasar] = useState("");
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState("");

  // Step 2 (lagu/profan)
  const [lirik, setLirik] = useState("");
  const [birama, setBirama] = useState("");
  const [tempo, setTempo] = useState("");
  const [notAwal, setNotAwal] = useState("");

  // Step 2 (mazmur)
  const [perayaan, setPerayaan] = useState("");
  const [siklusTahun, setSiklusTahun] = useState("");
  const [masa, setMasa] = useState("");
  const [refAlkitab, setRefAlkitab] = useState("");
  const [ulangan, setUlangan] = useState("");
  const [bait, setBait] = useState("");
  const [bpi, setBpi] = useState("");

  // Step 3
  const [kreditLagu, setKreditLagu] = useState("");
  const [kreditSyair, setKreditSyair] = useState("");
  const [kreditAransemen, setKreditAransemen] = useState("");
  const [media, setMedia] = useState<CantusMedia[]>([]);

  const createCantus = useCreateCantus();

  if (me && !allowed) {
    return <CantusForbiddenCard />;
  }

  function syncAutoSlug(
    nextSumber: string,
    nextNomor: string,
    nextJudul: string,
  ) {
    if (!slugManual) {
      setSlug(slugifyCantus(nextSumber, nextNomor, nextJudul));
    }
  }

  function addTagsFromInput(raw: string) {
    const parts = raw
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    if (parts.length === 0) return;
    setTags((prev) => normalizeTagList([...prev, ...parts]));
    setTagInput("");
  }

  function validateStep1(): boolean {
    if (!judul.trim()) {
      setErrorMsg("Judul nyanyian wajib diisi.");
      return false;
    }
    if (!validateCantusSlug(slug)) {
      setErrorMsg(
        "Slug wajib diisi (2-64 karakter huruf kecil, angka, tanda hubung atau garis bawah).",
      );
      return false;
    }
    return true;
  }

  function validateMediaList(): boolean {
    for (const item of media) {
      if (item.viewUrl && !validateCantusMediaURL(item.viewUrl)) {
        setErrorMsg(
          "URL tampilan media harus menggunakan skema http:// atau https:// yang valid.",
        );
        return false;
      }
      if (item.downloadUrl && !validateCantusMediaURL(item.downloadUrl)) {
        setErrorMsg(
          "URL unduhan media harus menggunakan skema http:// atau https:// yang valid.",
        );
        return false;
      }
    }
    return true;
  }

  async function handleSubmit() {
    setErrorMsg(null);
    if (!validateStep1() || !validateMediaList()) return;

    const meta =
      jenis === "mazmur"
        ? buildMazmurMeta({
            perayaan,
            siklusTahun,
            masa,
            refAlkitab,
            ulangan,
            bait,
            bpi,
          })
        : buildLaguMeta({ birama, tempo, notAwal });

    const kredit: Record<string, string> = {};
    if (kreditLagu.trim()) kredit.lagu = kreditLagu.trim();
    if (kreditSyair.trim()) kredit.syair = kreditSyair.trim();
    if (kreditAransemen.trim()) kredit.aransemen = kreditAransemen.trim();

    const cleanedMedia = media
      .map((m) => ({
        type: m.type,
        ...(m.title?.trim() ? { title: m.title.trim() } : {}),
        ...(m.viewUrl?.trim() ? { viewUrl: m.viewUrl.trim() } : {}),
        ...(m.downloadUrl?.trim() ? { downloadUrl: m.downloadUrl.trim() } : {}),
      }))
      .filter((m) => m.title || m.viewUrl || m.downloadUrl);

    try {
      const created = await createCantus.mutateAsync({
        slug: slug.trim().toLowerCase(),
        jenis,
        judul: judul.trim(),
        ...(sumber.trim() ? { sumber: sumber.trim() } : {}),
        ...(nomor.trim() ? { nomor: nomor.trim() } : {}),
        ...(kategori.trim() ? { kategori: kategori.trim() } : {}),
        ...(nadaDasar.trim() ? { nadaDasar: nadaDasar.trim() } : {}),
        ...(tags.length > 0 ? { tag: tags } : {}),
        ...(lirik.trim() ? { lirik: lirik.trim() } : {}),
        ...(Object.keys(kredit).length > 0 ? { kredit } : {}),
        ...(cleanedMedia.length > 0 ? { media: cleanedMedia } : {}),
        ...(Object.keys(meta).length > 0 ? { meta } : {}),
      });
      if (created?.id) {
        navigate(`/cantus/${created.id}`);
      } else {
        navigate("/cantus");
      }
    } catch (err) {
      if (err instanceof ApiError && err.code === "CANTUS_CODE_CONFLICT") {
        setStep(1);
      }
      setErrorMsg(resolveCantusErrorMessage(err));
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Nyanyian", to: "/cantus" },
          { label: "Tambah Nyanyian" },
        ]}
        title="Tambah Nyanyian"
        subtitle="Tambahkan entri lagu liturgi, mazmur tanggapan & BPI, atau lagu rohani ke katalog global."
      />

      <Card className="flex flex-wrap items-center gap-2 p-4">
        {STEP_LABELS.map((label, idx) => {
          const num = (idx + 1) as 1 | 2 | 3;
          const active = step === num;
          return (
            <span
              key={label}
              className={`rounded-full px-3 py-1 text-xs font-semibold ${
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

      {errorMsg && (
        <div
          role="alert"
          className="rounded-lg border border-danger/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
        >
          {errorMsg}
        </div>
      )}

      {step === 1 && (
        <InductiveFormSection
          title="Langkah 1 — Identitas & Klasifikasi"
          description="Jenis, judul, dan slug wajib diisi. Anda dapat langsung menyimpan setelah langkah ini."
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Jenis Nyanyian *
              <select
                value={jenis}
                onChange={(e) =>
                  setJenis(e.target.value as "lagu" | "mazmur" | "profan")
                }
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              >
                {CANTUS_JENIS_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Judul *
              <input
                type="text"
                required
                value={judul}
                onChange={(e) => {
                  setJudul(e.target.value);
                  syncAutoSlug(sumber, nomor, e.target.value);
                }}
                placeholder="mis. Tuhan Kasihanilah Kami"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Sumber
              <input
                type="text"
                list="cantus-sumber-list"
                value={sumber}
                onChange={(e) => {
                  setSumber(e.target.value);
                  syncAutoSlug(e.target.value, nomor, judul);
                }}
                placeholder="PS, MB, KA, Mazmur"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
              <datalist id="cantus-sumber-list">
                {CANTUS_SUMBER_SUGGESTIONS.map((s) => (
                  <option key={s} value={s} />
                ))}
              </datalist>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Nomor Lagu / Mazmur
              <input
                type="text"
                value={nomor}
                onChange={(e) => {
                  setNomor(e.target.value);
                  syncAutoSlug(sumber, e.target.value, judul);
                }}
                placeholder="mis. PS-542 atau MT-012"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
              Slug *
              <input
                type="text"
                required
                value={slug}
                onChange={(e) => {
                  setSlugManual(true);
                  setSlug(e.target.value);
                }}
                placeholder="mis. ps-542-tuhan-kasihanilah-kami"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 font-mono text-base text-fg sm:text-sm"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Kategori
              <input
                type="text"
                list="cantus-kategori-list"
                value={kategori}
                onChange={(e) => setKategori(e.target.value)}
                placeholder="mis. Pembuka, Persembahan"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
              <datalist id="cantus-kategori-list">
                {CANTUS_KATEGORI_SUGGESTIONS.map((k) => (
                  <option key={k} value={k} />
                ))}
              </datalist>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Nada Dasar
              <input
                type="text"
                value={nadaDasar}
                onChange={(e) => setNadaDasar(e.target.value)}
                placeholder="mis. D, C, 1=Do"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>

            <div className="flex flex-col gap-2 sm:col-span-2">
              <span className="text-sm font-medium text-fg">Tag</span>
              <div className="flex flex-col gap-2 sm:flex-row">
                <input
                  type="text"
                  value={tagInput}
                  onChange={(e) => setTagInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === ",") {
                      e.preventDefault();
                      addTagsFromInput(tagInput);
                    }
                  }}
                  placeholder="Ketik tag lalu tekan Enter atau Tambah Tag"
                  className="h-11 flex-1 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
                <Button
                  type="button"
                  variant="outline"
                  className="h-11 w-full sm:w-auto"
                  onClick={() => addTagsFromInput(tagInput)}
                >
                  <Plus aria-hidden className="size-4" />
                  Tambah Tag
                </Button>
              </div>
              {tags.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {tags.map((tg) => (
                    <span
                      key={tg}
                      className="inline-flex min-h-10 items-center gap-1.5 rounded-full border border-line bg-surface-2 px-3 text-xs font-medium text-fg"
                    >
                      #{tg}
                      <button
                        type="button"
                        onClick={() =>
                          setTags((prev) => prev.filter((item) => item !== tg))
                        }
                        aria-label={`Hapus tag ${tg}`}
                        className="inline-flex size-6 items-center justify-center rounded-full text-muted hover:text-fg"
                      >
                        <X aria-hidden className="size-3.5" />
                      </button>
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className="mt-6 flex flex-col gap-2 border-t border-line pt-4 sm:flex-row sm:items-center sm:justify-between">
            <Link to="/cantus" className="w-full no-underline sm:w-auto">
              <Button variant="outline" className="h-11 w-full sm:w-auto">
                Batal
              </Button>
            </Link>
            <div className="flex flex-col gap-2 sm:flex-row">
              <Button
                type="button"
                variant="outline"
                disabled={createCantus.isPending}
                className="h-11 w-full sm:w-auto"
                onClick={handleSubmit}
              >
                {createCantus.isPending
                  ? "Menyimpan…"
                  : "Simpan & Buat Nyanyian"}
              </Button>
              <Button
                type="button"
                className="h-11 w-full sm:w-auto"
                onClick={() => {
                  setErrorMsg(null);
                  if (validateStep1()) setStep(2);
                }}
              >
                Lanjut ke Lirik / Mazmur
                <ArrowRight aria-hidden className="size-4" />
              </Button>
            </div>
          </div>
        </InductiveFormSection>
      )}

      {step === 2 && (
        <InductiveFormSection
          title={
            jenis === "mazmur"
              ? "Langkah 2 — Bait Mazmur & BPI"
              : "Langkah 2 — Lirik & Notasi Dasar"
          }
          description={
            jenis === "mazmur"
              ? "Isi ulangan, bait mazmur, ayat Alkitab, dan Bait Pengantar Injil."
              : "Masukkan lirik lengkap serta info musik opsional."
          }
        >
          {jenis === "mazmur" ? (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Perayaan / Hari Minggu
                <input
                  type="text"
                  value={perayaan}
                  onChange={(e) => setPerayaan(e.target.value)}
                  placeholder="mis. Minggu Adven I"
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Siklus Tahun
                <select
                  value={siklusTahun}
                  onChange={(e) => setSiklusTahun(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                >
                  <option value="">Pilih siklus…</option>
                  {CANTUS_SIKLUS_OPTIONS.map((opt) => (
                    <option key={opt} value={opt}>
                      Tahun {opt}
                    </option>
                  ))}
                </select>
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Masa Liturgi
                <input
                  type="text"
                  value={masa}
                  onChange={(e) => setMasa(e.target.value)}
                  placeholder="mis. Adven, Paskah, Biasa"
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Referensi Alkitab
                <input
                  type="text"
                  value={refAlkitab}
                  onChange={(e) => setRefAlkitab(e.target.value)}
                  placeholder="mis. Mzm 25:4-5, 8-10"
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
                Ulangan (Refrain)
                <textarea
                  rows={2}
                  value={ulangan}
                  onChange={(e) => setUlangan(e.target.value)}
                  placeholder="Refrain yang dinyanyikan umat…"
                  className="rounded-lg border border-line bg-canvas p-3.5 text-base text-fg sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
                Bait Mazmur
                <textarea
                  rows={5}
                  value={bait}
                  onChange={(e) => setBait(e.target.value)}
                  placeholder="Bait-bait mazmur…"
                  className="rounded-lg border border-line bg-canvas p-3.5 text-base text-fg sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
                Bait Pengantar Injil (BPI)
                <textarea
                  rows={3}
                  value={bpi}
                  onChange={(e) => setBpi(e.target.value)}
                  placeholder="Alleluya / Terpujilah Kristus Tuhan…"
                  className="rounded-lg border border-line bg-canvas p-3.5 text-base text-fg sm:text-sm"
                />
              </label>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-3">
                Lirik Lagu
                <textarea
                  rows={7}
                  value={lirik}
                  onChange={(e) => setLirik(e.target.value)}
                  placeholder="Masukkan teks lirik lengkap lagu…"
                  className="rounded-lg border border-line bg-canvas p-3.5 text-base text-fg sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Birama
                <input
                  type="text"
                  value={birama}
                  onChange={(e) => setBirama(e.target.value)}
                  placeholder="mis. 4/4, 3/4"
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Tempo
                <input
                  type="text"
                  value={tempo}
                  onChange={(e) => setTempo(e.target.value)}
                  placeholder="mis. Andante, 84 BPM"
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Not Awal
                <input
                  type="text"
                  value={notAwal}
                  onChange={(e) => setNotAwal(e.target.value)}
                  placeholder="mis. 3 . 4 5 |"
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 font-mono text-base text-fg sm:text-sm"
                />
              </label>
            </div>
          )}

          <div className="mt-6 flex flex-col gap-2 border-t border-line pt-4 sm:flex-row sm:items-center sm:justify-between">
            <Button
              type="button"
              variant="outline"
              className="h-11 w-full sm:w-auto"
              onClick={() => setStep(1)}
            >
              <ArrowLeft aria-hidden className="size-4" />
              Kembali
            </Button>
            <div className="flex flex-col gap-2 sm:flex-row">
              <Button
                type="button"
                variant="outline"
                disabled={createCantus.isPending}
                className="h-11 w-full sm:w-auto"
                onClick={handleSubmit}
              >
                {createCantus.isPending
                  ? "Menyimpan…"
                  : "Simpan & Buat Nyanyian"}
              </Button>
              <Button
                type="button"
                className="h-11 w-full sm:w-auto"
                onClick={() => {
                  setErrorMsg(null);
                  setStep(3);
                }}
              >
                Lanjut ke Kredit & Media
                <ArrowRight aria-hidden className="size-4" />
              </Button>
            </div>
          </div>
        </InductiveFormSection>
      )}

      {step === 3 && (
        <InductiveFormSection
          title="Langkah 3 — Kredit & Media Partitur / Audio"
          description="Opsional. Tambahkan nama pencipta dan tautan partitur (gambar/PDF) atau rekaman latihan (audio/video)."
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Pencipta Lagu
              <input
                type="text"
                value={kreditLagu}
                onChange={(e) => setKreditLagu(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Pengarang Syair
              <input
                type="text"
                value={kreditSyair}
                onChange={(e) => setKreditSyair(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Aransemen
              <input
                type="text"
                value={kreditAransemen}
                onChange={(e) => setKreditAransemen(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>
          </div>

          <div className="mt-6 flex flex-col gap-3 border-t border-line pt-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <span className="text-sm font-semibold text-fg">
                Lampiran Media
              </span>
              <Button
                type="button"
                variant="outline"
                className="h-11 w-full sm:w-auto"
                onClick={() =>
                  setMedia((prev) => [
                    ...prev,
                    { type: "pdf", title: "", viewUrl: "", downloadUrl: "" },
                  ])
                }
              >
                <Plus aria-hidden className="size-4" />
                Tambah Media
              </Button>
            </div>

            {media.map((m, idx) => (
              <div
                key={`media-${idx}`}
                className="grid grid-cols-1 gap-3 rounded-lg border border-line bg-surface-2/30 p-3 sm:grid-cols-2"
              >
                <label className="flex flex-col gap-1 text-xs font-medium text-fg">
                  Tipe
                  <select
                    value={m.type}
                    onChange={(e) =>
                      setMedia((prev) =>
                        prev.map((row, i) =>
                          i === idx ? { ...row, type: e.target.value } : row,
                        ),
                      )
                    }
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg sm:text-sm"
                  >
                    {CANTUS_MEDIA_TYPES.map((mt) => (
                      <option key={mt} value={mt}>
                        {mt.toUpperCase()}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="flex flex-col gap-1 text-xs font-medium text-fg">
                  Judul Media
                  <input
                    type="text"
                    value={m.title ?? ""}
                    onChange={(e) =>
                      setMedia((prev) =>
                        prev.map((row, i) =>
                          i === idx ? { ...row, title: e.target.value } : row,
                        ),
                      )
                    }
                    placeholder="mis. Partitur SATB, Audio Sopran"
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg sm:text-sm"
                  />
                </label>

                <label className="flex flex-col gap-1 text-xs font-medium text-fg">
                  URL Tampilan (http/https)
                  <input
                    type="url"
                    value={m.viewUrl ?? ""}
                    onChange={(e) =>
                      setMedia((prev) =>
                        prev.map((row, i) =>
                          i === idx ? { ...row, viewUrl: e.target.value } : row,
                        ),
                      )
                    }
                    placeholder="https://..."
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg sm:text-sm"
                  />
                </label>

                <label className="flex flex-col gap-1 text-xs font-medium text-fg">
                  URL Unduhan (opsional)
                  <input
                    type="url"
                    value={m.downloadUrl ?? ""}
                    onChange={(e) =>
                      setMedia((prev) =>
                        prev.map((row, i) =>
                          i === idx
                            ? { ...row, downloadUrl: e.target.value }
                            : row,
                        ),
                      )
                    }
                    placeholder="https://..."
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg sm:text-sm"
                  />
                </label>

                <div className="sm:col-span-2">
                  <Button
                    type="button"
                    variant="outline"
                    className="h-11 w-full text-danger-fg sm:w-auto"
                    onClick={() =>
                      setMedia((prev) => prev.filter((_, i) => i !== idx))
                    }
                  >
                    <Trash2 aria-hidden className="size-4" />
                    Hapus Media
                  </Button>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-6 flex flex-col gap-2 border-t border-line pt-4 sm:flex-row sm:items-center sm:justify-between">
            <Button
              type="button"
              variant="outline"
              className="h-11 w-full sm:w-auto"
              onClick={() => setStep(2)}
            >
              <ArrowLeft aria-hidden className="size-4" />
              Kembali
            </Button>
            <Button
              type="button"
              disabled={createCantus.isPending}
              className="h-11 w-full sm:w-auto"
              onClick={handleSubmit}
            >
              {createCantus.isPending ? "Menyimpan…" : "Simpan & Buat Nyanyian"}
            </Button>
          </div>
        </InductiveFormSection>
      )}
    </div>
  );
}
