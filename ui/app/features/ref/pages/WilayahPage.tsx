import { ArrowLeft, Pencil, Plus, Power, Trash2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import {
  useCreateLingkungan,
  useCreateWilayah,
  useDeleteLingkungan,
  useDeleteWilayah,
  useUpdateLingkungan,
  useUpdateWilayah,
} from "../api/actions";
import {
  useCurrentIdentity,
  useLingkunganSearch,
  useWilayahSearch,
} from "../api/queries";
import type { Lingkungan, Wilayah } from "../api/types";
import { isParokiAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Wilayah & Lingkungan — Domus" }];
}

export default function WilayahPage() {
  const { t } = useTranslation();
  const { data: me } = useCurrentIdentity();
  const allowed = isParokiAdmin(me?.roles);
  const hasTenant = Boolean(me?.activeParokiId);
  const enabled = allowed && hasTenant;

  const [wilayahJenisFilter, setWilayahJenisFilter] = useState<
    "" | "wilayah" | "stasi"
  >("");
  const [selectedWilayahId, setSelectedWilayahId] = useState<string>("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Wilayah form state
  const [editingWilayah, setEditingWilayah] = useState<Wilayah | null>(null);
  const [wJenis, setWJenis] = useState<"wilayah" | "stasi">("wilayah");
  const [wKode, setWKode] = useState("");
  const [wNama, setWNama] = useState("");
  const [wKoordinator, setWKoordinator] = useState("");
  const [wTelepon, setWTelepon] = useState("");

  // Lingkungan form state
  const [editingLingkungan, setEditingLingkungan] = useState<Lingkungan | null>(
    null,
  );
  const [lWilayahId, setLWilayahId] = useState("");
  const [lKode, setLKode] = useState("");
  const [lNama, setLNama] = useState("");
  const [lPelindung, setLPelindung] = useState("");
  const [lKetua, setLKetua] = useState("");
  const [lTelepon, setLTelepon] = useState("");

  const wilayahQuery = useWilayahSearch(
    {
      filter: wilayahJenisFilter ? { jenis: wilayahJenisFilter } : undefined,
    },
    enabled,
  );
  const lingkunganQuery = useLingkunganSearch(
    {
      filter: selectedWilayahId ? { wilayahId: selectedWilayahId } : undefined,
    },
    enabled,
  );
  const wilayahList =
    wilayahQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const lingkunganList =
    lingkunganQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const createWilayah = useCreateWilayah();
  const updateWilayah = useUpdateWilayah();
  const deleteWilayah = useDeleteWilayah();
  const createLingkungan = useCreateLingkungan();
  const updateLingkungan = useUpdateLingkungan();
  const deleteLingkungan = useDeleteLingkungan();

  function resetWilayahForm() {
    setEditingWilayah(null);
    setWJenis("wilayah");
    setWKode("");
    setWNama("");
    setWKoordinator("");
    setWTelepon("");
  }

  function startEditWilayah(z: Wilayah) {
    setEditingWilayah(z);
    setWJenis((z.jenis as "wilayah" | "stasi") || "wilayah");
    setWKode(z.kode ?? "");
    setWNama(z.nama ?? "");
    setWKoordinator(z.koordinator ?? "");
    setWTelepon(z.teleponKoordinator ?? "");
  }

  function resetLingkunganForm() {
    setEditingLingkungan(null);
    setLWilayahId(selectedWilayahId || wilayahList[0]?.id || "");
    setLKode("");
    setLNama("");
    setLPelindung("");
    setLKetua("");
    setLTelepon("");
  }

  function startEditLingkungan(w: Lingkungan) {
    setEditingLingkungan(w);
    setLWilayahId(w.wilayahId ?? "");
    setLKode(w.kode ?? "");
    setLNama(w.nama ?? "");
    setLPelindung(w.pelindung ?? "");
    setLKetua(w.ketua ?? "");
    setLTelepon(w.teleponKetua ?? "");
  }

  async function handleSubmitWilayah(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      const kode = slugifyCode(wKode || wNama);
      if (editingWilayah?.id) {
        await updateWilayah.mutateAsync({
          id: editingWilayah.id,
          body: {
            jenis: wJenis,
            kode,
            nama: wNama.trim(),
            koordinator: wKoordinator.trim(),
            teleponKoordinator: wTelepon.trim(),
          },
        });
      } else {
        const created = await createWilayah.mutateAsync({
          jenis: wJenis,
          kode,
          nama: wNama.trim(),
          koordinator: wKoordinator.trim() || undefined,
          teleponKoordinator: wTelepon.trim() || undefined,
        });
        if (!lWilayahId && created.id) setLWilayahId(created.id);
      }
      resetWilayahForm();
    } catch (err) {
      setErrorMsg(
        err instanceof Error ? err.message : "Gagal menyimpan wilayah.",
      );
    }
  }

  async function handleSubmitLingkungan(e: FormEvent) {
    e.preventDefault();
    const targetWilayahId =
      lWilayahId || selectedWilayahId || wilayahList[0]?.id;
    if (!targetWilayahId) {
      setErrorMsg(t("ref.selectZoneFirst"));
      return;
    }
    setErrorMsg(null);
    try {
      const kode = slugifyCode(lKode || lNama);
      if (editingLingkungan?.id) {
        await updateLingkungan.mutateAsync({
          id: editingLingkungan.id,
          body: {
            wilayahId: targetWilayahId,
            kode,
            nama: lNama.trim(),
            pelindung: lPelindung.trim(),
            ketua: lKetua.trim(),
            teleponKetua: lTelepon.trim(),
          },
        });
      } else {
        await createLingkungan.mutateAsync({
          wilayahId: targetWilayahId,
          kode,
          nama: lNama.trim(),
          pelindung: lPelindung.trim() || undefined,
          ketua: lKetua.trim() || undefined,
          teleponKetua: lTelepon.trim() || undefined,
        });
      }
      resetLingkunganForm();
    } catch (err) {
      setErrorMsg(
        err instanceof Error ? err.message : "Gagal menyimpan lingkungan.",
      );
    }
  }

  const wilayahNameById = new Map(
    wilayahList.map((z) => [z.id ?? "", z.nama ?? ""]),
  );

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-8 sm:px-6">
      <div>
        <Link
          to="/ref"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-muted no-underline hover:text-fg"
        >
          <ArrowLeft aria-hidden className="size-4" />
          {t("ref.backToSettings")}
        </Link>
        <h1 className="mt-2 text-2xl font-bold text-fg">
          {t("ref.territoryTitle")}
        </h1>
        <p className="text-sm text-muted">{t("ref.territorySubtitle")}</p>
      </div>

      {!hasTenant ? (
        <div className="rounded-xl border border-line bg-surface p-6 text-sm text-muted">
          {t("ref.noActiveParish")}
        </div>
      ) : !allowed ? (
        <div className="rounded-xl border border-line bg-surface p-6 text-sm text-muted">
          {t("ref.forbiddenParishAdmin")}
        </div>
      ) : (
        <>
          {errorMsg && (
            <div
              role="alert"
              className="rounded-xl border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger"
            >
              {errorMsg}
            </div>
          )}

          <div className="grid gap-6 lg:grid-cols-2">
            {/* Wilayah / Stasi */}
            <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-lg font-semibold text-fg">
                  {t("ref.zonesSection")}
                </h2>
                <select
                  aria-label={t("ref.zoneTypeLabel")}
                  value={wilayahJenisFilter}
                  onChange={(e) =>
                    setWilayahJenisFilter(
                      e.target.value as "" | "wilayah" | "stasi",
                    )
                  }
                  className="h-8 rounded-lg border border-line bg-canvas px-2 text-xs text-fg"
                >
                  <option value="">{t("ref.allTypes")}</option>
                  <option value="wilayah">{t("ref.typeZone")}</option>
                  <option value="stasi">{t("ref.typeStation")}</option>
                </select>
              </div>

              <form
                onSubmit={handleSubmitWilayah}
                className="grid gap-2.5 rounded-xl border border-line bg-canvas p-4 sm:grid-cols-2"
              >
                <select
                  value={wJenis}
                  onChange={(e) =>
                    setWJenis(e.target.value as "wilayah" | "stasi")
                  }
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                >
                  <option value="wilayah">{t("ref.typeZone")}</option>
                  <option value="stasi">{t("ref.typeStation")}</option>
                </select>
                <input
                  required
                  value={wKode}
                  onChange={(e) => setWKode(e.target.value)}
                  placeholder={t("ref.zoneCodePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  required
                  value={wNama}
                  onChange={(e) => setWNama(e.target.value)}
                  placeholder={t("ref.zoneNamePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg sm:col-span-2"
                />
                <input
                  value={wKoordinator}
                  onChange={(e) => setWKoordinator(e.target.value)}
                  placeholder={t("ref.coordinatorNamePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  value={wTelepon}
                  onChange={(e) => setWTelepon(e.target.value)}
                  placeholder={t("ref.phoneLabel")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <div className="flex justify-end gap-2 sm:col-span-2">
                  {editingWilayah && (
                    <button
                      type="button"
                      onClick={resetWilayahForm}
                      className="h-9 rounded-lg border border-line px-3 text-xs font-medium text-fg"
                    >
                      {t("users.cancel")}
                    </button>
                  )}
                  <button
                    type="submit"
                    disabled={
                      createWilayah.isPending || updateWilayah.isPending
                    }
                    className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-primary px-4 text-sm font-semibold text-on-primary disabled:opacity-50"
                  >
                    <Plus aria-hidden className="size-4" />
                    {editingWilayah ? t("ref.updateItem") : t("ref.addZone")}
                  </button>
                </div>
              </form>

              <div className="divide-y divide-line">
                {wilayahList.map((z) => (
                  <div
                    key={z.id}
                    className={`flex items-center justify-between gap-2 py-2.5 ${
                      selectedWilayahId === z.id
                        ? "bg-primary/5 px-2 rounded-lg"
                        : ""
                    }`}
                  >
                    <button
                      type="button"
                      onClick={() =>
                        setSelectedWilayahId((prev) =>
                          prev === z.id ? "" : (z.id ?? ""),
                        )
                      }
                      className="flex-1 text-left"
                    >
                      <p className="text-sm font-semibold text-fg">
                        {z.nama}{" "}
                        <span className="text-xs font-normal text-muted">
                          ({z.kode} ·{" "}
                          {z.jenis === "stasi"
                            ? t("ref.typeStation")
                            : t("ref.typeZone")}
                          )
                        </span>
                      </p>
                      {(z.koordinator || z.teleponKoordinator) && (
                        <p className="text-xs text-muted">
                          {z.koordinator}
                          {z.teleponKoordinator
                            ? ` · ${z.teleponKoordinator}`
                            : ""}
                        </p>
                      )}
                    </button>
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        title={z.active ? t("ref.active") : t("ref.inactive")}
                        onClick={async () => {
                          if (!z.id) return;
                          setErrorMsg(null);
                          try {
                            await updateWilayah.mutateAsync({
                              id: z.id,
                              body: { active: !z.active },
                            });
                          } catch (err) {
                            setErrorMsg(
                              err instanceof Error
                                ? err.message
                                : "Gagal mengubah status.",
                            );
                          }
                        }}
                        className={`inline-flex size-8 items-center justify-center rounded-lg hover:bg-fg/8 ${
                          z.active ? "text-primary" : "text-muted"
                        }`}
                      >
                        <Power aria-hidden className="size-4" />
                      </button>
                      <button
                        type="button"
                        aria-label={`${t("ref.editItem")} ${z.nama}`}
                        onClick={() => startEditWilayah(z)}
                        className="inline-flex size-8 items-center justify-center rounded-lg text-muted hover:bg-fg/8 hover:text-fg"
                      >
                        <Pencil aria-hidden className="size-4" />
                      </button>
                      <button
                        type="button"
                        aria-label={`${t("users.deleteUser")} ${z.nama}`}
                        onClick={async () => {
                          if (!z.id) return;
                          setErrorMsg(null);
                          try {
                            await deleteWilayah.mutateAsync(z.id);
                            if (selectedWilayahId === z.id)
                              setSelectedWilayahId("");
                          } catch (err) {
                            setErrorMsg(
                              err instanceof Error
                                ? err.message
                                : "Gagal menghapus wilayah.",
                            );
                          }
                        }}
                        className="inline-flex size-8 items-center justify-center rounded-lg text-danger hover:bg-danger/10"
                      >
                        <Trash2 aria-hidden className="size-4" />
                      </button>
                    </div>
                  </div>
                ))}
                {wilayahList.length === 0 && (
                  <p className="py-3 text-sm text-muted">
                    {t("ref.emptyList")}
                  </p>
                )}
              </div>
            </section>

            {/* Lingkungan */}
            <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-lg font-semibold text-fg">
                  {t("ref.wardsSection")}
                </h2>
                <select
                  aria-label={t("ref.filterByZone")}
                  value={selectedWilayahId}
                  onChange={(e) => setSelectedWilayahId(e.target.value)}
                  className="h-8 rounded-lg border border-line bg-canvas px-2 text-xs text-fg"
                >
                  <option value="">{t("ref.allZones")}</option>
                  {wilayahList.map((z) => (
                    <option key={z.id} value={z.id}>
                      {z.nama}
                    </option>
                  ))}
                </select>
              </div>

              <form
                onSubmit={handleSubmitLingkungan}
                className="grid gap-2.5 rounded-xl border border-line bg-canvas p-4 sm:grid-cols-2"
              >
                <select
                  value={
                    lWilayahId || selectedWilayahId || wilayahList[0]?.id || ""
                  }
                  onChange={(e) => setLWilayahId(e.target.value)}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                >
                  <option value="">{t("ref.selectZone")}</option>
                  {wilayahList.map((z) => (
                    <option key={z.id} value={z.id}>
                      {z.nama}
                    </option>
                  ))}
                </select>
                <input
                  required
                  value={lKode}
                  onChange={(e) => setLKode(e.target.value)}
                  placeholder={t("ref.wardCodePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  required
                  value={lNama}
                  onChange={(e) => setLNama(e.target.value)}
                  placeholder={t("ref.wardNamePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  value={lPelindung}
                  onChange={(e) => setLPelindung(e.target.value)}
                  placeholder={t("ref.patronSaintLabel")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  value={lKetua}
                  onChange={(e) => setLKetua(e.target.value)}
                  placeholder={t("ref.headNamePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  value={lTelepon}
                  onChange={(e) => setLTelepon(e.target.value)}
                  placeholder={t("ref.phoneLabel")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <div className="flex justify-end gap-2 sm:col-span-2">
                  {editingLingkungan && (
                    <button
                      type="button"
                      onClick={resetLingkunganForm}
                      className="h-9 rounded-lg border border-line px-3 text-xs font-medium text-fg"
                    >
                      {t("users.cancel")}
                    </button>
                  )}
                  <button
                    type="submit"
                    disabled={
                      createLingkungan.isPending || updateLingkungan.isPending
                    }
                    className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-primary px-4 text-sm font-semibold text-on-primary disabled:opacity-50"
                  >
                    <Plus aria-hidden className="size-4" />
                    {editingLingkungan ? t("ref.updateItem") : t("ref.addWard")}
                  </button>
                </div>
              </form>

              <div className="divide-y divide-line">
                {lingkunganList.map((w) => (
                  <div
                    key={w.id}
                    className="flex items-center justify-between gap-2 py-2.5"
                  >
                    <div>
                      <p className="text-sm font-semibold text-fg">
                        {w.nama}{" "}
                        <span className="text-xs font-normal text-muted">
                          ({w.kode}
                          {w.wilayahId && wilayahNameById.get(w.wilayahId)
                            ? ` · ${wilayahNameById.get(w.wilayahId)}`
                            : ""}
                          )
                        </span>
                      </p>
                      {(w.ketua || w.teleponKetua) && (
                        <p className="text-xs text-muted">
                          {w.ketua}
                          {w.teleponKetua ? ` · ${w.teleponKetua}` : ""}
                        </p>
                      )}
                    </div>
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        title={w.active ? t("ref.active") : t("ref.inactive")}
                        onClick={async () => {
                          if (!w.id) return;
                          setErrorMsg(null);
                          try {
                            await updateLingkungan.mutateAsync({
                              id: w.id,
                              body: { active: !w.active },
                            });
                          } catch (err) {
                            setErrorMsg(
                              err instanceof Error
                                ? err.message
                                : "Gagal mengubah status.",
                            );
                          }
                        }}
                        className={`inline-flex size-8 items-center justify-center rounded-lg hover:bg-fg/8 ${
                          w.active ? "text-primary" : "text-muted"
                        }`}
                      >
                        <Power aria-hidden className="size-4" />
                      </button>
                      <button
                        type="button"
                        aria-label={`${t("ref.editItem")} ${w.nama}`}
                        onClick={() => startEditLingkungan(w)}
                        className="inline-flex size-8 items-center justify-center rounded-lg text-muted hover:bg-fg/8 hover:text-fg"
                      >
                        <Pencil aria-hidden className="size-4" />
                      </button>
                      <button
                        type="button"
                        aria-label={`${t("users.deleteUser")} ${w.nama}`}
                        onClick={async () => {
                          if (!w.id) return;
                          setErrorMsg(null);
                          try {
                            await deleteLingkungan.mutateAsync(w.id);
                          } catch (err) {
                            setErrorMsg(
                              err instanceof Error
                                ? err.message
                                : "Gagal menghapus lingkungan.",
                            );
                          }
                        }}
                        className="inline-flex size-8 items-center justify-center rounded-lg text-danger hover:bg-danger/10"
                      >
                        <Trash2 aria-hidden className="size-4" />
                      </button>
                    </div>
                  </div>
                ))}
                {lingkunganList.length === 0 && (
                  <p className="py-3 text-sm text-muted">
                    {t("ref.emptyList")}
                  </p>
                )}
              </div>
            </section>
          </div>
        </>
      )}
    </div>
  );
}
