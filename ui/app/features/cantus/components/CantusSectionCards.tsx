import {
  ExternalLink,
  FileText,
  Music2,
  Pencil,
  Plus,
  Trash2,
  Video,
  X,
} from "lucide-react";
import { type FormEvent, useState } from "react";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import { InductiveDangerZone } from "~/shared/ui/iui";
import { useDeleteCantus, useUpdateCantus } from "../api/actions";
import { resolveCantusErrorMessage } from "../api/errors";
import type { Cantus, CantusMedia } from "../api/types";
import {
  buildCantusKredit,
  buildLaguMeta,
  buildMazmurMeta,
  CANTUS_JENIS_OPTIONS,
  CANTUS_KATEGORI_SUGGESTIONS,
  CANTUS_MEDIA_TYPES,
  CANTUS_SIKLUS_OPTIONS,
  CANTUS_SUMBER_SUGGESTIONS,
  cantusJenisLabel,
  normalizeTagList,
  parseCantusKredit,
  parseLaguMeta,
  parseMazmurMeta,
  validateCantusMediaURL,
  validateCantusSlug,
} from "../model";

export type CantusSectionKey = "identity" | "lyrics" | "kredit" | "media";

function DetailRow({ label, value }: { label: string; value?: string | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-xs font-medium text-muted">{label}</dt>
      <dd className="text-sm font-medium text-fg">{value || "-"}</dd>
    </div>
  );
}

export function CantusSectionCards({
  cantus,
  canManage,
  onFeedback,
  onDeleted,
}: {
  cantus: Cantus;
  canManage: boolean;
  onFeedback: (
    fb: { tone: "success" | "error"; message: string } | null,
  ) => void;
  onDeleted: () => void;
}) {
  const [editing, setEditing] = useState<CantusSectionKey | null>(null);
  const [pendingConfirm, setPendingConfirm] = useState<
    "deactivate" | "delete" | null
  >(null);

  const updateCantus = useUpdateCantus();
  const deleteCantus = useDeleteCantus();

  // Section 1 state
  const [jenis, setJenis] = useState<"lagu" | "mazmur" | "profan">("lagu");
  const [judul, setJudul] = useState("");
  const [sumber, setSumber] = useState("");
  const [nomor, setNomor] = useState("");
  const [slug, setSlug] = useState("");
  const [kategori, setKategori] = useState("");
  const [nadaDasar, setNadaDasar] = useState("");
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState("");

  // Section 2 state
  const [lirik, setLirik] = useState("");
  const [birama, setBirama] = useState("");
  const [tempo, setTempo] = useState("");
  const [notAwal, setNotAwal] = useState("");
  const [perayaan, setPerayaan] = useState("");
  const [siklusTahun, setSiklusTahun] = useState("");
  const [masa, setMasa] = useState("");
  const [refAlkitab, setRefAlkitab] = useState("");
  const [ulangan, setUlangan] = useState("");
  const [bait, setBait] = useState("");
  const [bpi, setBpi] = useState("");

  // Section 3 state
  const [kreditLagu, setKreditLagu] = useState("");
  const [kreditSyair, setKreditSyair] = useState("");
  const [kreditAransemen, setKreditAransemen] = useState("");

  // Section 4 state
  const [media, setMedia] = useState<CantusMedia[]>([]);

  const cantusId = cantus.id ?? "";
  const existingMeta =
    (cantus.meta as Record<string, unknown> | undefined) ?? {};
  const mazmurMeta = parseMazmurMeta(existingMeta);
  const laguMeta = parseLaguMeta(existingMeta);
  const kreditMap = cantus.kredit ?? {};
  const extraKreditEntries = Object.entries(kreditMap).filter(
    ([k]) => k !== "lagu" && k !== "syair" && k !== "aransemen",
  );

  function startEdit(section: CantusSectionKey) {
    onFeedback(null);
    if (section === "identity") {
      setJenis((cantus.jenis as "lagu" | "mazmur" | "profan") ?? "lagu");
      setJudul(cantus.judul ?? "");
      setSumber(cantus.sumber ?? "");
      setNomor(cantus.nomor ?? "");
      setSlug(cantus.slug ?? "");
      setKategori(cantus.kategori ?? "");
      setNadaDasar(cantus.nadaDasar ?? "");
      setTags(cantus.tag ?? []);
      setTagInput("");
    } else if (section === "lyrics") {
      setLirik(cantus.lirik ?? "");
      setBirama(laguMeta.birama ?? "");
      setTempo(laguMeta.tempo ?? "");
      setNotAwal(laguMeta.notAwal ?? "");
      setPerayaan(mazmurMeta.perayaan ?? "");
      setSiklusTahun(mazmurMeta.siklusTahun ?? "");
      setMasa(mazmurMeta.masa ?? "");
      setRefAlkitab(mazmurMeta.refAlkitab ?? "");
      setUlangan(mazmurMeta.ulangan ?? "");
      setBait(mazmurMeta.bait ?? "");
      setBpi(mazmurMeta.bpi ?? "");
    } else if (section === "kredit") {
      const parsed = parseCantusKredit(cantus.kredit);
      setKreditLagu(parsed.lagu ?? "");
      setKreditSyair(parsed.syair ?? "");
      setKreditAransemen(parsed.aransemen ?? "");
    } else if (section === "media") {
      setMedia((cantus.media ?? []).map((m) => ({ ...m })));
    }
    setEditing(section);
  }

  async function saveIdentity(e: FormEvent) {
    e.preventDefault();
    onFeedback(null);
    if (!judul.trim() || !validateCantusSlug(slug)) {
      onFeedback({
        tone: "error",
        message: "Judul wajib diisi dan slug harus valid (2-64 karakter).",
      });
      return;
    }
    try {
      await updateCantus.mutateAsync({
        id: cantusId,
        body: {
          jenis,
          judul: judul.trim(),
          slug: slug.trim().toLowerCase(),
          sumber: sumber.trim(),
          nomor: nomor.trim(),
          kategori: kategori.trim(),
          nadaDasar: nadaDasar.trim(),
          tag: tags,
        },
      });
      setEditing(null);
      onFeedback({
        tone: "success",
        message: "Identitas nyanyian diperbarui.",
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveCantusErrorMessage(err) });
    }
  }

  async function saveLyrics(e: FormEvent) {
    e.preventDefault();
    onFeedback(null);
    const mergedMeta =
      cantus.jenis === "mazmur"
        ? buildMazmurMeta(
            {
              perayaan,
              siklusTahun,
              masa,
              refAlkitab,
              ulangan,
              bait,
              bpi,
            },
            existingMeta,
          )
        : buildLaguMeta({ birama, tempo, notAwal }, existingMeta);

    try {
      await updateCantus.mutateAsync({
        id: cantusId,
        body: {
          ...(cantus.jenis !== "mazmur" ? { lirik: lirik.trim() } : {}),
          meta: mergedMeta,
        },
      });
      setEditing(null);
      onFeedback({
        tone: "success",
        message: "Lirik dan metadata musik diperbarui.",
      });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveCantusErrorMessage(err) });
    }
  }

  async function saveKredit(e: FormEvent) {
    e.preventDefault();
    onFeedback(null);
    const merged = buildCantusKredit(
      { lagu: kreditLagu, syair: kreditSyair, aransemen: kreditAransemen },
      cantus.kredit,
    );
    try {
      await updateCantus.mutateAsync({
        id: cantusId,
        body: { kredit: merged },
      });
      setEditing(null);
      onFeedback({ tone: "success", message: "Kredit pencipta diperbarui." });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveCantusErrorMessage(err) });
    }
  }

  async function saveMedia(e: FormEvent) {
    e.preventDefault();
    onFeedback(null);
    for (const item of media) {
      if (item.viewUrl && !validateCantusMediaURL(item.viewUrl)) {
        onFeedback({
          tone: "error",
          message: "URL tampilan harus menggunakan http:// atau https://.",
        });
        return;
      }
      if (item.downloadUrl && !validateCantusMediaURL(item.downloadUrl)) {
        onFeedback({
          tone: "error",
          message: "URL unduhan harus menggunakan http:// atau https://.",
        });
        return;
      }
    }
    const cleaned = media
      .map((m) => ({
        type: m.type,
        ...(m.title?.trim() ? { title: m.title.trim() } : {}),
        ...(m.viewUrl?.trim() ? { viewUrl: m.viewUrl.trim() } : {}),
        ...(m.downloadUrl?.trim() ? { downloadUrl: m.downloadUrl.trim() } : {}),
      }))
      .filter((m) => m.title || m.viewUrl || m.downloadUrl);

    try {
      await updateCantus.mutateAsync({
        id: cantusId,
        body: { media: cleaned },
      });
      setEditing(null);
      onFeedback({ tone: "success", message: "Lampiran media diperbarui." });
    } catch (err) {
      onFeedback({ tone: "error", message: resolveCantusErrorMessage(err) });
    }
  }

  async function handleConfirm() {
    const action = pendingConfirm;
    setPendingConfirm(null);
    onFeedback(null);
    try {
      if (action === "deactivate") {
        await updateCantus.mutateAsync({
          id: cantusId,
          body: { active: false },
        });
        onFeedback({ tone: "success", message: "Nyanyian dinonaktifkan." });
      } else if (action === "delete") {
        await deleteCantus.mutateAsync(cantusId);
        onDeleted();
      }
    } catch (err) {
      onFeedback({ tone: "error", message: resolveCantusErrorMessage(err) });
    }
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Section 1: Identity & Classification */}
      <Card className="p-6">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-lg font-bold text-fg">
              1. Identitas & Klasifikasi
            </h2>
            <Badge tone={cantus.active !== false ? "success" : "danger"}>
              {cantus.active !== false ? "Aktif" : "Nonaktif"}
            </Badge>
          </div>
          {canManage && editing !== "identity" && (
            <Button
              variant="outline"
              className="h-11 w-full sm:w-auto"
              onClick={() => startEdit("identity")}
            >
              <Pencil aria-hidden className="size-4" />
              Ubah Identitas
            </Button>
          )}
        </div>

        {editing === "identity" ? (
          <form
            onSubmit={saveIdentity}
            className="grid grid-cols-1 gap-4 sm:grid-cols-2"
          >
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Jenis
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
                onChange={(e) => setJudul(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Sumber
              <input
                type="text"
                list="cantus-edit-sumber"
                value={sumber}
                onChange={(e) => setSumber(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
              <datalist id="cantus-edit-sumber">
                {CANTUS_SUMBER_SUGGESTIONS.map((s) => (
                  <option key={s} value={s} />
                ))}
              </datalist>
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Nomor
              <input
                type="text"
                value={nomor}
                onChange={(e) => setNomor(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Slug *
              <input
                type="text"
                required
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 font-mono text-base text-fg sm:text-sm"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Kategori
              <input
                type="text"
                list="cantus-edit-kategori"
                value={kategori}
                onChange={(e) => setKategori(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
              <datalist id="cantus-edit-kategori">
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
                      const next = tagInput
                        .split(",")
                        .map((s) => s.trim())
                        .filter(Boolean);
                      if (next.length > 0) {
                        setTags((prev) => normalizeTagList([...prev, ...next]));
                        setTagInput("");
                      }
                    }
                  }}
                  placeholder="Tambah tag lalu Enter"
                  className="h-11 flex-1 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
                <Button
                  type="button"
                  variant="outline"
                  className="h-11 w-full sm:w-auto"
                  onClick={() => {
                    const next = tagInput
                      .split(",")
                      .map((s) => s.trim())
                      .filter(Boolean);
                    if (next.length > 0) {
                      setTags((prev) => normalizeTagList([...prev, ...next]));
                      setTagInput("");
                    }
                  }}
                >
                  Tambah Tag
                </Button>
              </div>
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
            </div>
            <div className="flex flex-col gap-2 sm:col-span-2 sm:flex-row">
              <Button
                type="submit"
                disabled={updateCantus.isPending}
                className="h-11 w-full sm:w-auto"
              >
                {updateCantus.isPending ? "Menyimpan…" : "Simpan"}
              </Button>
              <Button
                type="button"
                variant="outline"
                className="h-11 w-full sm:w-auto"
                onClick={() => setEditing(null)}
              >
                Batal
              </Button>
            </div>
          </form>
        ) : (
          <dl className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <DetailRow label="Jenis" value={cantusJenisLabel(cantus.jenis)} />
            <DetailRow label="Sumber" value={cantus.sumber} />
            <DetailRow label="Nomor" value={cantus.nomor} />
            <DetailRow label="Slug" value={cantus.slug} />
            <DetailRow label="Kategori" value={cantus.kategori} />
            <DetailRow label="Nada Dasar" value={cantus.nadaDasar} />
            <div className="sm:col-span-3">
              <dt className="text-xs font-medium text-muted">Tag</dt>
              <dd className="mt-1 flex flex-wrap gap-1.5">
                {(cantus.tag ?? []).length > 0 ? (
                  (cantus.tag ?? []).map((tg) => (
                    <Badge key={tg} tone="neutral">
                      #{tg}
                    </Badge>
                  ))
                ) : (
                  <span className="text-sm text-muted">-</span>
                )}
              </dd>
            </div>
          </dl>
        )}
      </Card>

      {/* Section 2: Lyrics / Mazmur & BPI */}
      <Card className="p-6">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg font-bold text-fg">
            {cantus.jenis === "mazmur"
              ? "2. Bait Mazmur & BPI"
              : "2. Lirik & Notasi Dasar"}
          </h2>
          {canManage && editing !== "lyrics" && (
            <Button
              variant="outline"
              className="h-11 w-full sm:w-auto"
              onClick={() => startEdit("lyrics")}
            >
              <Pencil aria-hidden className="size-4" />
              Ubah Lirik / Mazmur
            </Button>
          )}
        </div>

        {editing === "lyrics" ? (
          <form onSubmit={saveLyrics} className="flex flex-col gap-4">
            {cantus.jenis === "mazmur" ? (
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Perayaan
                  <input
                    type="text"
                    value={perayaan}
                    onChange={(e) => setPerayaan(e.target.value)}
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
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Referensi Alkitab
                  <input
                    type="text"
                    value={refAlkitab}
                    onChange={(e) => setRefAlkitab(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
                  Ulangan (Refrain)
                  <textarea
                    rows={2}
                    value={ulangan}
                    onChange={(e) => setUlangan(e.target.value)}
                    className="rounded-lg border border-line bg-canvas p-3.5 text-base text-fg sm:text-sm"
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
                  Bait Mazmur
                  <textarea
                    rows={5}
                    value={bait}
                    onChange={(e) => setBait(e.target.value)}
                    className="rounded-lg border border-line bg-canvas p-3.5 text-base text-fg sm:text-sm"
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
                  Bait Pengantar Injil (BPI)
                  <textarea
                    rows={3}
                    value={bpi}
                    onChange={(e) => setBpi(e.target.value)}
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
                    className="rounded-lg border border-line bg-canvas p-3.5 text-base text-fg sm:text-sm"
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Birama
                  <input
                    type="text"
                    value={birama}
                    onChange={(e) => setBirama(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Tempo
                  <input
                    type="text"
                    value={tempo}
                    onChange={(e) => setTempo(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                  />
                </label>
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  Not Awal
                  <input
                    type="text"
                    value={notAwal}
                    onChange={(e) => setNotAwal(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3.5 font-mono text-base text-fg sm:text-sm"
                  />
                </label>
              </div>
            )}
            <div className="flex flex-col gap-2 sm:flex-row">
              <Button
                type="submit"
                disabled={updateCantus.isPending}
                className="h-11 w-full sm:w-auto"
              >
                {updateCantus.isPending ? "Menyimpan…" : "Simpan"}
              </Button>
              <Button
                type="button"
                variant="outline"
                className="h-11 w-full sm:w-auto"
                onClick={() => setEditing(null)}
              >
                Batal
              </Button>
            </div>
          </form>
        ) : cantus.jenis === "mazmur" ? (
          <div className="flex flex-col gap-4">
            <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <DetailRow label="Perayaan" value={mazmurMeta.perayaan} />
              <DetailRow
                label="Siklus Tahun"
                value={
                  mazmurMeta.siklusTahun
                    ? `Tahun ${mazmurMeta.siklusTahun}`
                    : undefined
                }
              />
              <DetailRow label="Masa Liturgi" value={mazmurMeta.masa} />
              <DetailRow
                label="Referensi Alkitab"
                value={mazmurMeta.refAlkitab}
              />
            </dl>
            {mazmurMeta.ulangan && (
              <div className="rounded-lg border border-primary/25 bg-primary/5 p-4">
                <p className="text-xs font-semibold uppercase tracking-wide text-primary">
                  Ulangan (Refrain)
                </p>
                <p className="mt-1 font-sans leading-relaxed whitespace-pre-wrap text-sm font-medium text-fg">
                  {mazmurMeta.ulangan}
                </p>
              </div>
            )}
            {mazmurMeta.bait && (
              <div>
                <p className="text-xs font-medium text-muted">Bait Mazmur</p>
                <p className="mt-1 font-sans leading-relaxed whitespace-pre-wrap text-sm text-fg">
                  {mazmurMeta.bait}
                </p>
              </div>
            )}
            {mazmurMeta.bpi && (
              <div className="rounded-lg border border-line bg-surface-2/40 p-4">
                <p className="text-xs font-semibold uppercase tracking-wide text-muted">
                  Bait Pengantar Injil (BPI)
                </p>
                <p className="mt-1 font-sans leading-relaxed whitespace-pre-wrap text-sm text-fg">
                  {mazmurMeta.bpi}
                </p>
              </div>
            )}
            {!mazmurMeta.ulangan && !mazmurMeta.bait && !mazmurMeta.bpi && (
              <p className="text-sm text-muted">
                Belum ada teks mazmur atau BPI.
              </p>
            )}
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            <dl className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <DetailRow label="Birama" value={laguMeta.birama} />
              <DetailRow label="Tempo" value={laguMeta.tempo} />
              <DetailRow label="Not Awal" value={laguMeta.notAwal} />
            </dl>
            <div>
              <p className="text-xs font-medium text-muted">Lirik</p>
              <p className="mt-1 font-sans leading-relaxed whitespace-pre-wrap text-sm text-fg">
                {cantus.lirik?.trim() || "Belum ada lirik yang dicatat."}
              </p>
            </div>
          </div>
        )}
      </Card>

      {/* Section 3: Credits */}
      <Card className="p-6">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg font-bold text-fg">
            3. Kredit Pencipta & Aransemen
          </h2>
          {canManage && editing !== "kredit" && (
            <Button
              variant="outline"
              className="h-11 w-full sm:w-auto"
              onClick={() => startEdit("kredit")}
            >
              <Pencil aria-hidden className="size-4" />
              Ubah Kredit
            </Button>
          )}
        </div>

        {editing === "kredit" ? (
          <form
            onSubmit={saveKredit}
            className="grid grid-cols-1 gap-4 sm:grid-cols-3"
          >
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
            <div className="flex flex-col gap-2 sm:col-span-3 sm:flex-row">
              <Button
                type="submit"
                disabled={updateCantus.isPending}
                className="h-11 w-full sm:w-auto"
              >
                {updateCantus.isPending ? "Menyimpan…" : "Simpan"}
              </Button>
              <Button
                type="button"
                variant="outline"
                className="h-11 w-full sm:w-auto"
                onClick={() => setEditing(null)}
              >
                Batal
              </Button>
            </div>
          </form>
        ) : (
          <dl className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <DetailRow label="Lagu" value={kreditMap.lagu} />
            <DetailRow label="Syair" value={kreditMap.syair} />
            <DetailRow label="Aransemen" value={kreditMap.aransemen} />
            {extraKreditEntries.map(([k, v]) => (
              <DetailRow key={k} label={k} value={v} />
            ))}
          </dl>
        )}
      </Card>

      {/* Section 4: Media */}
      <Card className="p-6">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg font-bold text-fg">
            4. Media Partitur & Rekaman Latihan
          </h2>
          {canManage && editing !== "media" && (
            <Button
              variant="outline"
              className="h-11 w-full sm:w-auto"
              onClick={() => startEdit("media")}
            >
              <Pencil aria-hidden className="size-4" />
              Ubah Media
            </Button>
          )}
        </div>

        {editing === "media" ? (
          <form onSubmit={saveMedia} className="flex flex-col gap-4">
            <div className="flex justify-end">
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
                key={`edit-media-${idx}`}
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
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg sm:text-sm"
                  />
                </label>
                <label className="flex flex-col gap-1 text-xs font-medium text-fg">
                  URL Unduhan
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
            <div className="flex flex-col gap-2 sm:flex-row">
              <Button
                type="submit"
                disabled={updateCantus.isPending}
                className="h-11 w-full sm:w-auto"
              >
                {updateCantus.isPending ? "Menyimpan…" : "Simpan"}
              </Button>
              <Button
                type="button"
                variant="outline"
                className="h-11 w-full sm:w-auto"
                onClick={() => setEditing(null)}
              >
                Batal
              </Button>
            </div>
          </form>
        ) : (cantus.media ?? []).length === 0 ? (
          <p className="text-sm text-muted">Belum ada lampiran media.</p>
        ) : (
          <div className="flex flex-col gap-4">
            {(cantus.media ?? []).map((m, idx) => {
              const src = m.viewUrl || m.downloadUrl;
              if (m.type === "image" && src) {
                return (
                  <div
                    key={`media-view-${idx}`}
                    className="flex flex-col gap-2 rounded-lg border border-line p-3"
                  >
                    <span className="text-sm font-semibold text-fg">
                      {m.title || "Partitur Gambar"}
                    </span>
                    <img
                      src={src}
                      alt={m.title || cantus.judul || "Partitur"}
                      loading="lazy"
                      className="max-h-[32rem] w-full rounded-lg border border-line bg-canvas object-contain"
                    />
                    {m.downloadUrl && (
                      <a
                        href={m.downloadUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex h-11 w-full items-center justify-center gap-1.5 rounded-lg border border-line px-4 text-sm font-semibold text-fg no-underline hover:bg-surface-2 sm:w-auto"
                      >
                        <ExternalLink aria-hidden className="size-4" />
                        Unduh Partitur
                      </a>
                    )}
                  </div>
                );
              }
              if (m.type === "audio" && src) {
                return (
                  <div
                    key={`media-view-${idx}`}
                    className="flex flex-col gap-2 rounded-lg border border-line bg-surface-2/40 p-3"
                  >
                    <div className="flex items-center gap-2 text-sm font-semibold text-fg">
                      <Music2 aria-hidden className="size-4 text-primary" />
                      <span>{m.title || "Audio Latihan"}</span>
                    </div>
                    {/* biome-ignore lint/a11y/useMediaCaption: practice rehearsal audio tracks */}
                    <audio
                      controls
                      preload="none"
                      src={src}
                      className="w-full"
                    />
                    {m.downloadUrl && (
                      <a
                        href={m.downloadUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex h-11 w-full items-center justify-center gap-1.5 rounded-lg border border-line px-4 text-sm font-semibold text-fg no-underline hover:bg-surface-2 sm:w-auto"
                      >
                        <ExternalLink aria-hidden className="size-4" />
                        Unduh Audio
                      </a>
                    )}
                  </div>
                );
              }
              const Icon = m.type === "video" ? Video : FileText;
              return (
                <div
                  key={`media-view-${idx}`}
                  className="flex flex-col gap-3 rounded-lg border border-line p-3 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div className="flex items-center gap-2">
                    <Icon aria-hidden className="size-4 text-primary" />
                    <span className="text-sm font-semibold text-fg">
                      {m.title || (m.type ?? "").toUpperCase()}
                    </span>
                    <Badge tone="neutral">{(m.type ?? "").toUpperCase()}</Badge>
                  </div>
                  <div className="flex flex-col gap-2 sm:flex-row">
                    {m.viewUrl && (
                      <a
                        href={m.viewUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex h-11 w-full items-center justify-center gap-1.5 rounded-lg border border-line px-4 text-sm font-semibold text-fg no-underline hover:bg-surface-2 sm:w-auto"
                      >
                        <ExternalLink aria-hidden className="size-4" />
                        Buka
                      </a>
                    )}
                    {m.downloadUrl && (
                      <a
                        href={m.downloadUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex h-11 w-full items-center justify-center gap-1.5 rounded-lg border border-line px-4 text-sm font-semibold text-fg no-underline hover:bg-surface-2 sm:w-auto"
                      >
                        <ExternalLink aria-hidden className="size-4" />
                        Unduh
                      </a>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </Card>

      {/* Section 5: Danger Zone */}
      {canManage && (
        <InductiveDangerZone
          title="5. Zona Bahaya"
          description="Nonaktifkan sementara atau hapus nyanyian ini dari katalog global."
          action={
            <div className="flex flex-col gap-2 sm:flex-row">
              {cantus.active !== false ? (
                <Button
                  type="button"
                  variant="outline"
                  disabled={updateCantus.isPending}
                  className="h-11 w-full sm:w-auto"
                  onClick={() => setPendingConfirm("deactivate")}
                >
                  Nonaktifkan Nyanyian
                </Button>
              ) : (
                <Button
                  type="button"
                  variant="outline"
                  disabled={updateCantus.isPending}
                  className="h-11 w-full sm:w-auto"
                  onClick={async () => {
                    onFeedback(null);
                    try {
                      await updateCantus.mutateAsync({
                        id: cantusId,
                        body: { active: true },
                      });
                      onFeedback({
                        tone: "success",
                        message: "Nyanyian diaktifkan kembali.",
                      });
                    } catch (err) {
                      onFeedback({
                        tone: "error",
                        message: resolveCantusErrorMessage(err),
                      });
                    }
                  }}
                >
                  Aktifkan Kembali Nyanyian
                </Button>
              )}

              <Button
                type="button"
                variant="danger"
                disabled={deleteCantus.isPending}
                className="h-11 w-full sm:w-auto"
                onClick={() => setPendingConfirm("delete")}
              >
                <Trash2 aria-hidden className="size-4" />
                Hapus Nyanyian
              </Button>
            </div>
          }
        />
      )}

      <ConfirmDialog
        open={Boolean(pendingConfirm)}
        title={
          pendingConfirm === "delete"
            ? "Hapus Nyanyian?"
            : "Nonaktifkan Nyanyian?"
        }
        description={
          pendingConfirm === "delete"
            ? `Nyanyian "${cantus.judul}" akan dihapus dari katalog aktif.`
            : `Nyanyian "${cantus.judul}" tidak akan muncul di pencarian aktif.`
        }
        confirmLabel={
          pendingConfirm === "delete" ? "Hapus Nyanyian" : "Nonaktifkan"
        }
        cancelLabel="Batal"
        pending={updateCantus.isPending || deleteCantus.isPending}
        onConfirm={handleConfirm}
        onCancel={() => setPendingConfirm(null)}
      />
    </div>
  );
}
